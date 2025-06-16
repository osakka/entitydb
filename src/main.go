// Package main provides the EntityDB server implementation.
//
// EntityDB is a high-performance temporal database where every tag is timestamped
// with nanosecond precision. It features a custom binary format (EBF) with 
// Write-Ahead Logging, ACID compliance, and enterprise-grade RBAC.
//
// The server supports multiple storage backends including high-performance
// memory-mapped files and standard file-based storage with various indexing
// strategies for optimal query performance.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	
	"entitydb/models"
	"entitydb/storage/binary"
	"entitydb/api"
	"entitydb/logger"
	"entitydb/config"
	
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	
	_ "entitydb/docs" // This is required for swagger
)

// @title EntityDB API
// @version 2.32.0-dev
// @description A temporal database with pure entity-based architecture
// @termsOfService https://github.com/osakka/entitydb

// @contact.name EntityDB Support
// @contact.email support@entitydb.io

// @license.name MIT
// @license.url https://github.com/osakka/entitydb/blob/main/LICENSE

// @host localhost:8085
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication. Example: "Bearer <token>"

// =============================================================================
// Global Variables and Configuration
// =============================================================================

// Build-time version information that can be overridden during compilation.
//
// These variables are designed to be set at build time using Go's -ldflags
// to embed version and build information directly into the binary. This
// provides accurate version reporting regardless of deployment environment.
//
// Usage with go build:
//   go build -ldflags "-X main.Version=2.29.0 -X main.BuildDate=$(date +%Y-%m-%d)"
//
// Usage with Makefile:
//   VERSION := $(shell git describe --tags --always)
//   BUILD_DATE := $(shell date +%Y-%m-%d)
//   LDFLAGS := -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)
//
// These values are displayed by the --version flag and included in API responses
// to help with deployment tracking and support diagnostics.
var (
	// Version is the EntityDB version string, typically in semantic versioning format.
	// Default: "2.32.0" (current development version)
	// Build override: -ldflags "-X main.Version=x.y.z"
	// Used in: version command output, API responses, swagger documentation
	Version = "2.32.0"
	
	// BuildDate is the date when the binary was compiled.
	// Default: "unknown" (for development builds)
	// Build override: -ldflags "-X main.BuildDate=YYYY-MM-DD" 
	// Format: YYYY-MM-DD (ISO 8601 date format)
	// Used in: version command output, diagnostics, support information
	BuildDate = "unknown"
)

// Global application state and configuration
var (
	// configManager handles the three-tier configuration hierarchy system.
	// Initialized during application startup and used throughout the server
	// lifecycle to provide centralized configuration management.
	// See config.ConfigManager for detailed documentation.
	configManager *config.ConfigManager
)

// Command-line flag variables for essential functions.
//
// EntityDB follows a strict policy of using long-form flags (--entitydb-*)
// for configuration and reserving short flags only for essential functions
// like help and version display. This prevents conflicts with other tools
// and provides clear, unambiguous flag names.
//
// Flag Processing:
//   These flags are processed before full configuration initialization
//   to allow immediate exit for help/version requests without requiring
//   full server initialization.
var (
	// showVersion indicates whether to display version information and exit.
	// Triggered by: -v, --version
	// Action: Print version and build date, then exit with code 0
	showVersion bool
	
	// showHelp indicates whether to display help information and exit.
	// Triggered by: -h, --help
	// Action: Print usage information and flag descriptions, then exit with code 0
	showHelp bool
)

// =============================================================================
// Type Definitions
// =============================================================================


// EntityDBServer represents the main server instance with all its dependencies
type EntityDBServer struct {
	entityRepo       models.EntityRepository
	securityManager  *models.SecurityManager
	securityInit     *models.SecurityInitializer
	mu               sync.RWMutex
	server           *http.Server
	entityHandler    *api.EntityHandler
	userHandler      *api.UserHandler
	authHandler      *api.AuthHandler
	securityMiddleware *api.SecurityMiddleware
	config           *config.Config
}

