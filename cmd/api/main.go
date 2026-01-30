package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/azizjon12/greenlight/internal/data"
	_ "github.com/lib/pq"
)

// Declare a string containing the application version number.
// Later we will try to generate this automatically at build time
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}

	// Add a new limiter struct containing fields for the requests-per-second and bursts values
	// and a boolean which can use to enable/disable rate limiting
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware.
type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	// Declare an instance of the config struct.
	var cfg config

	// Read the value of the port and env command-line flags into config structs.
	// Default port number 4000 and "development" env are used if not corresponding flags are provided
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PosgreSQL DSN")

	// Read the connection pool settings from command-line flags into the config struct.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// Create command-line flags to read settings values into the config struct
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	// Initialize a new structured logger which writes log entries to the std out stream
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create db connection
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Defer a call to db.Close() so the connection pool is closed before the main() exits
	defer db.Close()

	// Log a message to say that the connection has been successful
	logger.Info("database connection pool established")

	// Declare an instance of the application struct, containing the config struct and logger
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Call app.serve() to start the server
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Use the httprouter instance returned by app.routes() as the server handler
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Start the HTTP server
	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// It returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in pool. No limit if <= 0
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Set the maximum idle timeout for the connection pool. If <= means conn not closed due to their idle time
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// Create a context with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Establish a new connection to the db, passing in the context as parameter.
	// If not connected within 5 seconds, will return error, and close the connection
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Return the sql.DB connection pool
	return db, nil
}
