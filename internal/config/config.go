package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// REQ-062: Configuration validation on startup
// REQ-063: Sensible defaults for optional settings
// REQ-064: Multiple environments support
type Config struct {
	Environment string         `mapstructure:"environment" validate:"oneof=development staging production"`
	LogLevel    string         `mapstructure:"log_level" validate:"oneof=debug info warn error"`
	Database    DatabaseConfig `mapstructure:"database"`
	Alpaca      AlpacaConfig   `mapstructure:"alpaca"`
	Server      ServerConfig   `mapstructure:"server"`
	Worker      WorkerConfig   `mapstructure:"worker"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host" validate:"required"`
	Port            int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	User            string `mapstructure:"user" validate:"required"`
	Password        string `mapstructure:"password" validate:"required"`
	Name            string `mapstructure:"name" validate:"required"`
	SSLMode         string `mapstructure:"ssl_mode" validate:"oneof=disable require verify-ca verify-full"`
	MaxConnections  int    `mapstructure:"max_connections" validate:"min=1,max=100"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" validate:"min=60"`
}

type AlpacaConfig struct {
	APIKey    string `mapstructure:"api_key" validate:"required"`
	SecretKey string `mapstructure:"secret_key" validate:"required"`
	BaseURL   string `mapstructure:"base_url" validate:"required,url"`
	WSBaseURL string `mapstructure:"ws_base_url" validate:"required"`
	IsPaper   bool   `mapstructure:"is_paper"`
	UseMock   bool   `mapstructure:"use_mock"`
}

type ServerConfig struct {
	HTTPPort      int    `mapstructure:"http_port" validate:"min=1024,max=65535"`
	WebSocketPort int    `mapstructure:"websocket_port" validate:"min=1024,max=65535"`
	Host          string `mapstructure:"host"`
	ReadTimeout   int    `mapstructure:"read_timeout" validate:"min=1"`
	WriteTimeout  int    `mapstructure:"write_timeout" validate:"min=1"`
	EnableCORS    bool   `mapstructure:"enable_cors"`
}

type WorkerConfig struct {
	BufferSize          int `mapstructure:"buffer_size" validate:"min=100,max=10000"`
	MaxWorkersPerSymbol int `mapstructure:"max_workers_per_symbol" validate:"min=1,max=10"`
	AggregationTimeout  int `mapstructure:"aggregation_timeout" validate:"min=1,max=60"`
}

// REQ-061: Load configuration from .env files and environment variables
func Load() (*Config, error) {
	// Load .env file if exists (development)
	if err := godotenv.Load("config/.env"); err != nil {
		// Don't fail if .env doesn't exist in production
		if os.Getenv("ENVIRONMENT") == "" {
			fmt.Printf("Warning: No .env file found, using environment variables only\n")
		}
	}

	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly bind environment variables for nested structs
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.name", "DATABASE_NAME")
	viper.BindEnv("database.ssl_mode", "DATABASE_SSL_MODE")
	viper.BindEnv("database.max_connections", "DATABASE_MAX_CONNECTIONS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")

	// Alpaca configuration binding
	viper.BindEnv("alpaca.api_key", "ALPACA_API_KEY")
	viper.BindEnv("alpaca.secret_key", "ALPACA_SECRET_KEY")
	viper.BindEnv("alpaca.base_url", "ALPACA_BASE_URL")
	viper.BindEnv("alpaca.ws_base_url", "ALPACA_WS_BASE_URL")
	viper.BindEnv("alpaca.is_paper", "ALPACA_IS_PAPER")
	viper.BindEnv("alpaca.use_mock", "ALPACA_USE_MOCK")

	// Server configuration binding
	viper.BindEnv("server.http_port", "SERVER_HTTP_PORT")
	viper.BindEnv("server.websocket_port", "SERVER_WEBSOCKET_PORT")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	// Worker configuration binding
	viper.BindEnv("worker.buffer_size", "WORKER_BUFFER_SIZE")
	viper.BindEnv("worker.max_workers_per_symbol", "WORKER_MAX_WORKERS_PER_SYMBOL")
	viper.BindEnv("worker.aggregation_timeout", "WORKER_AGGREGATION_TIMEOUT")

	// REQ-063: Set sensible defaults
	setDefaults()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// REQ-062: Validate configuration on startup
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// REQ-062: Configuration validation
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return errors.New("database host is required")
	}
	if c.Database.Port == 0 {
		return errors.New("database port is required")
	}

	// Validate Alpaca configuration if streaming is enabled
	if c.Alpaca.APIKey == "" {
		return errors.New("alpaca API key is required")
	}
	if c.Alpaca.SecretKey == "" {
		return errors.New("alpaca secret key is required")
	}
	if c.Alpaca.BaseURL == "" {
		return errors.New("alpaca base URL is required")
	}

	if c.Server.HTTPPort == 0 {
		return errors.New("HTTP port is required")
	}

	return nil
}

// REQ-065: Mask sensitive values in logs
func (c *Config) String() string {
	masked := *c
	masked.Database.Password = "***"
	masked.Alpaca.APIKey = "***"
	masked.Alpaca.SecretKey = "***"
	return fmt.Sprintf("%+v", masked)
}

func setDefaults() {
	// Environment
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "jonbu_ohlcv")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 300)

	// Alpaca defaults
	viper.SetDefault("alpaca.base_url", "https://paper-api.alpaca.markets")
	viper.SetDefault("alpaca.ws_base_url", "wss://stream.data.alpaca.markets")
	viper.SetDefault("alpaca.is_paper", true)
	viper.SetDefault("alpaca.use_mock", false)

	// Server defaults
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.websocket_port", 8081)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 10)
	viper.SetDefault("server.enable_cors", true)

	// Worker defaults
	viper.SetDefault("worker.buffer_size", 1000)
	viper.SetDefault("worker.max_workers_per_symbol", 5)
	viper.SetDefault("worker.aggregation_timeout", 5)
}
