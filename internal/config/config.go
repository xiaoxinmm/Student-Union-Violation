package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string
	UploadDir  string
	MaxUpload  int64 // bytes
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "suv"),
		DBPassword: getEnv("DB_PASSWORD", "suv_password"),
		DBName:     getEnv("DB_NAME", "suv"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production-32chars!"),
		Port:       getEnv("PORT", "8080"),
		UploadDir:  getEnv("UPLOAD_DIR", "./uploads"),
		MaxUpload:  5 * 1024 * 1024, // 5MB
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
