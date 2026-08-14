package config

import "os"

type Config struct {
	Port      string
	UploadDir string
	DBPath    string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/travel.db"
	}

	return &Config{
		Port:      port,
		UploadDir: uploadDir,
		DBPath:    dbPath,
	}
}