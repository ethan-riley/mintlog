package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Postgres   PostgresConfig
	NATS       NATSConfig
	OpenSearch OpenSearchConfig
	Redis      RedisConfig
	MinIO      MinIOConfig
	Ingest     ServerConfig
	API        ServerConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

func (p PostgresConfig) DSN() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + viper.GetString("postgres_port") + "/" + p.DB + "?sslmode=disable"
}

type NATSConfig struct {
	URL string
}

type OpenSearchConfig struct {
	URL      string
	User     string
	Password string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type ServerConfig struct {
	Addr string
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Defaults — keys use underscores to match .env file keys
	viper.SetDefault("postgres_host", "localhost")
	viper.SetDefault("postgres_port", 6543)
	viper.SetDefault("postgres_user", "mintlog")
	viper.SetDefault("postgres_password", "mintlog")
	viper.SetDefault("postgres_db", "mintlog")
	viper.SetDefault("nats_url", "nats://localhost:4222")
	viper.SetDefault("opensearch_url", "http://localhost:9200")
	viper.SetDefault("opensearch_user", "admin")
	viper.SetDefault("opensearch_password", "M1ntl0g!Pass")
	viper.SetDefault("redis_addr", "localhost:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("minio_endpoint", "localhost:9000")
	viper.SetDefault("minio_access_key", "minioadmin")
	viper.SetDefault("minio_secret_key", "minioadmin")
	viper.SetDefault("minio_use_ssl", false)
	viper.SetDefault("ingest_addr", ":8080")
	viper.SetDefault("api_addr", ":8081")

	// Try reading .env file; ignore if not found
	_ = viper.ReadInConfig()

	cfg := &Config{
		Postgres: PostgresConfig{
			Host:     viper.GetString("postgres_host"),
			Port:     viper.GetInt("postgres_port"),
			User:     viper.GetString("postgres_user"),
			Password: viper.GetString("postgres_password"),
			DB:       viper.GetString("postgres_db"),
		},
		NATS: NATSConfig{
			URL: viper.GetString("nats_url"),
		},
		OpenSearch: OpenSearchConfig{
			URL:      viper.GetString("opensearch_url"),
			User:     viper.GetString("opensearch_user"),
			Password: viper.GetString("opensearch_password"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("redis_addr"),
			Password: viper.GetString("redis_password"),
			DB:       viper.GetInt("redis_db"),
		},
		MinIO: MinIOConfig{
			Endpoint:  viper.GetString("minio_endpoint"),
			AccessKey: viper.GetString("minio_access_key"),
			SecretKey: viper.GetString("minio_secret_key"),
			UseSSL:    viper.GetBool("minio_use_ssl"),
		},
		Ingest: ServerConfig{
			Addr: viper.GetString("ingest_addr"),
		},
		API: ServerConfig{
			Addr: viper.GetString("api_addr"),
		},
	}

	return cfg, nil
}
