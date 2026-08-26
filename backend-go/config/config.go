package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port          string
	Listen        string
	UploadDir     string
	DBPath        string
	StaticDir     string
	JWTSecret     string
	PublicURL     string
	AmapKey       string
	AdminUsername string
	AdminPassword string
	LimePassword  string
	EiinPassword  string
}

func Load() *Config {
	loadDotEnv(".env", "../.env")
	port := envOr("PORT", "5000")
	return &Config{
		Port:          port,
		Listen:        envOr("LISTEN", ""),
		UploadDir:     envOr("UPLOAD_DIR", "./uploads"),
		DBPath:        envOr("DB_PATH", "./data/travel.db"),
		StaticDir:     envOr("STATIC_DIR", "./dist"),
		JWTSecret:     envOr("JWT_SECRET", "travel-footprints-dev-secret-change-me"),
		PublicURL:     envOr("PUBLIC_URL", ""),
		AmapKey:       envOr("AMAP_KEY", envOr("VITE_AMAP_KEY", "")),
		AdminUsername: envOr("ADMIN_USERNAME", "admin"),
		AdminPassword: envOr("ADMIN_PASSWORD", ""),
		LimePassword:  envOr("LIME_PASSWORD", ""),
		EiinPassword:  envOr("EIINXYZ_PASSWORD", ""),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(paths ...string) {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, val, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
			val = strings.Trim(val, `"'`)
			if key != "" && os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		_ = f.Close()
	}
}
