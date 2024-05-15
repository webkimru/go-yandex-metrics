package config

import (
	"crypto/rsa"
)

type Store int

const (
	Database Store = iota + 1
	File
	Memory
)

type RecorderConfig struct {
	FilePath string `json:"filepath"`
	Interval int    `json:"interval"`
	Restore  bool   `json:"restore"`
}

type AppConfig struct {
	ServerAddress string          `json:"address,omitempty"`
	SecretKey     string          `json:"key,omitempty"`
	CryptoKey     string          `json:"crypto_key,omitempty"`
	PrivateKeyPEM *rsa.PrivateKey `json:"-"`
	TrustedSubnet string          `json:"trusted_subnet,omitempty"`
	DatabaseDSN   string          `json:"database_dsn,omitempty"`
	FileStore     RecorderConfig  `json:"store_file"`
	StorePriority Store           `json:"-"`
}
