package serve

import (
	"context"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chainlaunch/chainlaunch/config"
	_ "github.com/chainlaunch/chainlaunch/docs" // swagger docs
	"github.com/chainlaunch/chainlaunch/pkg/auth"
	backuphttp "github.com/chainlaunch/chainlaunch/pkg/backups/http"
	backupservice "github.com/chainlaunch/chainlaunch/pkg/backups/service"
	"github.com/chainlaunch/chainlaunch/pkg/db"
	fabrichandler "github.com/chainlaunch/chainlaunch/pkg/fabric/handler"
	fabricservice "github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/handler"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	networkshttp "github.com/chainlaunch/chainlaunch/pkg/networks/http"
	networksservice "github.com/chainlaunch/chainlaunch/pkg/networks/service"
	nodeshttp "github.com/chainlaunch/chainlaunch/pkg/nodes/http"
	nodesservice "github.com/chainlaunch/chainlaunch/pkg/nodes/service"
	notificationhttp "github.com/chainlaunch/chainlaunch/pkg/notifications/http"
	notificationservice "github.com/chainlaunch/chainlaunch/pkg/notifications/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	port    int
	dbPath  string
	queries *db.Queries
	dev     bool
	// HTTP TLS configuration variables
	tlsCertFile string
	tlsKeyFile  string
)

// spaHandler implements the http.Handler interface for serving a Single Page Application
type spaHandler struct {
	indexPath  string
	fsys       embed.FS
	staticPath string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute path to prevent directory traversal
	path := filepath.Join(h.staticPath, strings.TrimPrefix(r.URL.Path, "/"))

	// Try to serve the requested file
	content, err := h.fsys.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, serve index.html
		content, err = h.fsys.ReadFile(filepath.Join(h.staticPath, h.indexPath))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Set content type based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	}

	w.Write(content)
}

// @title ChainLaunch API
// @version 1.0
// @description ChainLaunch API provides services for managing blockchain networks and cryptographic keys
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.chainlaunch.com/support
// @contact.email support@chainlaunch.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8100
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.basic BasicAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name session_id

// @tag.name Keys
// @tag.description Cryptographic key management operations

// @tag.name Providers
// @tag.description Key provider management operations

// @tag.name Networks
// @tag.description Blockchain network management operations

// @tag.name Nodes
// @tag.description Network node management operations

// Add these constants at the top level
const (
	keyLength         = 32 // 256 bits
	encryptionKeyFile = "encryption_key"
	sessionKeyFile    = "session_key"
	configDirName     = ".chainlaunch"
)

// Add these new functions
func generateRandomKey(length int) ([]byte, error) {
	// Generate random bytes
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %v", err)
	}

	return key, nil
}

func getConfigDir() (string, error) {
	// First check XDG_CONFIG_HOME
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "chainlaunch"), nil
	}

	// Then check HOME
	home := os.Getenv("HOME")
	if home == "" {
		// Fallback to user home dir
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not find home directory: %v", err)
		}
	}

	// For Linux/Mac: ~/.chainlaunch
	return filepath.Join(home, configDirName), nil
}

func ensureKeyExists(filename string) (string, error) {
	// First check if the key is already set in environment
	envKey := strings.ToUpper(strings.TrimSuffix(filename, "_key"))
	if key := os.Getenv(envKey); key != "" {
		return key, nil
	}

	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(configDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	// Check if key file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Generate new key
		key, err := generateRandomKey(keyLength)
		if err != nil {
			return "", err
		}

		// Encode key as hex string
		keyString := hex.EncodeToString(key)

		// Write key to file with restricted permissions
		if err := ioutil.WriteFile(keyPath, []byte(keyString), 0600); err != nil {
			return "", fmt.Errorf("failed to write key file: %v", err)
		}

		return keyString, nil
	}

	// Read existing key
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %v", err)
	}

	return string(keyBytes), nil
}

// setupServer configures and returns the HTTP server
func setupServer(queries *db.Queries, authService *auth.AuthService, views embed.FS) *chi.Mux {
	// Initialize services
	keyManagementService, err := service.NewKeyManagementService(queries)
	if err != nil {
		log.Fatal("Failed to initialize key management service:", err)
	}
	if err := keyManagementService.InitializeKeyProviders(context.Background()); err != nil {
		log.Fatal("Failed to initialize key providers:", err)
	}

	organizationService := fabricservice.NewOrganizationService(queries, keyManagementService)
	logger := logger.NewDefault()

	nodeEventService := nodesservice.NewNodeEventService(queries, logger)
	nodesService := nodesservice.NewNodeService(queries, logger, keyManagementService, organizationService, nodeEventService)
	networksService := networksservice.NewNetworkService(queries, nodesService, keyManagementService, logger, organizationService)
	notificationService := notificationservice.NewNotificationService(queries, logger)
	backupService := backupservice.NewBackupService(queries, logger, notificationService, dbPath)

	// Initialize handlers
	keyManagementHandler := handler.NewKeyManagementHandler(keyManagementService)
	organizationHandler := fabrichandler.NewOrganizationHandler(organizationService)
	nodesHandler := nodeshttp.NewNodeHandler(nodesService, logger)
	// Start periodic ping service
	networksHandler := networkshttp.NewHandler(
		networksService,
		nodesService,
	)
	backupHandler := backuphttp.NewHandler(backupService)
	notificationHandler := notificationhttp.NewNotificationHandler(notificationService)

	// Setup router
	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"}, // Allow all origins
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		r.Post("/auth/login", auth.LoginHandler(authService))

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(authService))

			r.Post("/auth/logout", auth.LogoutHandler(authService))
			r.Get("/auth/me", auth.GetCurrentUserHandler(authService))

			// Mount key management routes
			keyManagementHandler.RegisterRoutes(r)
			// Mount organization routes
			organizationHandler.RegisterRoutes(r)
			// Mount nodes routes
			nodesHandler.RegisterRoutes(r)
			// Mount networks routes
			networksHandler.RegisterRoutes(r)
			// Mount backups routes
			backupHandler.RegisterRoutes(r)
			// Mount notifications routes
			notificationHandler.RegisterRoutes(r)

		})
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Serve UI static files in production mode
	if !dev {
		spa := spaHandler{
			staticPath: "web/dist",
			indexPath:  "index.html",
			fsys:       views,
		}
		r.Handle("/*", spa)
	}

	return r
}

