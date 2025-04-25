package main

import (
	"errors"
	"final-project/config"
	_ "final-project/docs"
	"final-project/utils/helpers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// @title           ToyRental API
// @version         1.0
// @description     REST API for toy rental service

// @contact.name   Indrawansyah
// @contact.email  indra@dev.com

// @host      localhost:8080
// @BasePath  /api
// @securityDefinitions.cookie ApiCookieAuth
// @in cookie
// @name access_tokend
func main() {
	// Load konfigurasi
	cfg := config.LoadConfig()

	// Setup logger
	helpers.SetupLogger(cfg.IsProd)
	log := helpers.Logger

	// Inisialisasi database
	db := config.NewDatabase(cfg)

	// Auto migrate
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Setup routes
	r := setupRoutes(cfg, db.DB)
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Buat channel untuk menangkap signal interupsi
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Jalankan server di goroutine terpisah
	go func() {
		log.Printf("Server running on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()

	// Tunggu signal untuk shutdown
	<-quit
	log.Println("Shutting down server...")

	// Tutup koneksi database
	db.CloseConnection()

	log.Println("Server exited properly")
}
