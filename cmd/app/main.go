package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/handlers"
	"github.com/UserNameShouldBeHere/AvitoTask/internal/infrastructure/postgres"
	"github.com/UserNameShouldBeHere/AvitoTask/internal/services"
)

const (
	backEndPort       = 8080
	sessionExpiration = 60
)

func main() {
	var (
		dbUser     string
		dbPassword string
	)

	flag.StringVar(&dbUser, "dbuser", "postgres", "database user")
	flag.StringVar(&dbPassword, "dbpass", "root1234", "database password")

	flag.Parse()

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}
	sugarLogger := logger.Sugar()

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		"5432",
		dbUser,
		dbPassword,
		"shop",
	))
	if err != nil {
		log.Fatalf("error in postgres initialization: %v\n", err)
	}

	authStorage, err := postgres.NewAuthStorage(pool)
	if err != nil {
		log.Fatalf("error in auth storage initialization: %v\n", err)
	}
	shopStorage, err := postgres.NewShopStorage(pool)
	if err != nil {
		log.Fatalf("error in shop storage initialization: %v\n", err)
	}

	authService, err := services.NewAuthService(authStorage, sugarLogger, 10, sessionExpiration)
	if err != nil {
		log.Fatalf("error in auth service initialization: %v\n", err)
	}

	shopService, err := services.NewShopService(shopStorage, sugarLogger)
	if err != nil {
		log.Fatalf("error in auth service initialization: %v\n", err)
	}

	authHandler, err := handlers.NewAuthHandler(authService, sugarLogger, sessionExpiration)
	if err != nil {
		log.Fatalf("error in auth handler initialization: %v\n", err)
	}
	shopHandler, err := handlers.NewShopHandler(authService, shopService, sugarLogger)
	if err != nil {
		log.Fatalf("error in shop handler initialization: %v\n", err)
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /api/info", shopHandler.Info)
	router.HandleFunc("POST /api/auth", authHandler.Auth)
	router.HandleFunc("POST /api/sendCoin", shopHandler.SendCoin)
	router.HandleFunc("POST /api/buy/{item}", shopHandler.BuyItem)

	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", backEndPort),
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Server shutdown error: %v\n", err)
		}
	}()

	fmt.Printf("Starting server at %s%s\n", "localhost", fmt.Sprintf(":%d", backEndPort))

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-stopped

	fmt.Println("Server stopped")
}
