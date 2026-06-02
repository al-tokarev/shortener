package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"github.com/al-tokarev/shortener/internal/router"
	"github.com/al-tokarev/shortener/internal/service/urlservices"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	config.RunFlags()
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	conn, err := sql.Open("pgx", config.Options.DatabaseDSN)
	if err != nil {
		logger.Fatal(zap.Error(err))
	}

	repository := urlrepository.NewRepository(conn, logger)
	service := urlservices.NewService(repository, logger)
	handler := urlhandlers.NewHandler(service, logger)

	if err := repository.InitializeStorage(); err != nil {
		logger.Fatalw("failed to initialize storage: %v", err)
	}

	r := router.NewRouter(handler, logger)
	server := &http.Server{
		Addr:              config.Options.AddrServe,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	logger.Infow("Server is starting", "addr", server.Addr)
	return server.ListenAndServe()
}