// NewEntityDBServer creates a new server instance
func NewEntityDBServer(cfg *config.Config) *EntityDBServer {
	server := &EntityDBServer{
		config: cfg,
	}
	return server
}

func init() {
	// Register configuration manager flags
	// All configuration flags use long names only
	// Short flags are reserved for essential functions only
	
	// Flags are registered in ConfigManager.RegisterFlags()
}

func main() {
	// Initialize repositories as nil first for configuration manager
	var entityRepo models.EntityRepository
	// Relationship system removed - use pure tag-based relationships
	
	// Create configuration manager
	configManager = config.NewConfigManager(entityRepo)
	
	// Register all configuration flags
	configManager.RegisterFlags()
	
	// Parse command line flags
	flag.Parse()
	
	// Handle essential flags
	// Check for version and help flags
	if flag.Lookup("v").Value.String() == "true" || flag.Lookup("version").Value.String() == "true" {
		fmt.Printf("%s v%s (built %s)\n", config.Load().AppName, Version, BuildDate)
		os.Exit(0)
	}
	
	if flag.Lookup("h").Value.String() == "true" || flag.Lookup("help").Value.String() == "true" {
		fmt.Printf("EntityDB Server v%s\n\n", Version)
		fmt.Println("Usage: entitydb [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nAll options can also be set via environment variables.")
		fmt.Println("See documentation for complete configuration guide.")
		os.Exit(0)
	}
	
	// Initialize configuration with proper hierarchy
	cfg, err := configManager.Initialize()
	if err != nil {
		// Cannot use logger yet since configuration initialization failed
		fmt.Fprintf(os.Stderr, "Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Configure logging from configuration
	logger.Configure()
	
	
	// Set log level from configuration
	if err := logger.SetLogLevel(cfg.LogLevel); err != nil {
		logger.Fatalf("Invalid log level: %v", err)
	}
	
	// Check for trace subsystems from environment
	if traceSubsystems := os.Getenv("ENTITYDB_TRACE_SUBSYSTEMS"); traceSubsystems != "" {
		subsystems := strings.Split(traceSubsystems, ",")
		for i, s := range subsystems {
			subsystems[i] = strings.TrimSpace(s)
		}
		logger.EnableTrace(subsystems...)
		logger.Info("trace subsystems enabled: %s", strings.Join(subsystems, ", "))
	}
	
	
	logger.Info("starting entitydb with log level %s", strings.ToUpper(logger.GetLogLevel()))

	// Async metrics system - eliminates authentication deadlocks
	// Initialize early before repository creation for optimal performance
	var asyncMetricsCollector *binary.AsyncMetricsCollector
	if cfg.MetricsEnableStorageTracking {
		logger.Info("Initializing async metrics collection system")
	} else {
		logger.Info("Storage metrics tracking disabled")
	}
	
	// Initialize binary repositories
	// Use factory to create appropriate repository based on settings
	factory := &binary.RepositoryFactory{}
	
	// Set environment variable for high performance mode based on configuration
	if cfg.HighPerformance {
		logger.Info("High performance mode enabled")
		os.Setenv("ENTITYDB_HIGH_PERFORMANCE", "true")
	}
	
	// Create entity repository based on configuration
	entityRepo, err = factory.CreateRepository(cfg)
	if err != nil {
		logger.Fatalf("Failed to create entity repository: %v", err)
	}
	
	// Relationship repository removed - use pure tag-based relationships
	
	// Initialize async metrics system after repository is ready
	if cfg.MetricsEnableStorageTracking {
		asyncConfig := binary.DefaultAsyncMetricsConfig()
		asyncConfig.BufferSize = 5000
		asyncConfig.WorkerCount = 2
		asyncConfig.FlushInterval = 30 * time.Second
		
		var err error
		asyncMetricsCollector, err = binary.NewAsyncMetricsCollector(entityRepo, asyncConfig)
		if err != nil {
			logger.Warn("Failed to create async metrics collector: %v", err)
		} else {
			// Start the async metrics system
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			
			if err := asyncMetricsCollector.Start(ctx); err != nil {
				logger.Warn("Failed to start async metrics collector: %v", err)
			} else {
				// Initialize storage metrics with async collection
				binary.InitAsyncStorageMetrics(entityRepo, asyncMetricsCollector)
				logger.Info("Async metrics collection system started successfully")
			}
		}
	}
	
	// Update configuration manager with repository
	configManager = config.NewConfigManager(entityRepo)
	// Don't register flags again - already done before flag.Parse()
	
	// Refresh configuration from database now that repository is available
	if updatedCfg, err := configManager.Initialize(); err == nil {
		cfg = updatedCfg
		logger.Info("Configuration refreshed from database")
	}
	
	// Create server
	server := NewEntityDBServer(cfg)
	server.entityRepo = entityRepo
	
	// Initialize security system
	server.securityManager = models.NewSecurityManager(entityRepo)
	server.securityInit = models.NewSecurityInitializer(server.securityManager, entityRepo)
	
	// Create handlers
	server.entityHandler = api.NewEntityHandler(entityRepo)
	server.userHandler = api.NewUserHandler(entityRepo)
	server.authHandler = api.NewAuthHandler(server.securityManager)
	server.securityMiddleware = api.NewSecurityMiddleware(server.securityManager)
	
	// Initialize with default entities
	server.initializeEntities()

	// Set up HTTP server with gorilla/mux 
	// Using gorilla/mux provides better route ordering control than standard ServeMux
	// This prevents the static file handler from intercepting API routes
	router := mux.NewRouter()
	
	// Create RBAC-enabled handlers using SecurityMiddleware
	// Note: EntityHandlerRBAC is being deprecated in favor of SecurityMiddleware

	// API routes on subrouter (for better ordering)
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	
	// Swagger documentation - serve spec.json at the swagger directory
	router.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(cfg.DataPath, "../src/docs", "swagger.json"))
	}).Methods("GET")
	
	// Legacy and test endpoints (non-authenticated) - will be removed in future versions
	apiRouter.HandleFunc("/status", server.handleStatus).Methods("GET") 
	
	// Entity endpoints with RBAC (all entity operations require authentication and permissions)
	// Use SecurityMiddleware for modern tag-based RBAC
	apiRouter.HandleFunc("/entities/list", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.ListEntities)).Methods("GET")
	apiRouter.HandleFunc("/entities/get", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetEntity)).Methods("GET")
	apiRouter.HandleFunc("/entities/create", server.securityMiddleware.RequirePermission("entity", "create")(server.entityHandler.CreateEntity)).Methods("POST")
	apiRouter.HandleFunc("/entities/update", server.securityMiddleware.RequirePermission("entity", "update")(server.entityHandler.UpdateEntity)).Methods("PUT")
	apiRouter.HandleFunc("/entities/query", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.QueryEntities)).Methods("GET")
	apiRouter.HandleFunc("/entities/listbytag", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.ListEntities)).Methods("GET")
	apiRouter.HandleFunc("/entities/summary", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetEntitySummary)).Methods("GET")
	
	// Entity temporal operations with RBAC
	apiRouter.HandleFunc("/entities/as-of", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetEntityAsOf)).Methods("GET")
	apiRouter.HandleFunc("/entities/history", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetEntityHistory)).Methods("GET")
	apiRouter.HandleFunc("/entities/changes", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetRecentChanges)).Methods("GET")
	apiRouter.HandleFunc("/entities/diff", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetEntityDiff)).Methods("GET")
	
	// Chunking endpoints with RBAC  
	apiRouter.HandleFunc("/entities/get-chunk", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.GetEntity)).Methods("GET")
	apiRouter.HandleFunc("/entities/stream-content", server.securityMiddleware.RequirePermission("entity", "view")(server.entityHandler.StreamEntity)).Methods("GET")
	
	// Deprecated temporal patch endpoint
	apiRouter.HandleFunc("/patches/reindex-tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method not allowed"}`))
			return
		}
		
		// Tag fix has been integrated into the main codebase
		// No longer need to call the separate fix function
		
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"Temporal tag index has been fixed"}`))
	}).Methods("POST")
	
	// Entity relationships are now pure tag-based (e.g., "relates_to:entity_id" tags)
	// No separate relationship endpoints needed - use entity tag operations instead
	
	// Auth routes - New relationship-based security
	apiRouter.HandleFunc("/auth/login", server.authHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/auth/logout", server.authHandler.Logout).Methods("POST")
	apiRouter.HandleFunc("/auth/whoami", server.securityMiddleware.RequireAuthentication(server.authHandler.WhoAmI)).Methods("GET")
	apiRouter.HandleFunc("/auth/refresh", server.authHandler.RefreshToken).Methods("POST")
	
	
	// User management routes with RBAC
	userHandlerRBAC := api.NewUserHandlerRBAC(server.userHandler, server.entityRepo, server.securityManager)
	apiRouter.HandleFunc("/users/create", userHandlerRBAC.CreateUser()).Methods("POST")
		apiRouter.HandleFunc("/users/change-password", userHandlerRBAC.ChangePassword()).Methods("POST")
		apiRouter.HandleFunc("/users/reset-password", userHandlerRBAC.ResetPassword()).Methods("POST")
	
	// Dashboard routes with RBAC
	dashboardHandler := api.NewDashboardHandler(server.entityRepo)
	dashboardHandlerRBAC := api.NewDashboardHandlerRBAC(dashboardHandler, server.entityRepo, server.securityManager)
	apiRouter.HandleFunc("/dashboard/stats", dashboardHandlerRBAC.GetDashboardStats()).Methods("GET")
	
	// Configuration routes with RBAC
	configHandler := api.NewEntityConfigHandler(server.entityRepo)
	configHandlerRBAC := api.NewEntityConfigHandlerRBAC(configHandler, server.entityRepo, server.securityManager)
	apiRouter.HandleFunc("/config", configHandlerRBAC.GetConfig()).Methods("GET")
	apiRouter.HandleFunc("/config/set", configHandlerRBAC.SetConfig()).Methods("POST")
	apiRouter.HandleFunc("/feature-flags", configHandlerRBAC.GetFeatureFlags()).Methods("GET")
	apiRouter.HandleFunc("/feature-flags/set", configHandlerRBAC.SetFeatureFlag()).Methods("POST")
	
	// Admin routes with RBAC (require admin permission)
	adminHandler := api.NewAdminHandler(server.entityRepo)
	apiRouter.HandleFunc("/admin/reindex", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "reindex"})(adminHandler.ReindexHandler)).Methods("POST")
	apiRouter.HandleFunc("/admin/health", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "health"})(adminHandler.HealthCheckHandler)).Methods("GET")
	
	// Health endpoint (no authentication required)
	healthHandler := api.NewHealthHandler(server.entityRepo)
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	
	// Metrics endpoint (Prometheus format, no authentication required)
	metricsHandler := api.NewMetricsHandler(server.entityRepo)
	router.HandleFunc("/metrics", metricsHandler.PrometheusMetrics).Methods("GET")
	
	// Temporal metrics collection endpoints
	metricsCollector := api.NewMetricsCollector(server.entityRepo)
	apiRouter.HandleFunc("/metrics/collect", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "metrics", Action: "write"})(metricsCollector.CollectMetric)).Methods("POST")
	// apiRouter.HandleFunc("/metrics/history", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "metrics", Action: "read"})(metricsCollector.GetMetricHistory)).Methods("GET") // Disabled - using public endpoint below
	apiRouter.HandleFunc("/metrics/current", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "metrics", Action: "read"})(metricsCollector.GetCurrentMetrics)).Methods("GET")
	
	// New metrics history handler for real-time chart data (no authentication required)
	metricsHistoryHandler := api.NewMetricsHistoryHandler(server.entityRepo)
	apiRouter.HandleFunc("/metrics/history", metricsHistoryHandler.GetMetricHistory).Methods("GET")
	apiRouter.HandleFunc("/metrics/available", metricsHistoryHandler.GetAvailableMetrics).Methods("GET")
	
	// Comprehensive metrics endpoint for 70T scale monitoring
	comprehensiveMetricsHandler := api.NewComprehensiveMetricsHandler(server.entityRepo)
	apiRouter.HandleFunc("/metrics/comprehensive", comprehensiveMetricsHandler.ServeHTTP).Methods("GET")
	
	// TEMPORARY FIX: Disable background metrics collector to prevent authentication deadlock
	// Background metrics collector - designed to handle entities with many temporal tags
	// The v2.32.0 system with sharded indexing, tag caching, and memory-mapped files
	// should efficiently handle entities with 100s or 1000s of tags (this is normal for temporal data)
	logger.Info("Metrics collection interval set to %v (DISABLED)", cfg.MetricsInterval)
	
	// DISABLED: Background metrics collection temporarily disabled to fix authentication hang
	// backgroundCollector := api.NewBackgroundMetricsCollector(server.entityRepo, cfg.MetricsInterval)
	// backgroundCollector.Start()
	// defer backgroundCollector.Stop()
	
	// DISABLED: Metrics retention manager - part of metrics feedback loop
	/*
	if cfg.MetricsRetentionRaw > 0 {
		retentionManager := api.NewMetricsRetentionManager(
			server.entityRepo,
			cfg.MetricsRetentionRaw,
			cfg.MetricsRetention1Min,
			cfg.MetricsRetention1Hour,
			cfg.MetricsRetention1Day,
		)
		retentionManager.Start()
		defer retentionManager.Stop()
		logger.Info("Metrics retention manager started")
	}
	*/
	
	// Initialize query metrics collector
	api.InitQueryMetrics(server.entityRepo)
	
	// Storage metrics already initialized early, no need to reinitialize
	
	// Initialize error metrics collector
	api.InitErrorMetrics(server.entityRepo)
	
	// DISABLED: Metrics aggregator - part of metrics feedback loop
	/*
	logger.Info("Metrics aggregation interval set to %v", cfg.AggregationInterval)
	
	metricsAggregator := api.NewMetricsAggregator(server.entityRepo, cfg.AggregationInterval)
	metricsAggregator.Start()
	defer metricsAggregator.Stop()
	*/
	
	// Generic application metrics endpoint - applications can filter by namespace
	applicationMetricsHandler := api.NewApplicationMetricsHandler(server.entityRepo)
	apiRouter.HandleFunc("/application/metrics", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "metrics", Action: "read"})(applicationMetricsHandler.GetApplicationMetrics)).Methods("GET")
	
	// System metrics endpoint (EntityDB-specific, no authentication required)
	systemMetricsHandler := api.NewSystemMetricsHandler(server.entityRepo)
	apiRouter.HandleFunc("/system/metrics", systemMetricsHandler.SystemMetrics).Methods("GET")
	
	// RBAC metrics endpoints
	rbacMetricsHandler := api.NewTemporalRBACMetricsHandler(server.entityRepo, server.securityManager)
	// Admin-only detailed metrics
	apiRouter.HandleFunc("/rbac/metrics", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "view"})(rbacMetricsHandler.GetRBACMetricsFromTemporal)).Methods("GET")
	// Public basic metrics (no auth required)
	apiRouter.HandleFunc("/rbac/metrics/public", rbacMetricsHandler.GetPublicRBACMetrics).Methods("GET")
	
	// Log control endpoints (admin only)
	logControlHandler := api.NewLogControlHandler()
	apiRouter.HandleFunc("/admin/log-level", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "update"})(logControlHandler.SetLogLevel)).Methods("POST")
	apiRouter.HandleFunc("/admin/log-level", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "view"})(logControlHandler.GetLogLevel)).Methods("GET")
	apiRouter.HandleFunc("/admin/trace-subsystems", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "update"})(logControlHandler.SetTraceSubsystems)).Methods("POST")
	apiRouter.HandleFunc("/admin/trace-subsystems", api.RBACMiddleware(server.entityRepo, server.securityManager, api.RBACPermission{Resource: "admin", Action: "view"})(logControlHandler.GetTraceSubsystems)).Methods("GET")
	
	// Dataset management routes with RBAC
	datasetHandler := api.NewDatasetHandler(server.entityRepo)
	datasetHandlerRBAC := api.NewDatasetHandlerRBAC(datasetHandler, server.entityRepo, server.securityManager)
	
	// Dataset CRUD operations
	apiRouter.HandleFunc("/datasets", datasetHandlerRBAC.ListDatasets).Methods("GET")
	apiRouter.HandleFunc("/datasets", datasetHandlerRBAC.CreateDataset).Methods("POST")
	apiRouter.HandleFunc("/datasets/{id}", datasetHandlerRBAC.GetDataset).Methods("GET")
	apiRouter.HandleFunc("/datasets/{id}", datasetHandlerRBAC.UpdateDataset).Methods("PUT")
	apiRouter.HandleFunc("/datasets/{id}", datasetHandlerRBAC.DeleteDataset).Methods("DELETE")
	
	// Dataset management operations - removed grant/revoke until implemented
	
	// Dataset-scoped entity operations with RBAC
	datasetEntityHandlerRBAC := api.NewDatasetEntityHandlerRBAC(server.entityHandler, server.entityRepo, server.securityManager)
	
	// Basic dataset entity operations  
	apiRouter.HandleFunc("/datasets/{dataset}/entities/create", datasetEntityHandlerRBAC.CreateDatasetEntity()).Methods("POST")
	apiRouter.HandleFunc("/datasets/{dataset}/entities/query", datasetEntityHandlerRBAC.QueryDatasetEntities()).Methods("GET")
	
	// Dataset relationship operations - removed until implemented
	
	// Swagger UI route
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods("GET")
	
	// Static file serving with proper precedence (last)
	// This must be registered last to ensure API routes take precedence
	router.PathPrefix("/").Handler(http.HandlerFunc(server.serveStaticFile))

	// Add TE header middleware to prevent hangs with browser headers
	teHeaderMiddleware := api.NewTEHeaderMiddleware()
	
	// Add request metrics middleware (conditionally)
	var requestMetrics *api.RequestMetricsMiddleware
	// TEMPORARY FIX: Disable request metrics to prevent authentication deadlock
	if false && cfg.MetricsEnableRequestTracking {
		requestMetrics = api.NewRequestMetricsMiddleware(server.entityRepo)
		logger.Info("Request metrics tracking enabled")
	} else {
		logger.Info("Request metrics tracking disabled (FORCED)")
	}
	
	// Chain middleware together
	chainedMiddleware := func(h http.Handler) http.Handler {
		// Apply in order: TE header fix -> request metrics -> handler
		h = teHeaderMiddleware.Middleware(h)
		if requestMetrics != nil {
			h = requestMetrics.Middleware(h)
		}
		return h
	}
	
	// Add CORS middleware with very permissive settings
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Very permissive CORS settings for debugging
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Access-Control-Expose-Headers", "*")
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			h.ServeHTTP(w, r)
		})
	}
	
	// Create HTTP server with timeouts
	if cfg.UseSSL {
		// SSL enabled - create HTTPS server with HTTP/1.1 only (disable HTTP/2)
		// This fixes ERR_HTTP2_PROTOCOL_ERROR issues with some clients
		tlsConfig := &tls.Config{
			NextProtos: []string{"http/1.1"}, // Disable HTTP/2
		}
		
		server.server = &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.SSLPort),
			Handler:      corsHandler(chainedMiddleware(router)),
			TLSConfig:    tlsConfig,
			ReadTimeout:  cfg.HTTPReadTimeout,
			WriteTimeout: cfg.HTTPWriteTimeout,
			IdleTimeout:  cfg.HTTPIdleTimeout,
		}
		
		logger.Info("Starting EntityDB server on HTTPS port %d with SSL enabled", cfg.SSLPort)
		logger.Info("Server URL: https://localhost:%d", cfg.SSLPort)
		logger.Info("API documentation: https://localhost:%d/swagger/", cfg.SSLPort)
		logger.Info("Dashboard: https://localhost:%d/", cfg.SSLPort)
		
		// Start HTTPS server
		go func() {
			if err := server.server.ListenAndServeTLS(cfg.SSLCert, cfg.SSLKey); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("HTTPS server failed: %v", err)
			}
		}()
	} else {
		// SSL disabled - create HTTP server
		server.server = &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      corsHandler(chainedMiddleware(router)),
			ReadTimeout:  cfg.HTTPReadTimeout,
			WriteTimeout: cfg.HTTPWriteTimeout,
			IdleTimeout:  cfg.HTTPIdleTimeout,
		}
		
		logger.Info("Starting EntityDB server on HTTP port %d (SSL disabled)", cfg.Port)
		logger.Info("Server URL: http://localhost:%d", cfg.Port)
		logger.Info("API documentation: http://localhost:%d/swagger/", cfg.Port)
		logger.Info("Dashboard: http://localhost:%d/", cfg.Port)
		logger.Warn("SSL is disabled. For production use, enable SSL by setting ENTITYDB_USE_SSL=true")
		
		// Start HTTP server
		go func() {
			if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("HTTP server failed: %v", err)
			}
		}()
	}
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("Received signal %v, initiating graceful shutdown...", sig)
	
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	
	// Shutdown HTTP server
	if err := server.server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error: %v", err)
	}
	
	// Close repositories
	// Repository close not needed - handled by OS on process termination
	
	logger.Info("EntityDB server shutdown complete")
}

