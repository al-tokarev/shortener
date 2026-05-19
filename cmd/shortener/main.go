package main

import (
	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/router"
)

func main() {
	if err := run(); err != nil {
		logger.Sugar.Fatalw(err.Error(), "event", "start server")
		panic(err)
	}
}

func run() error {
	logger.Initialize()
	config.RunFlags()
	return router.GoRouter()
}
