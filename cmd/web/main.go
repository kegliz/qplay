package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kegnet.dev/qplay/internal/app"
	"kegnet.dev/qplay/internal/config"
)

var (
	//go:embed VERSION.txt
	Version string
)

func main() {
	fmt.Println(Version)
	if err := appMain(); err != nil {
		log.Fatalf("error: %+v", err)
	}
}

// appMain is the main function for the application.
func appMain() error {
	conf, err := config.NewConfig()
	if err != nil {
		return err
	}

	srv, err := app.NewServer(app.ServerOptions{
		C:       conf,
		Version: Version,
	})
	if err != nil {
		return fmt.Errorf("failed to create server: %+v", err)
	}
	go func() {
		if err := srv.Listen(conf.GetInt("port")); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	gracefulShutdownTimeout := time.Duration(conf.GetInt("gracefulshutdowntimeout")) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %s", err)
	}
	log.Println("Server exited gracefully")

	return nil
}
