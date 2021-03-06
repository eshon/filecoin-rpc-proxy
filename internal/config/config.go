package config

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type MethodType string
type CacheStorage string

const (
	// in seconds
	DefaultCacheCleanupInterval              = -1
	DefaultCacheExpiration                   = 0
	defaultLogLevel                          = "INFO"
	defaultPort                              = 8080
	defaultHost                              = "0.0.0.0"
	defaultJWTAlgorithm                      = "HS256"
	defaultSystemCachePeriod                 = 600
	defaultUserCachePeriod                   = 3600
	defaultRequestsBatchSize                 = 5
	defaultRequestsConcurrency               = 10
	defaultShutdownTimeout                   = 20
	CustomMethod                MethodType   = "custom"
	RegularMethod               MethodType   = "regular"
	MemoryCacheStorage          CacheStorage = "memory"
	RedisCacheStorage           CacheStorage = "redis"
	RedisPoolSize               int          = 10
)

var (
	defaultJWTPermissions = []string{"read"}
)

func (t MethodType) IsCustom() bool {
	return t == CustomMethod
}

func (t MethodType) IsRegular() bool {
	return t == RegularMethod
}

func (t MethodType) Valid() error {
	switch t {
	case CustomMethod, RegularMethod:
		return nil
	default:
		return fmt.Errorf("unknown method type: %s", t)
	}
}

func (c CacheStorage) IsMemory() bool {
	return c == MemoryCacheStorage
}

func (c CacheStorage) IsRedis() bool {
	return c == RedisCacheStorage
}

func (c CacheStorage) Valid() error {
	switch c {
	case MemoryCacheStorage, RedisCacheStorage:
		return nil
	default:
		return fmt.Errorf("unknown cache storage: %s", c)
	}
}

func (t *MethodType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var buf string
	if err := unmarshal(&buf); err != nil {
		return err
	}
	newT := MethodType(buf)
	if err := newT.Valid(); err != nil {
		return err
	}
	*t = newT
	return nil
}

func (t MethodType) MarshalYAML() (interface{}, error) {
	return string(t), nil
}

type CacheMethod struct {
	Name                string      `yaml:"name"`
	Enabled             bool        `yaml:"enabled,omitempty"`
	CacheByParams       bool        `yaml:"cache_by_params,omitempty"`
	NoStoreCache        bool        `yaml:"no_store_cache"`
	NoUpdateCache       bool        `yaml:"no_update_cache"`
	ParamsInCacheByID   []int       `yaml:"params_in_cache_by_id,omitempty"`
	ParamsInCacheByName []string    `yaml:"params_in_cache_by_name,omitempty"`
	Kind                *MethodType `yaml:"kind,omitempty"`
	ParamsForRequest    interface{} `yaml:"params_for_request,omitempty"`
}

func (c *CacheMethod) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type cacheMethod CacheMethod
	m := cacheMethod{
		Enabled: true,
	}
	if err := unmarshal(&m); err != nil {
		return err
	}
	*c = CacheMethod(m)
	return nil
}

type MemoryCacheSettings struct {
	DefaultExpiration int `yaml:"expiration,omitempty"`
	CleanupInterval   int `yaml:"cleanup_interval,omitempty"`
}

type RedisCacheSettings struct {
	URI      string `yaml:"uri,omitempty"`
	PoolSize int    `yaml:"pool_size,omitempty"`
}

type CacheSettings struct {
	Storage CacheStorage        `yaml:"storage,omitempty"`
	Memory  MemoryCacheSettings `yaml:"memory,omitempty"`
	Redis   RedisCacheSettings  `yaml:"redis,omitempty"`
}

