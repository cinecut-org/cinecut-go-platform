package platform

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Serve runs handler on :PORT (default 8080) until SIGINT/SIGTERM, then gives in-flight requests 20 seconds.
func Serve(service string, handler http.Handler) error {
	addr := ":" + Getenv("PORT", "8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		slog.Info("listening", "service", service, "addr", addr)
		errs <- srv.ListenAndServe()
	}()
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}
	slog.Info("shutting down", "service", service)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
