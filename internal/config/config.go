package config

import (
	"flag"
	"os"
	"strconv"
)

var Options struct {
	AddrServe   string
	AddrResp    string
	StoragePath string
	DatabaseDSN string
	AuditFile   string
	AuditURL    string
	EnableHTTPS bool
}

func RunFlags() {
	flag.StringVar(&Options.AddrServe, "a", "localhost:8080", "Address server")
	flag.StringVar(&Options.AddrResp, "b", "http://localhost:8080", "Address response")
	flag.StringVar(&Options.StoragePath, "f", "storage.txt", "Storage Path")
	flag.StringVar(&Options.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&Options.AuditFile, "audit-file", "", "audit file listener")
	flag.StringVar(&Options.AuditURL, "audit-url", "", "audit url listener")
	flag.BoolVar(&Options.EnableHTTPS, "s", Options.EnableHTTPS, "enable HTTPS")
	flag.Parse()

	if envServAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		Options.AddrServe = envServAddr
	}
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		Options.AddrResp = envBaseURL
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
	if envEnableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if parsed, err := strconv.ParseBool(envEnableHTTPS); err == nil {
			Options.EnableHTTPS = parsed
		}
	}
}