// =============================================================================
// Server Methods
// =============================================================================

// initializeEntities creates default entities if they don't exist
func (s *EntityDBServer) initializeEntities() {
	logger.Info("initializing security system")
	
	// Initialize system user configuration from Config
	models.InitializeSystemUserConfiguration(s.config.SystemUserID, s.config.SystemUsername)
	
	// Initialize bcrypt cost from Config
	models.SetBcryptCost(s.config.BcryptCost)
	
	// Initialize default security entities with configurable admin credentials
	if err := s.securityInit.InitializeDefaultSecurityEntities(
		s.config.DefaultAdminUsername,
		s.config.DefaultAdminPassword,
		s.config.DefaultAdminEmail,
	); err != nil {
		logger.Error("failed to initialize security entities: %v", err)
		return
	}
	
	logger.Debug("security system initialized")
}

// Close cleans up server resources
func (s *EntityDBServer) Close() {
	// Close repositories if they have close methods
	logger.Debug("closing repositories")
	
	// Close entity repository to save tag index
	// Try different repository types
	switch repo := s.entityRepo.(type) {
	case *binary.EntityRepository:
		// All repository variants now merged into EntityRepository
		if err := repo.Close(); err != nil {
			logger.Error("Failed to close entity repository: %v", err)
		}
	case *binary.CachedRepository:
		// CachedRepository wraps another repository
		if entityRepo, ok := repo.EntityRepository.(*binary.EntityRepository); ok {
			if err := entityRepo.Close(); err != nil {
				logger.Error("Failed to close wrapped entity repository: %v", err)
			}
		}
	default:
		logger.Warn("Unknown repository type, cannot close: %T", s.entityRepo)
	}
}

