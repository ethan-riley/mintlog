package config

import (
	"strings"

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
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + viper.GetString("postgres.port") + "/" + p.DB + "?sslmode=disable"
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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("postgres.user", "mintlog")
	viper.SetDefault("postgres.password", "mintlog")
	viper.SetDefault("postgres.db", "mintlog")
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("opensearch.url", "http://localhost:9200")
	viper.SetDefault("opensearch.user", "admin")
	viper.SetDefault("opensearch.password", "M1ntl0g!Pass")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.accesskey", "minioadmin")
	viper.SetDefault("minio.secretkey", "minioadmin")
	viper.SetDefault("minio.usessl", false)
	viper.SetDefault("ingest.addr", ":8080")
	viper.SetDefault("api.addr", ":8081")

	// Try reading .env file; ignore if not found
	_ = viper.ReadInConfig()

	cfg := &Config{
		Postgres: PostgresConfig{
			Host:     viper.GetString("postgres.host"),
			Port:     viper.GetInt("postgres.port"),
			User:     viper.GetString("postgres.user"),
			Password: viper.GetString("postgres.password"),
			DB:       viper.GetString("postgres.db"),
		},
		NATS: NATSConfig{
			URL: viper.GetString("nats.url"),
		},
		OpenSearch: OpenSearchConfig{
			URL:      viper.GetString("opensearch.url"),
			User:     viper.GetString("opensearch.user"),
			Password: viper.GetString("opensearch.password"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		MinIO: MinIOConfig{
			Endpoint:  viper.GetString("minio.endpoint"),
			AccessKey: viper.GetString("minio.accesskey"),
			SecretKey: viper.GetString("minio.secretkey"),
			UseSSL:    viper.GetBool("minio.usessl"),
		},
		Ingest: ServerConfig{
			Addr: viper.GetString("ingest.addr"),
		},
		API: ServerConfig{
			Addr: viper.GetString("api.addr"),
		},
	}

	return cfg, nil
}
