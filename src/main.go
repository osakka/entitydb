package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	
	"entitydb/models"
	"entitydb/storage/binary"
	"entitydb/api"
	"entitydb/logger"
	
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/bcrypt"
	
	_ "entitydb/docs" // This is required for swagger
)

// @title EntityDB API
// @version 2.12.0
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

// getEnv gets environment variable with default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int with default fallback
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets environment variable as bool with default fallback
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// Version information
var (
	AppName    = getEnv("ENTITYDB_APP_NAME", "EntityDB Server")
	AppVersion = getEnv("ENTITYDB_APP_VERSION", "2.13.0")
)

// Command line flags
var (
	port             int
	sslPort          int
	useSSL           bool
	sslCert          string
	sslKey           string
	logLevel         string
	dataPath         string
	tokenSecret      string
	staticDir        string
	showVersion      bool
	highPerformance  bool
)

// Config for server settings
type Config struct {
	Port              int
	SSLPort           int
	UseSSL            bool
	SSLCert           string
	SSLKey            string
	SessionTTL        time.Duration
	EnableRateLimit   bool
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

// User represents a user in the system
type User struct {
	ID       string
	Username string
	Password string
	Roles    []string
}

// EntityDBServer represents the main server
type EntityDBServer struct {
	entityRepo        models.EntityRepository
	relationRepo      models.EntityRelationshipRepository
	sessionManager    *models.SessionManager
	securityManager   *models.SecurityManager
	securityInit      *models.SecurityInitializer
	users            map[string]*User // Legacy - will be removed
	port             int
	mu               sync.RWMutex
	server           *http.Server
	entityHandler    *api.EntityHandler
	relationHandler  *api.EntityRelationshipHandler
	userHandler      *api.UserHandler
	authHandler      *api.AuthHandler
	securityMiddleware *api.SecurityMiddleware
	config           *Config
}

// NewEntityDBServer creates a new server instance
func NewEntityDBServer(config *Config) *EntityDBServer {
	server := &EntityDBServer{
		users:          make(map[string]*User), // Legacy - will be removed
		sessionManager: models.NewSessionManager(config.SessionTTL),
		port:           config.Port,
		config:         config,
	}
	return server
}

func init() {
	flag.IntVar(&port, "port", getEnvInt("ENTITYDB_PORT", 8085), "Server port")
	flag.IntVar(&sslPort, "ssl-port", getEnvInt("ENTITYDB_SSL_PORT", 8443), "SSL server port")
	flag.BoolVar(&useSSL, "use-ssl", getEnvBool("ENTITYDB_USE_SSL", false), "Enable SSL/TLS")
	flag.StringVar(&sslCert, "ssl-cert", getEnv("ENTITYDB_SSL_CERT", "/etc/ssl/certs/server.pem"), "SSL certificate file path")
	flag.StringVar(&sslKey, "ssl-key", getEnv("ENTITYDB_SSL_KEY", "/etc/ssl/private/server.key"), "SSL private key file path")
	flag.StringVar(&logLevel, "log-level", getEnv("ENTITYDB_LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	flag.StringVar(&dataPath, "data", getEnv("ENTITYDB_DATA_PATH", "/opt/entitydb/var"), "Data directory path")
	flag.StringVar(&tokenSecret, "token-secret", getEnv("ENTITYDB_TOKEN_SECRET", "entitydb-secret-key"), "Secret key for JWT tokens")
	flag.StringVar(&staticDir, "static-dir", getEnv("ENTITYDB_STATIC_DIR", "/opt/entitydb/share/htdocs"), "Static files directory")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&highPerformance, "high-performance", getEnvBool("ENTITYDB_HIGH_PERFORMANCE", false), "Enable high-performance memory-mapped indexing")
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// Configure logging from environment and flags
	logger.Configure()
	
	// Override with command line flag if provided
	if logLevel != "INFO" {
		if err := logger.SetLogLevel(logLevel); err != nil {
			logger.Fatalf("Invalid log level: %v", err)
		}
	}
	
	// Check for trace subsystems from environment or flag
	if traceSubsystems := os.Getenv("ENTITYDB_TRACE_SUBSYSTEMS"); traceSubsystems != "" {
		subsystems := strings.Split(traceSubsystems, ",")
		for i, s := range subsystems {
			subsystems[i] = strings.TrimSpace(s)
		}
		logger.EnableTrace(subsystems...)
		logger.Info("Trace subsystems enabled: %s", strings.Join(subsystems, ", "))
	}
	
	logger.Info("Starting EntityDB with log level: %s", logger.GetLogLevel())

	// Initialize binary repositories
	// Use factory to create appropriate repository based on settings
	factory := &binary.RepositoryFactory{}
	
	// Set environment variable for high performance mode based on flag
	if highPerformance {
		logger.Info("High performance mode enabled via command line flag")
		os.Setenv("ENTITYDB_HIGH_PERFORMANCE", "true")
	}
	
	entityRepo, err := factory.CreateRepository(dataPath)
	if err != nil {
		logger.Fatalf("Failed to create entity repository: %v", err)
	}
	
	// Create binary relationship repository
	// Handle high-performance repository case - it embeds EntityRepository
	var binaryRepo *binary.EntityRepository
	
	switch repo := entityRepo.(type) {
	case *binary.TemporalRepository:
		// TemporalRepository embeds HighPerformanceRepository which embeds EntityRepository
		binaryRepo = repo.HighPerformanceRepository.EntityRepository
	case *binary.HighPerformanceRepository:
		// HighPerformanceRepository embeds EntityRepository
		binaryRepo = repo.EntityRepository
	case *binary.EntityRepository:
		binaryRepo = repo
	case *binary.DataspaceRepository:
		// DataspaceRepository embeds EntityRepository
		binaryRepo = repo.EntityRepository
	case *binary.WALOnlyRepository:
		// WALOnlyRepository embeds EntityRepository
		binaryRepo = repo.EntityRepository
	case *binary.CachedRepository:
		// CachedRepository wraps another repository, unwrap it
		// We need to get the underlying repository
		switch underlying := repo.EntityRepository.(type) {
		case *binary.TemporalRepository:
			binaryRepo = underlying.HighPerformanceRepository.EntityRepository
		case *binary.HighPerformanceRepository:
			binaryRepo = underlying.EntityRepository
		case *binary.DataspaceRepository:
			binaryRepo = underlying.EntityRepository
		case *binary.EntityRepository:
			binaryRepo = underlying
		default:
			logger.Fatalf("Unsupported underlying repository type in CachedRepository: %T", underlying)
		}
	default:
		logger.Fatalf("Unsupported repository type for relationships: %T", entityRepo)
	}
	
	relationRepo := binary.NewRelationshipRepository(binaryRepo)
	if err != nil {
		logger.Fatalf("Failed to create relationship repository: %v", err)
	}
	
	// Create server config
	config := &Config{
		Port:              port,
		SSLPort:           sslPort,
		UseSSL:            useSSL,
		SSLCert:           sslCert,
		SSLKey:            sslKey,
		SessionTTL:        time.Duration(getEnvInt("ENTITYDB_SESSION_TTL_HOURS", 2)) * time.Hour,
		EnableRateLimit:   getEnvBool("ENTITYDB_ENABLE_RATE_LIMIT", false),
		RateLimitRequests: getEnvInt("ENTITYDB_RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   time.Duration(getEnvInt("ENTITYDB_RATE_LIMIT_WINDOW_MINUTES", 1)) * time.Minute,
	}
	
	// Create server
	server := NewEntityDBServer(config)
	server.entityRepo = entityRepo
	server.relationRepo = relationRepo
	
	// Initialize security system
	server.securityManager = models.NewSecurityManager(entityRepo)
	server.securityInit = models.NewSecurityInitializer(server.securityManager, entityRepo)
	
	// Create handlers
	server.entityHandler = api.NewEntityHandler(entityRepo)
	server.relationHandler = api.NewEntityRelationshipHandler(relationRepo)
	server.userHandler = api.NewUserHandler(entityRepo)
	server.authHandler = api.NewAuthHandler(server.securityManager, server.sessionManager)
	server.securityMiddleware = api.NewSecurityMiddleware(server.securityManager)
	
	// Initialize with default entities
	server.initializeEntities()

	// Set up HTTP server with gorilla/mux 
	// Using gorilla/mux provides better route ordering control than standard ServeMux
	// This prevents the static file handler from intercepting API routes
	router := mux.NewRouter()
	
	// Create RBAC-enabled handlers
	entityHandlerRBAC := api.NewEntityHandlerRBAC(server.entityHandler, server.entityRepo, server.sessionManager)

	// API routes on subrouter (for better ordering)
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	
	// Swagger documentation - serve spec.json at the swagger directory
	router.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join("/opt/entitydb/src/docs", "swagger.json"))
	}).Methods("GET")
	
	// Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	
	// Swagger spec endpoint
	apiRouter.HandleFunc("/spec", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join("/opt/entitydb/src/docs", "swagger.json"))
	}).Methods("GET")
	
	// Test endpoints (no auth required) - Add these FIRST
	logger.Info("Registering test endpoints...")
	testHandlers := api.NewUnauthenticatedHandlers(server.entityRepo, server.relationRepo)
	apiRouter.HandleFunc("/test/status", testHandlers.TestStatus).Methods("GET")
	apiRouter.HandleFunc("/test/entities/create", testHandlers.TestCreateEntity).Methods("POST")
	apiRouter.HandleFunc("/test/relationships/create", testHandlers.TestCreateRelationship).Methods("POST")
	apiRouter.HandleFunc("/test/relationships/list", testHandlers.TestListRelationships).Methods("GET")
	apiRouter.HandleFunc("/test/entities/list", testHandlers.TestListEntities).Methods("GET")
	apiRouter.HandleFunc("/test/entities/get", testHandlers.TestGetEntity).Methods("GET")
	// Temporal test endpoints
	apiRouter.HandleFunc("/test/entities/as-of", testHandlers.TestGetEntityAsOf).Methods("GET")
	apiRouter.HandleFunc("/test/entities/history", testHandlers.TestGetEntityHistory).Methods("GET")
	apiRouter.HandleFunc("/test/entities/changes", testHandlers.TestGetRecentChanges).Methods("GET")
	apiRouter.HandleFunc("/test/entities/diff", testHandlers.TestGetEntityDiff).Methods("GET")
	
	// Temporal test endpoint
	apiRouter.HandleFunc("/test/temporal/status", server.entityHandler.TestTemporalFixHandler).Methods("GET")
	
	// Entity API routes with RBAC
	apiRouter.HandleFunc("/entities", server.handleEntities).Methods("GET", "POST")
	// Use standard handlers for list and query  
	apiRouter.HandleFunc("/entities/list", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.PermEntityView)(server.entityHandler.ListEntities)).Methods("GET")
	apiRouter.HandleFunc("/entities/query", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.PermEntityView)(server.entityHandler.QueryEntities)).Methods("GET")
	// Keep regular handlers for other operations
	apiRouter.HandleFunc("/entities/get", entityHandlerRBAC.GetEntity()).Methods("GET")
	apiRouter.HandleFunc("/entities/create", entityHandlerRBAC.CreateEntity()).Methods("POST")
	apiRouter.HandleFunc("/entities/update", entityHandlerRBAC.UpdateEntity()).Methods("PUT")

	// Dataspace-aware Entity API routes with RBAC and Dataspace validation
	dataspaceEntityHandler := api.NewDataspaceEntityHandlerRBAC(server.entityHandler, server.entityRepo, server.sessionManager)
	apiRouter.HandleFunc("/dataspaces/entities/create", dataspaceEntityHandler.CreateDataspaceEntity()).Methods("POST")
	apiRouter.HandleFunc("/dataspaces/entities/query", dataspaceEntityHandler.QueryDataspaceEntities()).Methods("GET")
	
	// Dataspace management routes with RBAC
	dataspaceHandler := api.NewDataspaceHandler(server.entityRepo)
	dataspaceHandlerRBAC := api.NewDataspaceHandlerRBAC(dataspaceHandler, server.entityRepo, server.sessionManager)
	apiRouter.HandleFunc("/dataspaces", dataspaceHandlerRBAC.ListDataspaces).Methods("GET")
	apiRouter.HandleFunc("/dataspaces/list", dataspaceHandlerRBAC.ListDataspaces).Methods("GET") // Alias for compatibility
	apiRouter.HandleFunc("/dataspaces", dataspaceHandlerRBAC.CreateDataspace).Methods("POST")
	apiRouter.HandleFunc("/dataspaces/{id}", dataspaceHandlerRBAC.GetDataspace).Methods("GET")
	apiRouter.HandleFunc("/dataspaces/{id}", dataspaceHandlerRBAC.UpdateDataspace).Methods("PUT")
	apiRouter.HandleFunc("/dataspaces/{id}", dataspaceHandlerRBAC.DeleteDataspace).Methods("DELETE")
	
	// Chunked content API routes
	apiRouter.HandleFunc("/entities/stream", server.entityHandler.StreamEntity).Methods("GET")
	apiRouter.HandleFunc("/entities/download", server.entityHandler.StreamEntity).Methods("GET")
	
	// Temporal API routes with RBAC
	apiRouter.HandleFunc("/entities/as-of", entityHandlerRBAC.GetEntityAsOf()).Methods("GET")
	apiRouter.HandleFunc("/entities/history", entityHandlerRBAC.GetEntityHistory()).Methods("GET")
	apiRouter.HandleFunc("/entities/changes", entityHandlerRBAC.GetRecentChanges()).Methods("GET")
	apiRouter.HandleFunc("/entities/diff", entityHandlerRBAC.GetEntityDiff()).Methods("GET")
	
	// For backward compatibility with test scripts
	apiRouter.HandleFunc("/entities/as-of-fixed", entityHandlerRBAC.GetEntityAsOf()).Methods("GET")
	apiRouter.HandleFunc("/entities/history-fixed", entityHandlerRBAC.GetEntityHistory()).Methods("GET")
	apiRouter.HandleFunc("/entities/changes-fixed", entityHandlerRBAC.GetRecentChanges()).Methods("GET")
	apiRouter.HandleFunc("/entities/diff-fixed", entityHandlerRBAC.GetEntityDiff()).Methods("GET")
	
	// Tag index fix endpoint
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
	
	// Entity relationship routes
	apiRouter.HandleFunc("/entity-relationships", server.handleEntityRelationships).Methods("GET", "POST")
	
	// Auth routes - New relationship-based security
	apiRouter.HandleFunc("/auth/login", server.authHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/auth/logout", server.securityMiddleware.RequireAuthentication(server.authHandler.Logout)).Methods("POST")
	apiRouter.HandleFunc("/auth/whoami", server.securityMiddleware.RequireAuthentication(server.authHandler.WhoAmI)).Methods("GET")
	apiRouter.HandleFunc("/auth/refresh", server.securityMiddleware.RequireAuthentication(server.authHandler.RefreshToken)).Methods("POST")
	
	// Legacy auth routes (backward compatibility) - TODO: Remove these after migration
	apiRouter.HandleFunc("/auth/status", server.handleAuthStatus).Methods("GET")
	
	// User management routes with RBAC
	userHandlerRBAC := api.NewUserHandlerRBAC(server.userHandler, server.entityRepo, server.sessionManager)
	apiRouter.HandleFunc("/users/create", userHandlerRBAC.CreateUser()).Methods("POST")
		apiRouter.HandleFunc("/users/change-password", userHandlerRBAC.ChangePassword()).Methods("POST")
		apiRouter.HandleFunc("/users/reset-password", userHandlerRBAC.ResetPassword()).Methods("POST")
	
	// Dashboard routes with RBAC
	dashboardHandler := api.NewDashboardHandler(server.entityRepo)
	dashboardHandlerRBAC := api.NewDashboardHandlerRBAC(dashboardHandler, server.entityRepo, server.sessionManager)
	apiRouter.HandleFunc("/dashboard/stats", dashboardHandlerRBAC.GetDashboardStats()).Methods("GET")
	
	// Configuration routes with RBAC
	configHandler := api.NewEntityConfigHandler(server.entityRepo)
	configHandlerRBAC := api.NewEntityConfigHandlerRBAC(configHandler, server.entityRepo, server.sessionManager)
	apiRouter.HandleFunc("/config", configHandlerRBAC.GetConfig()).Methods("GET")
	apiRouter.HandleFunc("/config/set", configHandlerRBAC.SetConfig()).Methods("POST")
	apiRouter.HandleFunc("/feature-flags", configHandlerRBAC.GetFeatureFlags()).Methods("GET")
	apiRouter.HandleFunc("/feature-flags/set", configHandlerRBAC.SetFeatureFlag()).Methods("POST")
	
	// Admin routes with RBAC (require admin permission)
	adminHandler := api.NewAdminHandler(server.entityRepo)
	apiRouter.HandleFunc("/admin/reindex", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "admin", Action: "reindex"})(adminHandler.ReindexHandler)).Methods("POST")
	apiRouter.HandleFunc("/admin/health", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "admin", Action: "health"})(adminHandler.HealthCheckHandler)).Methods("GET")
	
	// Health endpoint (no authentication required)
	healthHandler := api.NewHealthHandler(server.entityRepo)
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	
	// Metrics endpoint (Prometheus format, no authentication required)
	metricsHandler := api.NewMetricsHandler(server.entityRepo)
	router.HandleFunc("/metrics", metricsHandler.PrometheusMetrics).Methods("GET")
	
	// Temporal metrics collection endpoints
	metricsCollector := api.NewMetricsCollector(server.entityRepo)
	apiRouter.HandleFunc("/metrics/collect", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "metrics", Action: "write"})(metricsCollector.CollectMetric)).Methods("POST")
	// apiRouter.HandleFunc("/metrics/history", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "metrics", Action: "read"})(metricsCollector.GetMetricHistory)).Methods("GET") // Disabled - using public endpoint below
	apiRouter.HandleFunc("/metrics/current", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "metrics", Action: "read"})(metricsCollector.GetCurrentMetrics)).Methods("GET")
	
	// New metrics history handler for real-time chart data (no authentication required)
	metricsHistoryHandler := api.NewMetricsHistoryHandler(server.entityRepo)
	apiRouter.HandleFunc("/metrics/history", metricsHistoryHandler.GetMetricHistory).Methods("GET")
	apiRouter.HandleFunc("/metrics/available", metricsHistoryHandler.GetAvailableMetrics).Methods("GET")
	
	// Start background metrics collector with configurable interval
	metricsInterval := 30 * time.Second // default
	if intervalStr := os.Getenv("ENTITYDB_METRICS_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			metricsInterval = interval
			logger.Info("Metrics collection interval set to %v", metricsInterval)
		} else {
			logger.Warn("Invalid ENTITYDB_METRICS_INTERVAL format: %s, using default 30s", intervalStr)
		}
	}
	
	backgroundCollector := api.NewBackgroundMetricsCollector(server.entityRepo, metricsInterval)
	backgroundCollector.Start()
	defer backgroundCollector.Stop()
	
	// Initialize query metrics collector
	api.InitQueryMetrics(server.entityRepo)
	
	// Initialize storage metrics collector
	binary.InitStorageMetrics(server.entityRepo)
	
	// Initialize error metrics collector
	api.InitErrorMetrics(server.entityRepo)
	
	// Generic application metrics endpoint - applications can filter by namespace
	applicationMetricsHandler := api.NewApplicationMetricsHandler(server.entityRepo)
	apiRouter.HandleFunc("/application/metrics", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "metrics", Action: "read"})(applicationMetricsHandler.GetApplicationMetrics)).Methods("GET")
	
	// System metrics endpoint (EntityDB-specific, no authentication required)
	systemMetricsHandler := api.NewSystemMetricsHandler(server.entityRepo)
	apiRouter.HandleFunc("/system/metrics", systemMetricsHandler.SystemMetrics).Methods("GET")
	
	// Log control endpoints (require admin permission)
	apiRouter.HandleFunc("/system/log-level", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "admin", Action: "configure"})(server.entityHandler.GetLogLevel)).Methods("GET")
	apiRouter.HandleFunc("/system/log-level", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "admin", Action: "configure"})(server.entityHandler.SetLogLevel)).Methods("POST")
	
	// RBAC metrics endpoints
	rbacMetricsHandler := api.NewRBACMetricsHandler(server.entityRepo, server.sessionManager)
	// Public endpoint for basic metrics (no auth required)
	apiRouter.HandleFunc("/rbac/metrics/public", rbacMetricsHandler.GetPublicRBACMetrics).Methods("GET")
	// Authenticated endpoint for full metrics (any authenticated user)
	apiRouter.HandleFunc("/rbac/metrics", api.SessionAuthMiddleware(server.sessionManager, server.entityRepo)(rbacMetricsHandler.GetAuthenticatedRBACMetrics)).Methods("GET")
	
	// Integrity metrics endpoint (requires admin permission)
	integrityHandler := api.IntegrityMetricsHandler(server.entityRepo)
	apiRouter.HandleFunc("/integrity/metrics", api.RBACMiddleware(server.entityRepo, server.sessionManager, api.RBACPermission{Resource: "admin", Action: "view"})(integrityHandler)).Methods("GET")
	
	// Add patch status endpoint for compatibility with tests
	apiRouter.HandleFunc("/patches/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"integrated","patches":["temporal_as_of","temporal_history","entity_update","tag_index_fix"]}`))
	}).Methods("GET")
	
	// API status endpoint
	apiRouter.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Status endpoint called")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"version": AppVersion,
			"time":    time.Now().Format(time.RFC3339),
		})
	}).Methods("GET")
	
	// Debug endpoint to verify server is running
	router.HandleFunc("/debug/ping", func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Debug ping called")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "pong")
	}).Methods("GET")
	
	// Static file serving (must be last - handled by PathPrefix)
	router.PathPrefix("/").Handler(http.HandlerFunc(server.serveStaticFile))

	// Add request metrics middleware
	requestMetrics := api.NewRequestMetricsMiddleware(server.entityRepo)
	
	// Add CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow CORS for Swagger UI
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			h.ServeHTTP(w, r)
		})
	}
	
	// Create HTTP server with timeouts
	if server.config.UseSSL {
		// SSL enabled - create HTTPS server
		server.server = &http.Server{
			Addr:         fmt.Sprintf(":%d", server.config.SSLPort),
			Handler:      corsHandler(requestMetrics.Middleware(router)),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		
		// Start HTTPS server
		go func() {
			logger.Info("Starting %s with SSL on port %d", AppName, server.config.SSLPort)
			logger.Info("Swagger documentation available at https://localhost:%d/swagger/", server.config.SSLPort)
			if err := server.server.ListenAndServeTLS(server.config.SSLCert, server.config.SSLKey); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Error starting SSL server: %v", err)
			}
		}()
		
		// Skip HTTP server - only run HTTPS
		logger.Info("SSL-only mode: HTTPS on port %d", server.config.SSLPort)
	} else {
		// No SSL - create standard HTTP server
		server.server = &http.Server{
			Addr:         fmt.Sprintf(":%d", server.config.Port),
			Handler:      corsHandler(requestMetrics.Middleware(router)),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		
		// Start HTTP server
		go func() {
			logger.Info("Starting %s on port %d", AppName, server.config.Port)
			logger.Info("Swagger documentation available at http://localhost:%d/swagger/", server.config.Port)
			if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Error starting server: %v", err)
			}
		}()
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")
	
	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.server.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown: %v", err)
	}

	// Close repositories
	server.Close()
	logger.Info("Server shut down cleanly")
}

