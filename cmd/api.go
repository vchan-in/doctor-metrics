package cmd

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"vchan.in/doctor-metrics/handlers"
)

func Server(version string) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()
	e.HideBanner = true // Hide the echo server banner to avoid server version disclosure in logs

	// Root level middleware
	if os.Getenv("LOG_LEVEL") == "debug" {
		e.Use(middleware.Logger())
	}
	e.Use(middleware.Secure())                                                         // Use secure middleware to set security headers
	e.Use(middleware.Recover())                                                        // Recover middleware recovers from panics anywhere in the chain
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(5)))) // Rate limiter middleware
	e.Use(handlers.HandleAuthMiddleware)                                               // Auth middleware
	e.Use(handlers.HandleCORS)                                                         // CORS middleware

	// Routes
	e.GET("/", func(c echo.Context) error {
		return handlers.GetRoot(c, version)
	})
	e.GET("api/metrics", handlers.GetDockerMetrics)
	e.GET("api/metrics/:containerName", handlers.GetMetricsContainerByName)
	e.GET("api/metrics/:containerID", handlers.GetMetricsContainerByID)

	httpPort := os.Getenv("DM_SERVER_PORT")
	if httpPort == "" {
		httpPort = ":9095" // Default port if not provided
	}
	slog.Info(`
    ____             __             
   / __ \____  _____/ /_____  _____ 
  / / / / __ \/ ___/ __/ __ \/ ___/ 
 / /_/ / /_/ / /__/ /_/ /_/ / /     
/_____/\____/\____\__/\____/_/      
   /  |/  /__  / /______(_)_________
  / /|_/ / _ \/ __/ ___/ / ___/ ___/
 / /  / /  __/ /_/ /  / / /__(__  ) 
/_/  /_/\___/\__/_/  /_/\___/____/  v` + version + `
									
	`)
	slog.Info("Server started at 0.0.0.0:" + httpPort)
	server := e.Start(":" + httpPort)
	if server != nil {
		slog.Error(server.Error())
	}
}
