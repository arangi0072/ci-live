package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

//
// Config
//

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MediaMTX MediaMTXConfig
	MinIO    MinIOConfig
	CDN      CDNConfig
}

//
// App Configuration
//

type AppConfig struct {
	Name        string
	Environment string
	Version     string
}

//
// Server Configuration
//

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

//
// PostgreSQL Configuration
//

type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

//
// Redis Configuration
//

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
}

//
// JWT Configuration
//

type JWTConfig struct {
	PrivateKeyPath       string
	PublicKeyPath        string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
	Audience             string
}

//
// MediaMTX Configuration
//

type MediaMTXConfig struct {
	BaseURL        string
	APIURL         string
	RTMPPublishURL string
	HLSPlaybackURL string
	Username       string
	Password       string
}

//
// MinIO Configuration
//

type MinIOConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	Region          string
	Bucket          string
	PublicBaseURL   string
	PresignDuration time.Duration
}

//
// CDN Configuration
//

type CDNConfig struct {
	BaseURL string
}

//
// Load Configuration
//

func Load() (*Config, error) {

	//
	// Viper defaults
	//

	setDefaults()

	//
	// Environment variables
	//

	viper.SetEnvPrefix("CI")

	viper.SetEnvKeyReplacer(
		strings.NewReplacer(
			".",
			"_",
		),
	)

	viper.AutomaticEnv()

	//
	// Optional config file.
	//

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	//
	// Config file is optional.
	//

	if err := viper.ReadInConfig(); err != nil {

		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf(
				"failed to read config file: %w",
				err,
			)
		}
	}

	//
	// Unmarshal configuration.
	//

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal configuration: %w",
			err,
		)
	}

	//
	// Validate required configuration.
	//

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Default Values
//

func setDefaults() {

	//
	// App
	//

	viper.SetDefault(
		"app.name",
		"ci-live",
	)

	viper.SetDefault(
		"app.environment",
		"development",
	)

	viper.SetDefault(
		"app.version",
		"1.0.0",
	)

	//
	// Server
	//

	viper.SetDefault(
		"server.host",
		"0.0.0.0",
	)

	viper.SetDefault(
		"server.port",
		8080,
	)

	viper.SetDefault(
		"server.read_timeout",
		15*time.Second,
	)

	viper.SetDefault(
		"server.write_timeout",
		15*time.Second,
	)

	viper.SetDefault(
		"server.idle_timeout",
		60*time.Second,
	)

	viper.SetDefault(
		"server.shutdown_timeout",
		10*time.Second,
	)

	//
	// PostgreSQL
	//

	viper.SetDefault(
		"postgres.host",
		"localhost",
	)

	viper.SetDefault(
		"postgres.port",
		5432,
	)

	viper.SetDefault(
		"postgres.user",
		"postgres",
	)

	viper.SetDefault(
		"postgres.password",
		"",
	)

	viper.SetDefault(
		"postgres.database",
		"ci_live",
	)

	viper.SetDefault(
		"postgres.max_conns",
		20,
	)

	viper.SetDefault(
		"postgres.min_conns",
		5,
	)

	viper.SetDefault(
		"postgres.max_conn_lifetime",
		30*time.Minute,
	)

	viper.SetDefault(
		"postgres.max_conn_idle_time",
		5*time.Minute,
	)

	//
	// Redis
	//

	viper.SetDefault(
		"redis.host",
		"localhost",
	)

	viper.SetDefault(
		"redis.port",
		6379,
	)

	viper.SetDefault(
		"redis.password",
		"",
	)

	viper.SetDefault(
		"redis.database",
		0,
	)

	viper.SetDefault(
		"redis.max_retries",
		3,
	)

	viper.SetDefault(
		"redis.pool_size",
		20,
	)

	viper.SetDefault(
		"redis.min_idle_conns",
		5,
	)

	viper.SetDefault(
		"redis.dial_timeout",
		5*time.Second,
	)

	viper.SetDefault(
		"redis.read_timeout",
		3*time.Second,
	)

	viper.SetDefault(
		"redis.write_timeout",
		3*time.Second,
	)

	viper.SetDefault(
		"redis.pool_timeout",
		4*time.Second,
	)

	//
	// JWT
	//

	viper.SetDefault(
		"jwt.private_key_path",
		"./keys/private.pem",
	)

	viper.SetDefault(
		"jwt.public_key_path",
		"./keys/public.pem",
	)

	viper.SetDefault(
		"jwt.access_token_duration",
		15*time.Minute,
	)

	viper.SetDefault(
		"jwt.refresh_token_duration",
		30*24*time.Hour,
	)

	viper.SetDefault(
		"jwt.issuer",
		"ci-live",
	)

	viper.SetDefault(
		"jwt.audience",
		"ci-live-api",
	)

	//
	// MediaMTX
	//

	viper.SetDefault(
		"mediamtx.base_url",
		"http://localhost:8889",
	)

	viper.SetDefault(
		"mediamtx.api_url",
		"http://localhost:9997",
	)

	viper.SetDefault(
		"mediamtx.rtmp_publish_url",
		"rtmp://localhost:1935",
	)

	viper.SetDefault(
		"mediamtx.hls_playback_url",
		"http://localhost:8888",
	)

	viper.SetDefault(
		"mediamtx.username",
		"",
	)

	viper.SetDefault(
		"mediamtx.password",
		"",
	)

	//
	// MinIO
	//

	viper.SetDefault(
		"minio.endpoint",
		"localhost:9000",
	)

	viper.SetDefault(
		"minio.access_key",
		"",
	)

	viper.SetDefault(
		"minio.secret_key",
		"",
	)

	viper.SetDefault(
		"minio.use_ssl",
		false,
	)

	viper.SetDefault(
		"minio.region",
		"us-east-1",
	)

	viper.SetDefault(
		"minio.bucket",
		"ci-live",
	)

	viper.SetDefault(
		"minio.public_base_url",
		"",
	)

	viper.SetDefault(
		"minio.presign_duration",
		15*time.Minute,
	)

	//
	// CDN
	//

	viper.SetDefault(
		"cdn.base_url",
		"",
	)
}

//
// Validate Configuration
//

func (c *Config) Validate() error {

	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf(
			"app name is required",
		)
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf(
			"server port must be between 1 and 65535",
		)
	}

	if strings.TrimSpace(c.Postgres.Host) == "" {
		return fmt.Errorf(
			"postgres host is required",
		)
	}

	if c.Postgres.Port < 1 || c.Postgres.Port > 65535 {
		return fmt.Errorf(
			"postgres port must be between 1 and 65535",
		)
	}

	if strings.TrimSpace(c.Postgres.Database) == "" {
		return fmt.Errorf(
			"postgres database is required",
		)
	}

	if strings.TrimSpace(c.Redis.Host) == "" {
		return fmt.Errorf(
			"redis host is required",
		)
	}

	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		return fmt.Errorf(
			"redis port must be between 1 and 65535",
		)
	}

	if c.JWT.AccessTokenDuration <= 0 {
		return fmt.Errorf(
			"jwt access token duration must be greater than zero",
		)
	}

	if c.JWT.RefreshTokenDuration <= 0 {
		return fmt.Errorf(
			"jwt refresh token duration must be greater than zero",
		)
	}

	return nil
}

//
// Server Address
//

func (c *Config) ServerAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		c.Server.Host,
		c.Server.Port,
	)
}