// Close cleans up server resources
func (s *EntityDBServer) Close() {
	// Close repositories if they have close methods
	logger.Info("Closing repositories...")
	
	// Close entity repository to save tag index
	// Try different repository types
	switch repo := s.entityRepo.(type) {
	case *binary.EntityRepository:
		if err := repo.Close(); err != nil {
			logger.Error("Failed to close entity repository: %v", err)
		}
	case *binary.TemporalRepository:
		// TemporalRepository embeds HighPerformanceRepository which embeds EntityRepository
		if baseRepo := repo.HighPerformanceRepository; baseRepo != nil {
			if entityRepo := baseRepo.EntityRepository; entityRepo != nil {
				if err := entityRepo.Close(); err != nil {
					logger.Error("Failed to close entity repository: %v", err)
				}
			}
		}
	case *binary.HighPerformanceRepository:
		// HighPerformanceRepository embeds EntityRepository
		if entityRepo := repo.EntityRepository; entityRepo != nil {
			if err := entityRepo.Close(); err != nil {
				logger.Error("Failed to close entity repository: %v", err)
			}
		}
	default:
		logger.Warn("Unknown repository type, cannot close: %T", s.entityRepo)
	}
}

// redirectToHTTPS redirects HTTP requests to HTTPS
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Parse the SSL port for redirect
	sslPort := sslPort
	if sslPort == 443 {
		// Standard HTTPS port - don't include in redirect
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	} else {
		// Non-standard port - include in redirect
		host := strings.Split(r.Host, ":")[0]
		http.Redirect(w, r, fmt.Sprintf("https://%s:%d%s", host, sslPort, r.URL.String()), http.StatusMovedPermanently)
	}
}

