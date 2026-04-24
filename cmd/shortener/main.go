package main

import (
	"github.com/al-tokarev/shortener/internal/router"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return router.GoRouter()
}
