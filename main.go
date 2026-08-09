package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/kouleen/lets-encrypt/internal/repository"
	"github.com/kouleen/lets-encrypt/internal/routes"
	_ "github.com/kouleen/lets-encrypt/pkg/util"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	routes.Register(r)

	srv := &http.Server{
		Addr:         ":8099",
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server failed to start", err)
		}
	}()
	log.Println("Server started on port 8099")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", err)
	}

	log.Println("Server exiting")
}