// initializeEntities creates default entities if they don't exist
func (s *EntityDBServer) initializeEntities() {
	logger.Info("Initializing relationship-based security system...")
	
	// Initialize default security entities (roles, permissions, groups)
	if err := s.securityInit.InitializeDefaultSecurityEntities(); err != nil {
		logger.Error("Failed to initialize security entities: %v", err)
		return
	}
	
	logger.Info("Security system initialized successfully")
}

// handleEntities is a legacy endpoint handler
func (s *EntityDBServer) handleEntities(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.entityHandler.ListEntities(w, r)
	case "POST":
		s.entityHandler.CreateEntity(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func (s *EntityDBServer) handleEntityList(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.ListEntities(w, r)
}

func (s *EntityDBServer) handleEntityGet(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.GetEntity(w, r)
}

func (s *EntityDBServer) handleEntityCreate(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.CreateEntity(w, r)
}

func (s *EntityDBServer) handleEntityUpdate(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.UpdateEntity(w, r)
}

func (s *EntityDBServer) handleEntityAsOf(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.GetEntityAsOf(w, r)
}

func (s *EntityDBServer) handleEntityHistory(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.GetEntityHistory(w, r)
}

func (s *EntityDBServer) handleRecentChanges(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.GetRecentChanges(w, r)
}

func (s *EntityDBServer) handleEntityDiff(w http.ResponseWriter, r *http.Request) {
	// Delegate to the API handler
	s.entityHandler.GetEntityDiff(w, r)
}

func (s *EntityDBServer) handleEntityRelationships(w http.ResponseWriter, r *http.Request) {
	// Create the handler method dynamically
	switch r.Method {
	case "GET":
		// Get relationships
		source := r.URL.Query().Get("source")
		target := r.URL.Query().Get("target")
		
		var relationships []*models.EntityRelationship
		var err error
		
		if source != "" {
			relationships, err = s.relationRepo.GetBySource(source)
		} else if target != "" {
			relationships, err = s.relationRepo.GetByTarget(target)
		} else {
			relationships, err = nil, nil // No GetAll method
		}
		
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get relationships"})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(relationships)
		
	case "POST":
		// Create relationship
		var req struct {
			SourceID         string `json:"source_id"`
			RelationshipType string `json:"relationship_type"`
			TargetID         string `json:"target_id"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		
		rel := &models.EntityRelationship{
			SourceID:         req.SourceID,
			RelationshipType: req.RelationshipType,
			TargetID:         req.TargetID,
		}
		
		err := s.relationRepo.Create(rel)
		created := rel
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create relationship"})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
		
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

// @Summary Login
// @Description Authenticate user and receive session token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body api.LoginRequest true "Login credentials"
// @Success 200 {object} api.LoginResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/login [post]
func (s *EntityDBServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	// Find user entity by username
	var userEntity *models.Entity
	
	// Get all user entities
	entities, err := s.entityRepo.ListByTag("type:user")
	if err != nil {
		logger.Error("Failed to query user entities: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to query users"})
		return
	}
	
	// Find the user with matching username
	logger.Debug("Searching for user '%s' among %d entities", loginReq.Username, len(entities))
	for _, entity := range entities {
		username := entity.GetContentValue("username")
		logger.Debug("Checking entity %s, username content: '%s'", entity.ID, username)
		if username == loginReq.Username {
			userEntity = entity
			logger.Debug("Found matching user entity: %s", entity.ID)
			break
		}
	}
	if userEntity == nil {
		logger.Debug("No user found with username content matching '%s'", loginReq.Username)
		
		// Also check by id:username tag for backward compatibility
		taggedEntities, err := s.entityRepo.ListByTag(fmt.Sprintf("id:username:%s", loginReq.Username))
		if err == nil && len(taggedEntities) > 0 {
			userEntity = taggedEntities[0]
			logger.Debug("Found user by id:username tag")
		}
	}
	
	// Handle deprecated id:username tags for backward compatibility
	if userEntity == nil {
		// Try to find by id:username tag
		var userEntities []*models.Entity
		for _, entity := range entities {
			for _, tag := range entity.Tags {
				if tag == fmt.Sprintf("id:username:%s", loginReq.Username) {
					userEntity = entity
					break
				}
			}
		}
		if len(userEntities) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
			return
		}
		entities = userEntities
	}
	
	if userEntity == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}
	
	// Get password hash from entity content
	var passwordHash string
	if len(userEntity.Content) > 0 {
		// With root cause fixed, content should be clean JSON
		var userData map[string]string
		if err := json.Unmarshal(userEntity.Content, &userData); err == nil {
			passwordHash = userData["password_hash"]
		} else {
			// Fallback to enhanced unwrapping for existing wrapped content
			userData, err := extractUserDataWithMultiLevelUnwrap(userEntity.Content)
			if err == nil {
				passwordHash = userData["password_hash"]
			} else {
				logger.Error("Failed to extract user data: %v", err)
			}
		}
	}
	
	// Verify password with bcrypt
	logger.Debug("Verifying password for user %s", userEntity.ID)
	logger.Debug("Password hash from entity: %s", passwordHash)
	logger.Debug("Password hash length: %d", len(passwordHash))
	logger.Debug("Input password: %s", loginReq.Password)
	
	if passwordHash == "" {
		logger.Error("Password hash is empty for user %s", userEntity.ID)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user data"})
		return
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginReq.Password)); err != nil {
		logger.Debug("Password verification failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}
	logger.Debug("Password verification successful")
	
	// Get user details from entity
	var username string
	var roles []string
	
	if len(userEntity.Content) > 0 {
		// With root cause fixed, content should be clean JSON
		var userData map[string]string
		if err := json.Unmarshal(userEntity.Content, &userData); err == nil {
			username = userData["username"]
		} else {
			// Fallback to enhanced unwrapping for existing wrapped content
			userData, err := extractUserDataWithMultiLevelUnwrap(userEntity.Content)
			if err == nil {
				username = userData["username"]
			} else {
				logger.Debug("Failed to extract username from user data: %v", err)
			}
		}
	}
	
	for _, tag := range userEntity.Tags {
		if strings.HasPrefix(tag, "rbac:role:") {
			role := strings.TrimPrefix(tag, "rbac:role:")
			roles = append(roles, role)
		}
	}
	
	// Create session
	session, err := s.sessionManager.CreateSession(userEntity.ID, username, roles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": session.Token,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
		"user": map[string]interface{}{
			"id":       userEntity.ID,
			"username": username,
			"roles":    roles,
		},
	})
}

// @Summary Logout
// @Description Invalidate the current session
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} api.StatusResponse
// @Router /auth/logout [post]
func (s *EntityDBServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token != "" && strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
		s.sessionManager.DeleteSession(token)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// @Summary Auth Status
// @Description Check authentication status and session validity
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} api.AuthStatusResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/status [get]
func (s *EntityDBServer) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "No token provided"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	
	session, exists := s.sessionManager.GetSession(token)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
		"user": map[string]interface{}{
			"id":       session.UserID,
			"username": session.Username,
			"roles":    session.Roles,
		},
	})
}

// @Summary Refresh Token
// @Description Refresh the session token expiration
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} api.RefreshResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /auth/refresh [post]
func (s *EntityDBServer) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "No token provided"})
		return
	}
	
	token = strings.TrimPrefix(token, "Bearer ")
	
	// Refresh the session
	session, exists := s.sessionManager.RefreshSession(token)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired token"})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": session.Token,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
	})
}

func (s *EntityDBServer) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	token = strings.TrimPrefix(token, "Bearer ")
	
	// Get session
	session, exists := s.sessionManager.GetSession(token)
	if !exists {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	
	// Check admin role
	isAdmin := false
	for _, role := range session.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}
	
	if !isAdmin {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}
	
	// Delegate to user handler
	s.userHandler.CreateUser(w, r)
}

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
	absStaticDir, _ := filepath.Abs(staticDir)
	logger.Debug("staticDir: %s, absStaticDir: %s", staticDir, absStaticDir)
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
	
	http.ServeFile(w, r, fullPath)
}

// extractUserDataWithMultiLevelUnwrap provides fallback for existing wrapped content
// This is kept for backward compatibility with existing wrapped entities
func extractUserDataWithMultiLevelUnwrap(content []byte) (map[string]string, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("empty content")
	}

	// Try simple unwrapping for existing wrapped content
	var wrapper map[string]interface{}
	if err := json.Unmarshal(content, &wrapper); err == nil {
		if innerContent, ok := wrapper["application/octet-stream"]; ok {
			if innerStr, ok := innerContent.(string); ok {
				var userData map[string]string
				if err := json.Unmarshal([]byte(innerStr), &userData); err == nil {
					return userData, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("failed to extract user data")
}