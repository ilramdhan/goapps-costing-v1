package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	appuom "github.com/homindolenern/goapps-costing-v1/internal/application/uom"
	"github.com/homindolenern/goapps-costing-v1/internal/config"
	grpcdelivery "github.com/homindolenern/goapps-costing-v1/internal/delivery/grpc"
	"github.com/homindolenern/goapps-costing-v1/internal/delivery/grpc/interceptors"
	"github.com/homindolenern/goapps-costing-v1/internal/infrastructure/postgres"
)

func main() {
	// Setup zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Int("grpc_port", cfg.Server.GRPCPort).
		Int("http_port", cfg.Server.HTTPPort).
		Msg("Starting Master Service")

	// Run the server
	if err := run(cfg); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}
}

func run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	uomRepo := postgres.NewUOMRepository(db)

	// Initialize application handlers
	createHandler := appuom.NewCreateHandler(uomRepo)
	updateHandler := appuom.NewUpdateHandler(uomRepo)
	deleteHandler := appuom.NewDeleteHandler(uomRepo)
	getHandler := appuom.NewGetHandler(uomRepo)
	listHandler := appuom.NewListHandler(uomRepo)

	// Initialize gRPC handlers
	uomHandler := grpcdelivery.NewUOMHandler(
		createHandler,
		updateHandler,
		deleteHandler,
		getHandler,
		listHandler,
	)
	healthHandler := grpcdelivery.NewHealthHandler(db)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	g, ctx := errgroup.WithContext(ctx)

	// Start gRPC server
	g.Go(func() error {
		return runGRPCServer(ctx, cfg, uomHandler, healthHandler)
	})

	// Start HTTP gateway server
	g.Go(func() error {
		return runHTTPServer(ctx, cfg)
	})

	// Wait for shutdown signal
	g.Go(func() error {
		select {
		case sig := <-sigChan:
			log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
			cancel()
		case <-ctx.Done():
		}
		return nil
	})

	return g.Wait()
}

func runGRPCServer(
	ctx context.Context,
	cfg *config.Config,
	uomHandler *grpcdelivery.UOMHandler,
	healthHandler *grpcdelivery.HealthHandler,
) error {
	addr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Create protovalidate validator
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.Recovery(),
			interceptors.Logging(),
			interceptors.Validation(validator),
		),
	)

	// Register reflection for debugging
	reflection.Register(grpcServer)

	// Register service implementations
	pb.RegisterUOMServiceServer(grpcServer, uomHandler)
	pb.RegisterHealthServiceServer(grpcServer, healthHandler)

	log.Info().Str("addr", addr).Msg("gRPC server starting")

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	return nil
}

func runHTTPServer(ctx context.Context, cfg *config.Config) error {
	mux := runtime.NewServeMux()

	// Connect to gRPC server
	grpcAddr := fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register gateway handlers
	if err := pb.RegisterUOMServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register UOM gateway: %w", err)
	}
	if err := pb.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register Health gateway: %w", err)
	}

	// Create HTTP server with additional endpoints
	httpMux := http.NewServeMux()

	// gRPC-Gateway handler
	httpMux.Handle("/", mux)

	// Prometheus metrics endpoint
	httpMux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	server := &http.Server{
		Addr:    addr,
		Handler: httpMux,
	}

	log.Info().Str("addr", addr).Msg("HTTP gateway server starting")

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}
