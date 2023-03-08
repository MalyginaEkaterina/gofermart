package app

import (
	"context"
	"database/sql"
	"flag"
	"github.com/MalyginaEkaterina/gofermart/internal"
	"github.com/MalyginaEkaterina/gofermart/internal/handlers"
	"github.com/MalyginaEkaterina/gofermart/internal/service"
	"github.com/MalyginaEkaterina/gofermart/internal/storage"
	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
)

func Start() {
	var cfg internal.Config
	var secretFilePath string
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen on")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database connection string")
	flag.StringVar(&cfg.AccrualAddress, "r", "", "accrual system address")
	flag.StringVar(&secretFilePath, "s", "", "path to file with secret")
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal("Error while parsing env: ", err)
	}
	db, userStore, orderStore := initStore(cfg)
	defer db.Close()
	defer userStore.Close()
	defer orderStore.Close()
	secretKey, err := getSecret(secretFilePath)
	if err != nil {
		log.Fatal("Error while reading secret key", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authService := &service.AuthServiceImpl{Store: userStore, SecretKey: secretKey}
	orderService := &service.OrderServiceImpl{Store: orderStore}

	r := handlers.NewRouter(authService, orderService)
	orderWorker := service.NewOrderWorker(service.NewAccrualClient(cfg.AccrualAddress), orderStore)
	go orderWorker.Run(ctx)
	log.Printf("Started server on %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}

func initStore(cfg internal.Config) (*sql.DB, storage.UserStorage, storage.OrderStorage) {
	if cfg.DatabaseURI == "" {
		log.Fatal("Database URI must be configured")
	}
	db, err := sql.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		log.Fatal("Database connection error: ", err)
	}
	err = storage.DoMigrations(db)
	if err != nil {
		log.Fatal("Running migrations error: ", err)
	}
	log.Printf("Using database storage %s\n", cfg.DatabaseURI)
	userStore, err := storage.NewDBUserStorage(db)
	if err != nil {
		log.Fatal("Create user store error: ", err)
	}
	orderStore, err := storage.NewDBOrderStorage(db)
	if err != nil {
		log.Fatal("Create order store error: ", err)
	}
	return db, userStore, orderStore
}

func getSecret(path string) ([]byte, error) {
	if path == "" {
		// Only for tests.
		return []byte("my secret key"), nil
	}
	return os.ReadFile(path)
}
