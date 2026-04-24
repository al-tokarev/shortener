package main

import (
	"github.com/al-tokarev/shortener.git/internal/router"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return router.GoRouter()
}