func runMigrations(database *sql.DB, migrationsFS embed.FS) error {
	driver, err := sqlite3.WithInstance(database, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create sqlite driver: %v", err)
	}

	// Use embedded migrations instead of file system
	d, err := iofs.New(migrationsFS, "pkg/db/migrations")
	if err != nil {
		return fmt.Errorf("could not create iofs driver: %v", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", d,
		"sqlite3", driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %v", err)
	}

	return nil
}

// Command returns the serve command
func Command(configCMD config.ConfigCMD, logger *logger.Logger) *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		Long: `Start the HTTP API server on the specified port.
For example:
  chainlaunch serve --port 8100`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// Ensure the database directory exists
			dbDir := filepath.Dir(dbPath)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				log.Fatalf("Failed to create database directory: %v", err)
			}

			// Initialize database connection
			database, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			// Run migrations
			if err := runMigrations(database, configCMD.MigrationsFS); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}

			// Create queries instance
			queries = db.New(database)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize encryption key
			encryptionKey, err := ensureKeyExists(encryptionKeyFile)
			if err != nil {
				log.Fatalf("Failed to initialize encryption key: %v", err)
			}
			if err := os.Setenv("KEY_ENCRYPTION_KEY", encryptionKey); err != nil {
				log.Fatalf("Failed to set encryption key environment variable: %v", err)
			}

			// Initialize session key
			sessionKey, err := ensureKeyExists(sessionKeyFile)
			if err != nil {
				log.Fatalf("Failed to initialize session key: %v", err)
			}
			if err := os.Setenv("SESSION_ENCRYPTION_KEY", sessionKey); err != nil {
				log.Fatalf("Failed to set session key environment variable: %v", err)
			}

			fmt.Printf("Starting server on port %d...\n", port)
			fmt.Printf("Using database: %s\n", dbPath)
			if dev {
				fmt.Println("Running in development mode")
			} else {
				fmt.Println("Running in production mode")
			}

			// Initialize auth service with database
			authService := auth.NewAuthService(queries)

			// Check if any users exist
			users, err := authService.ListUsers(context.Background())
			if err != nil {
				log.Fatalf("Failed to check existing users: %v", err)
			}

			if len(users) == 0 {
				// No users exist, check for required environment variables
				username := os.Getenv("CHAINLAUNCH_USER")
				password := os.Getenv("CHAINLAUNCH_PASSWORD")

				if username == "" || password == "" {
					log.Fatal("No users found in database. CHAINLAUNCH_USER and CHAINLAUNCH_PASSWORD environment variables must be set for initial user creation")
				}

				// Create initial user with provided credentials
				if err := authService.CreateUser(context.Background(), username, password); err != nil {
					log.Fatalf("Failed to create initial user: %v", err)
				}
				log.Printf("Created initial user with username: %s", username)
			}

			// Setup and start HTTP server
			router := setupServer(queries, authService, configCMD.Views)

			// Start HTTP server in a goroutine
			httpServer := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: router,
			}

			isTLS := tlsCertFile != "" && tlsKeyFile != ""
			// Check if TLS cert and key files exist
			if isTLS {
				if _, err := os.Stat(tlsCertFile); os.IsNotExist(err) {
					log.Fatalf("TLS certificate file not found: %s", tlsCertFile)
				}
				if _, err := os.Stat(tlsKeyFile); os.IsNotExist(err) {
					log.Fatalf("TLS key file not found: %s", tlsKeyFile)
				}
			}
			if isTLS {
				logger.Infof("HTTPS server listening on :%d", port)
				err = httpServer.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
			} else {
				logger.Infof("HTTP server listening on :%d", port)
				err = httpServer.ListenAndServe()
			}

			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			// Clean up database connection
			if queries != nil {
				if err := queries.Close(); err != nil {
					log.Printf("Error closing database connection: %v", err)
				}
			}
		},
	}

	// Add port flags
	serveCmd.Flags().IntVarP(&port, "port", "p", 8100, "Port to run the HTTP server on")

	// Add database path flag
	defaultDBPath := filepath.Join("data", "chainlaunch.db")
	serveCmd.Flags().StringVar(&dbPath, "db", defaultDBPath, "Path to SQLite database file")

	// Add HTTP TLS configuration flags
	serveCmd.Flags().StringVar(&tlsCertFile, "tls-cert", "", "Path to TLS certificate file for HTTP server (required)")
	serveCmd.Flags().StringVar(&tlsKeyFile, "tls-key", "", "Path to TLS key file for HTTP server (required)")

	// Add development mode flag
	serveCmd.Flags().BoolVar(&dev, "dev", false, "Run in development mode")

	return serveCmd
}
