package config

import (
	"flag"
	"os"
)

var Options struct {
	AddrServe   string
	AddrResp    string
	StoragePath string
	DatabaseDSN string
	AuditFile   string
	AuditURL    string
}

func RunFlags() {
	flag.StringVar(&Options.AddrServe, "a", "localhost:8080", "Address server")
	flag.StringVar(&Options.AddrResp, "b", "http://localhost:8080", "Address response")
	flag.StringVar(&Options.StoragePath, "f", "storage.txt", "Storage Path")
	flag.StringVar(&Options.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&Options.AuditFile, "audit-file", "", "audit file listener")
	flag.StringVar(&Options.AuditURL, "audit-url", "", "audit url listener")
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
	if envDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		Options.DatabaseDSN = envDatabaseDSN
	}
	if envAuditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		Options.AuditFile = envAuditFile
	}
	if envAuditURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		Options.AuditURL = envAuditURL
	}
}
