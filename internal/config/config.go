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
	if envConfig, ok := os.LookupEnv("CONFIG"); ok {
		pathJSONConfig = envConfig
	}

	flagAddrServe := flag.String("a", "localhost:8080", "Address server")
	flagAddrResp := flag.String("b", "http://localhost:8080", "Address response")
	flagStoragePath := flag.String("f", "storage.txt", "Storage Path")
	flagDatabaseDSN := flag.String("d", "", "Database connection string")
	flagAuditFile := flag.String("audit-file", "", "audit file listener")
	flagAuditURL := flag.String("audit-url", "", "audit url listener")
	flagEnableHTTPS := flag.Bool("s", false, "enable HTTPS")
	flag.StringVar(&pathJSONConfig, "c", pathJSONConfig, "config path")
	flag.StringVar(&pathJSONConfig, "config", pathJSONConfig, "config path")

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

	passed := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		passed[f.Name] = true
	})

	if passed["a"] {
		Options.AddrServe = *flagAddrServe
	}
	if passed["b"] {
		Options.AddrResp = *flagAddrResp
	}
	if passed["f"] {
		Options.StoragePath = *flagStoragePath
	}
	if passed["d"] {
		Options.DatabaseDSN = *flagDatabaseDSN
	}
	if passed["audit-file"] {
		Options.AuditFile = *flagAuditFile
	}
	if passed["audit-url"] {
		Options.AuditURL = *flagAuditURL
	}
	if passed["s"] {
		Options.EnableHTTPS = *flagEnableHTTPS
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
