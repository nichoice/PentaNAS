package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pnas/internal/config"
	"pnas/internal/controllers"
	"pnas/internal/database"
	"pnas/internal/logging"
	"pnas/internal/middleware"
	"pnas/internal/routes"
	"pnas/internal/services"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HarborArk Web Server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	serverCmd.Flags().Bool("debug", false, "Enable debug mode")
	serverCmd.Flags().String("config", "config.yaml", "Path to config file")
	rootCmd.AddCommand(serverCmd)
}

// @title HarborArk
// @version 0.0.1
// @description HarborArk系统API文档
// @host localhost:8080
// @BasePath /api/v1

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func startServer() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logging.InitLogger(cfg.Server.Debug); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Sync()

	// Initialize database
	if err := database.InitDatabase(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize system (create default roles and admin user if needed)
	initService := services.NewInitService()
	if err := initService.InitializeSystem(); err != nil {
		log.Fatalf("Failed to initialize system: %v", err)
	}

	// Initialize monitoring services
	controllers.InitMonitoringServices()
	defer controllers.CleanupMonitoringServices()

	// Set Gin mode
	if cfg.Server.Debug {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin engine
	router := gin.New()

	// Use middleware
	router.Use(middleware.CORSMiddleware())

	// Setup routes
	routes.SetupRoutes(router)

	// Start server
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Server starting on", addr)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
	select {
	case <-ctx.Done():
		log.Println("Server shutdown completed")
	case <-time.After(5 * time.Second):
		log.Println("Server shutdown timed out")
	}
}
