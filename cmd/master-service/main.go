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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "github.com/homindolenern/goapps-costing-v1/gen/go/costing/v1"
	appparam "github.com/homindolenern/goapps-costing-v1/internal/application/parameter"
	appuom "github.com/homindolenern/goapps-costing-v1/internal/application/uom"
	"github.com/homindolenern/goapps-costing-v1/internal/config"
	grpcdelivery "github.com/homindolenern/goapps-costing-v1/internal/delivery/grpc"
	"github.com/homindolenern/goapps-costing-v1/internal/delivery/grpc/interceptors"
	httpdelivery "github.com/homindolenern/goapps-costing-v1/internal/delivery/http"
	"github.com/homindolenern/goapps-costing-v1/internal/infrastructure/postgres"
	"github.com/homindolenern/goapps-costing-v1/internal/infrastructure/redis"
)

// swaggerHTML is the Swagger UI HTML template
const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Costing API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { display: none !important; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/swagger/api.swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        plugins: [SwaggerUIBundle.plugins.DownloadUrl],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`

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

	// Initialize Redis (optional - continue without it)
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Warn().Err(err).Msg("Redis connection failed - caching disabled")
		redisClient = nil
	} else {
		defer redisClient.Close()
	}

	// Initialize repositories
	uomRepo := postgres.NewUOMRepository(db)
	paramRepo := postgres.NewParameterRepository(db)

	// Initialize UOM application handlers
	uomCreateHandler := appuom.NewCreateHandler(uomRepo)
	uomUpdateHandler := appuom.NewUpdateHandler(uomRepo)
	uomDeleteHandler := appuom.NewDeleteHandler(uomRepo)
	uomGetHandler := appuom.NewGetHandler(uomRepo)
	uomListHandler := appuom.NewListHandler(uomRepo)

	// Initialize Parameter application handlers
	paramCreateHandler := appparam.NewCreateHandler(paramRepo)
	paramUpdateHandler := appparam.NewUpdateHandler(paramRepo)
	paramDeleteHandler := appparam.NewDeleteHandler(paramRepo)
	paramGetHandler := appparam.NewGetHandler(paramRepo)
	paramListHandler := appparam.NewListHandler(paramRepo)

	// Initialize gRPC handlers
	uomHandler := grpcdelivery.NewUOMHandler(
		uomCreateHandler,
		uomUpdateHandler,
		uomDeleteHandler,
		uomGetHandler,
		uomListHandler,
	)
	paramHandler := grpcdelivery.NewParameterHandler(
		paramCreateHandler,
		paramUpdateHandler,
		paramDeleteHandler,
		paramGetHandler,
		paramListHandler,
	)
	healthHandler := grpcdelivery.NewHealthHandlerWithRedis(db, redisClient)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	g, ctx := errgroup.WithContext(ctx)

	// Start gRPC server
	g.Go(func() error {
		return runGRPCServer(ctx, cfg, uomHandler, paramHandler, healthHandler)
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
	paramHandler *grpcdelivery.ParameterHandler,
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
	pb.RegisterParameterServiceServer(grpcServer, paramHandler)
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
	mux := httpdelivery.NewServeMux()

	// Connect to gRPC server
	grpcAddr := fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register gateway handlers
	if err := pb.RegisterUOMServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register UOM gateway: %w", err)
	}
	if err := pb.RegisterParameterServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register Parameter gateway: %w", err)
	}
	if err := pb.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register Health gateway: %w", err)
	}

	// Create HTTP server with additional endpoints
	httpMux := http.NewServeMux()

	// Swagger UI (must be before gRPC-Gateway catch-all)
	httpMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger/" || r.URL.Path == "/swagger" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(swaggerHTML))
		} else if r.URL.Path == "/swagger/api.swagger.json" {
			http.ServeFile(w, r, "gen/openapi/api.swagger.json")
		} else {
			http.NotFound(w, r)
		}
	})
	httpMux.Handle("/swagger", http.RedirectHandler("/swagger/", http.StatusMovedPermanently))

	// Prometheus metrics endpoint
	httpMux.Handle("/metrics", promhttp.Handler())

	// gRPC-Gateway handler (catch-all, must be last)
	httpMux.Handle("/", mux)

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
