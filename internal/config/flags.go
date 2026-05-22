package config

import (
	"flag"
	"os"
)

var Options struct {
	AddrServe   string
	AddrResp    string
	StoragePath string
}

func RunFlags() {
	flag.StringVar(&Options.AddrServe, "a", "localhost:8080", "Address server")
	flag.StringVar(&Options.AddrResp, "b", "http://localhost:8080", "Address response")
	flag.StringVar(&Options.StoragePath, "f", "storage.txt", "Storage Path")
	flag.Parse()

	if envServAddr := os.Getenv("SERVER_ADDRESS"); envServAddr != "" {
		Options.AddrServe = envServAddr
	}
	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		Options.AddrResp = envBaseUrl
	}
	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		Options.StoragePath = envStoragePath
	}
}
