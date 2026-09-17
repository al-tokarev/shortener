package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

var Options struct {
	AddrServe   string `json:"server_address"`
	AddrResp    string `json:"base_url"`
	StoragePath string `json:"file_storage_path"`
	DatabaseDSN string `json:"database_dsn"`
	AuditFile   string
	AuditURL    string
	EnableHTTPS bool `json:"enable_https"`
}

func RunFlags() error {
	pathJSONConfig := ""

	flag.StringVar(&Options.AddrServe, "a", "localhost:8080", "Address server")
	flag.StringVar(&Options.AddrResp, "b", "http://localhost:8080", "Address response")
	flag.StringVar(&Options.StoragePath, "f", "storage.txt", "Storage Path")
	flag.StringVar(&Options.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&Options.AuditFile, "audit-file", "", "audit file listener")
	flag.StringVar(&Options.AuditURL, "audit-url", "", "audit url listener")
	flag.BoolVar(&Options.EnableHTTPS, "s", Options.EnableHTTPS, "enable HTTPS")
	flag.StringVar(&pathJSONConfig, "c", pathJSONConfig, "path for config from json")
	flag.StringVar(&pathJSONConfig, "config", pathJSONConfig, "path for config from json")
	flag.Parse()

	if pathJSONConfig != "" {
		file, err := os.Open(pathJSONConfig)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&Options); err != nil {
			return err
		}
	}

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

	return nil
}
