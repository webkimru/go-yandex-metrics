package config

type Store int

const (
	Database Store = iota + 1
	File
	Memory
)

type RecorderConfig struct {
	Interval int
	Restore  bool
	FilePath string
}

type AppConfig struct {
	ServerAddress string
	SecretKey     string
	DatabaseDSN   string
	StorePriority Store
	FileStore     RecorderConfig
}
