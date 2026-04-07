package main

import (
	"oss-gateway/internal/config"
	"oss-gateway/internal/database"
	"oss-gateway/internal/logger"
	"oss-gateway/internal/server"
	"oss-gateway/pkg/crypto"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg := config.NewConfig()
	log := logger.NewLogger(cfg)
	aesCrypto, err := crypto.NewAESCrypto(cfg.CryptoConfig.AESKey)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize AES crypto")
	}

	db := database.NewGorm(cfg, log.WithField("module", "database"))

	srv := server.NewServer(cfg, log, db, aesCrypto)
	srv.Start()
}
