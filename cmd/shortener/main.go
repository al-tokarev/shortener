package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/migrations"
	"github.com/al-tokarev/shortener/internal/observer"
	"github.com/al-tokarev/shortener/internal/observer/listeners/audit_listeners"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"github.com/al-tokarev/shortener/internal/router"
	"github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	go func() {
		log.Println("pprof server starting on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	runtime.GC()

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

	dispatcher := observer.NewDispatcher()
	registerEvents(dispatcher)

	var conn *sql.DB
	var repository urlrepository.RepositoryInterface
	if config.Options.DatabaseDSN != "" {
		if err := runMigrations(config.Options.DatabaseDSN); err != nil {
			logger.Fatalw("Failed to run migrations", "error", err)
		}
		conn, err = sql.Open("pgx", config.Options.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("failed to open db: %w", err)
		}
		if err = conn.Ping(); err != nil {
			return fmt.Errorf("failed to ping db: %w", err)
		}

		repository = urlrepository.NewDbRepository(conn, logger)
	} else {
		repository = urlrepository.NewLocalRepository(logger)
		if err := repository.InitializeStorage(); err != nil {
			logger.Fatalw("failed to initialize storage: %v", err)
		}
	}

	service := urlservices.NewService(repository, logger)
	handler := urlhandlers.NewHandler(service, dispatcher, logger)

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

func runMigrations(dsn string) error {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

func registerEvents(d *observer.Dispatcher) {
	if config.Options.AuditFile != "" {
		auditFileListener := audit_listeners.AuditFileListener{
			Path: config.Options.AuditFile,
		}
		d.Subscribe("audit", auditFileListener)
	}
	if config.Options.AuditURL != "" {
		auditURLListener := audit_listeners.AuditURLListener{
			URL: config.Options.AuditURL,
		}
		d.Subscribe("audit", auditURLListener)
	}
}
