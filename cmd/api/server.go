package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Create a shutdownError channel. Will be used to receive any errors
	// returned by graceful Shutdown() function
	shutdownError := make(chan error)

	// Start a background goroutine
	go func() {
		// Intercept the signals like before
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		// Update the log entry to read "shutting down server" instead of "caught signal"
		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the context we just made
		// It will return nil if the graceful shutdown is successful, or an error
		shutdownError <- srv.Shutdown(ctx)
	}()

	// Log a "starting server" message
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Calling Shutdown() on our server will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. It's a good thing as it indicates that the
	// graceful shutdown started. We check specifically for this if not return error
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, wait to receive return value from Shutdown() on the shutdownError channel
	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point, graceful shutdown completed successfully amd log a "stopped server" message
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