// =============================================================================
// Handler Methods
// =============================================================================

// handleStatus returns server status information
func (s *EntityDBServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"version": Version,
		"build_date": BuildDate,
	})
}

// testCreateEntity creates a test entity (for debugging/testing only)
func (s *EntityDBServer) testCreateEntity(w http.ResponseWriter, r *http.Request) {
	var req api.CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}
	
	// Determine entity type from tags (look for type: tag)
	entityType := "entity" // default type
	additionalTags := []string{}
	
	for _, tag := range req.Tags {
		if strings.HasPrefix(tag, "type:") {
			entityType = strings.TrimPrefix(tag, "type:")
		} else {
			// Add non-type tags to additional tags
			additionalTags = append(additionalTags, tag)
		}
	}

	// Determine dataset - default to "default" unless specified in tags
	dataset := "default"
	for _, tag := range req.Tags {
		if strings.HasPrefix(tag, "dataset:") {
			dataset = strings.TrimPrefix(tag, "dataset:")
			break
		}
	}
	
	// Create entity using UUID architecture with system user as creator (since this is debugging/testing endpoint)
	entity, err := models.NewEntityWithMandatoryTags(
		entityType,                    // entityType
		dataset,                       // dataset
		models.SystemUserID,           // createdBy (system user for unauthenticated endpoints)
		additionalTags,               // additional tags
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create entity: " + err.Error()})
		return
	}
	
	// Handle content if provided
	if req.Content != nil {
		switch content := req.Content.(type) {
		case string:
			entity.Content = []byte(content)
		case map[string]interface{}, []interface{}:
			jsonBytes, _ := json.Marshal(content)
			entity.Content = jsonBytes
		}
	}
	
	// Create in repository
	if err := s.entityRepo.Create(entity); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entity)
}

