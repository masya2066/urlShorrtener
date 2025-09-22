package models

type Config struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	CookieName      string `json:"cookie_name"`
	AuthSecret      string `json:"auth_secret"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
}
