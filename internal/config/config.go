package config

import (
    "os"
)

type Config struct {
    Addr     string
    CertFile string
    KeyFile  string
    DSN      string
}

func New() Config {
    // Получаем значения из переменных окружения или используем значения по умолчанию
    addr := os.Getenv("SERVER_ADDR")
    if addr == "" {
        addr = ":8443"
    }

    certFile := os.Getenv("CERT_FILE")
    if certFile == "" {
        certFile = "certs/server.crt"
    }

    keyFile := os.Getenv("KEY_FILE")
    if keyFile == "" {
        keyFile = "certs/server.key"
    }

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // Формат: postgres://username:password@localhost:5432/database?sslmode=disable
        dsn = "postgres://your_username:your_password@localhost:5432/study_security?sslmode=disable"
    }

    return Config{
        Addr:     addr,
        CertFile: certFile,
        KeyFile:  keyFile,
        DSN:      dsn,
    }
}
