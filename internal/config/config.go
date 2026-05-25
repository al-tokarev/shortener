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

	if envServAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		Options.AddrServe = envServAddr
	}
	if envBaseUrl, ok := os.LookupEnv("BASE_URL"); ok {
		Options.AddrResp = envBaseUrl
	}
	if envStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		Options.StoragePath = envStoragePath
	}
}
