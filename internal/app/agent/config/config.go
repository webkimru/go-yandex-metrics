package config

type AppConfig struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
	SecretKey      string
}
