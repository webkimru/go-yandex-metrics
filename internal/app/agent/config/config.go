package config

import (
	"crypto/rsa"
)

type AppConfig struct {
	SecretKey      string         `json:"key,omitempty"`
	ServerAddress  string         `json:"address,omitempty"`
	CryptoKey      string         `json:"crypto_key,omitempty"`
	PublicKeyPEM   *rsa.PublicKey `json:"-"`
	RealIP         string         `json:"real_ip,omitempty"`
	RateLimit      int            `json:"rate_limit,omitempty"`
	PollInterval   int            `json:"poll_interval,omitempty"`
	ReportInterval int            `json:"report_interval,omitempty"`
}