type Config struct {
	CacheMethods            []CacheMethod `yaml:"cache_methods,omitempty"`
	JWTAlgorithm            string        `yaml:"jwt_alg"`
	JWTSecret               string        `yaml:"jwt_secret"`
	JWTSecretBase64         string        `yaml:"jwt_secret_base64"`
	JWTPermissions          []string      `json:"jwt_permissions"`
	Host                    string        `yaml:"host"`
	Port                    int           `yaml:"port"`
	UpdateCustomCachePeriod int           `yaml:"update_custom_cache_period"`
	UpdateUserCachePeriod   int           `yaml:"update_user_cache_period"`
	RequestsBatchSize       int           `yaml:"requests_batch_size"`
	RequestsConcurrency     int           `yaml:"requests_concurrency"`
	ShutdownTimeout         int           `yaml:"shutdown_timeout"`
	ProxyURL                string        `yaml:"proxy_url"`
	CacheSettings           CacheSettings `yaml:"cache_settings,omitempty"`
	LogLevel                string        `yaml:"log_level"`
	LogPrettyPrint          bool          `yaml:"log_pretty_print"`
	DebugHTTPRequest        bool          `yaml:"debug_http_request,omitempty"`
	DebugHTTPResponse       bool          `yaml:"debug_http_response,omitempty"`
}

func (c *Config) JWT() []byte {
	if c.JWTSecret != "" {
		return []byte(c.JWTSecret)
	}
	jwt, _ := base64.StdEncoding.DecodeString(c.JWTSecretBase64)
	return jwt
}

func New(reader io.Reader) (*Config, error) {
	c := &Config{}
	if err := yaml.NewDecoder(reader).Decode(c); err != nil {
		return nil, err
	}
	c.Init()
	return c, c.Validate()
}

func (c *Config) Init() {
	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.Host == "" {
		c.Host = defaultHost
	}
	if c.JWTAlgorithm == "" {
		c.JWTAlgorithm = defaultJWTAlgorithm
	}
	if c.UpdateCustomCachePeriod == 0 {
		c.UpdateCustomCachePeriod = defaultSystemCachePeriod
	}
	if c.UpdateUserCachePeriod == 0 {
		c.UpdateUserCachePeriod = defaultUserCachePeriod
	}
	if c.RequestsBatchSize == 0 {
		c.RequestsBatchSize = defaultRequestsBatchSize
	}
	if c.RequestsConcurrency == 0 {
		c.RequestsConcurrency = defaultRequestsConcurrency
	}
	if c.DebugHTTPRequest || c.DebugHTTPResponse {
		c.LogLevel = "DEBUG"
	}
	if len(c.JWTPermissions) == 0 {
		c.JWTPermissions = defaultJWTPermissions
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}
	if c.CacheSettings.Storage == "" {
		c.CacheSettings.Storage = MemoryCacheStorage
	}
	if c.CacheSettings.Redis.PoolSize == 0 {
		c.CacheSettings.Redis.PoolSize = RedisPoolSize
	}
	if c.CacheSettings.Memory.CleanupInterval == 0 {
		c.CacheSettings.Memory.CleanupInterval = DefaultCacheCleanupInterval
	}
	if c.CacheSettings.Memory.DefaultExpiration == 0 {
		c.CacheSettings.Memory.DefaultExpiration = DefaultCacheExpiration
	}
	for idx := range c.CacheMethods {
		method := c.CacheMethods[idx]
		if method.Kind == nil {
			if method.ParamsForRequest == nil {
				mt := RegularMethod
				method.Kind = &mt
			} else {
				mt := CustomMethod
				method.Kind = &mt
			}
			c.CacheMethods[idx] = method
		}
	}
}

func (c *Config) Validate() error {
	for _, method := range c.CacheMethods {
		if err := method.Kind.Valid(); err != nil {
			return err
		}
		if method.Kind.IsCustom() && method.ParamsForRequest == nil {
			return fmt.Errorf("custom method type should have been set with params_for_request")
		}
		if method.Kind.IsRegular() && method.ParamsForRequest != nil {
			return fmt.Errorf("regular method type should not have been set with params_for_request")
		}
	}
	if err := c.CacheSettings.Storage.Valid(); err != nil {
		return err
	}
	if c.CacheSettings.Storage.IsRedis() && c.CacheSettings.Redis.URI == "" {
		return fmt.Errorf("URI is required parameter for redis cache")
	}
	if c.JWTSecret == "" && c.JWTSecretBase64 == "" {
		return fmt.Errorf("jwt secret is mandatory parameter")
	}
	return nil
}

func FromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return New(file)
}
