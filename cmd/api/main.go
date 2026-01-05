package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Declare a string containing the application version number.
// Later we will try to generate this automatically at build time
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
type config struct {
	port int
	env  string
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware.
type application struct {
	config config
	logger *slog.Logger
}

func main() {
	// Declare an instance of the config struct.
	var cfg config

	// Read the value of the port and env command-line flags into config structs.
	// Default port number 4000 and "development" env are used if not corresponding flags are provided
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new structured logger which writes log entries to the std out stream
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Declare an instance of the application struct, containing the config struct and logger
	app := application{
		config: cfg,
		logger: logger,
	}

	// Declare a new servemux and add a /v1/healthcheck route that shows requests to the
	// healthcheckHandler method
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Declare an HTTP server which listens on the port provided in the config struct,
	// uses the servemux we created above as the handler, has some sensible timeout settings
	// and writes any log messages to the structured logger at Error level
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Start the HTTP server
	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