// Relationship system removed - use pure tag-based relationships instead
// Example: To relate entity A to entity B, add tag "relates_to:entity_B_id" to entity A


// serveStaticFile serves static files from the configured directory
func (s *EntityDBServer) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	logger.Debug("serveStaticFile called for path: %s", r.URL.Path)
	
	// Only serve static files for non-API paths
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/debug/") {
		logger.Debug("Not serving static file for API/debug path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Resolve staticDir to absolute path
	absStaticDir, _ := filepath.Abs(s.config.StaticDir)
	logger.Debug("staticDir: %s, absStaticDir: %s", s.config.StaticDir, absStaticDir)
	fullPath := filepath.Join(absStaticDir, path)
	logger.Debug("fullPath: %s", fullPath)
	
	// Security check - prevent directory traversal
	cleanPath := filepath.Clean(fullPath)
	logger.Debug("cleanPath: %s", cleanPath)
	if !strings.HasPrefix(cleanPath, absStaticDir) {
		logger.Warn("Security check failed: cleanPath doesn't start with absStaticDir")
		http.NotFound(w, r)
		return
	}

	logger.Debug("Serving static file: %s", fullPath)
	
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		logger.Debug("File not found: %s", fullPath)
		http.NotFound(w, r)
		return
	}
	
	// Set proper MIME type and cache control headers based on file extension
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		// Short cache for JS files to allow updates
		w.Header().Set("Cache-Control", "public, max-age=300, must-revalidate")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		// Short cache for CSS files to allow updates
		w.Header().Set("Cache-Control", "public, max-age=300, must-revalidate")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// No cache for HTML files to ensure fresh content
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// No cache for JSON files to ensure fresh content
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		// Longer cache for SVG files as they change less frequently
		w.Header().Set("Cache-Control", "public, max-age=3600")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
		// Longer cache for images as they change less frequently
		w.Header().Set("Cache-Control", "public, max-age=3600")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
		// Longer cache for images as they change less frequently
		w.Header().Set("Cache-Control", "public, max-age=3600")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
		// Very long cache for favicon as it rarely changes
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
		// Long cache for fonts as they rarely change
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
		// Long cache for fonts as they rarely change
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
		// Long cache for fonts as they rarely change
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
		// Long cache for fonts as they rarely change
		w.Header().Set("Cache-Control", "public, max-age=86400")
	default:
		// Default no-cache for unknown file types
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}
	
	// Add ETag for proper cache validation
	fileInfo, err := os.Stat(fullPath)
	if err == nil {
		etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
		w.Header().Set("ETag", etag)
		
		// Check if client has current version
		if match := r.Header.Get("If-None-Match"); match != "" {
			if match == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}
	
	http.ServeFile(w, r, fullPath)
}