package model

type Config struct {
	Address       string `json:"address"`
	DBFilename    string `json:"db_filename"`
	LogFilename   string `json:"log_filename"`
	XorSecretKey  int64  `json:"xor_secret_key"`
	ShuffleKey    string `json:"shuffle_key"`
	LogLevel      string `json:"log_level"`
	Production    bool   `json:"production"`
	CacheCapacity int    `json:"cache_capacity"`
}
