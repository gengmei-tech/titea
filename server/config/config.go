package config

import (
	"github.com/BurntSushi/toml"
	"github.com/gengmei-tech/titea/pkg/util/logutil"
	"github.com/pingcap/pd/pkg/metricutil"
	log "github.com/sirupsen/logrus"
)

var (
	defaultLogFormat       = "text"
	defaultLogLevel        = "info"
	defaultLogMaxSize uint = 300 // MB
	defaultMaxConn    uint = 5000
)

// Config file
type Config struct {
	Appname  string                  `toml:"appname"`
	Host     string                  `toml:"host"`
	Port     uint                    `toml:"port"`
	MaxConn  uint                    `toml:"max_connection"`
	Auth     string                  `toml:"auth"`
	Backend  BackendConfig           `toml:"backend"`
	Log      LogConfig               `toml:"log"`
	Metric   metricutil.MetricConfig `toml:"metric"`
	RunExpGc bool
}

// LogConfig log config file
type LogConfig struct {
	Level   string `toml:"level"`
	Format  string `toml:"format"`
	LogPath string `toml:"log-path"`
	MaxSize uint   `toml:"max-size"`
	MaxDays uint   `toml:"max-days"`
}

// BackendConfig pd config
type BackendConfig struct {
	PdAddr string `toml:"pd-address"`
}

var defaultConfig = Config{
	Appname:  "kv",
	Host:     "0.0.0.0",
	Port:     5379,
	MaxConn:  defaultMaxConn,
	Auth:     "",
	RunExpGc: false,
	Backend: BackendConfig{
		PdAddr: "",
	},
	Log: LogConfig{
		Level:   defaultLogLevel,
		Format:  defaultLogFormat,
		MaxSize: defaultLogMaxSize,
	},
}

// NewConfig init config
func NewConfig() *Config {
	conf := defaultConfig
	return &conf
}

//LoadConfig load config file from path
func (c *Config) LoadConfig(path string) (*Config, error) {
	if _, err := toml.DecodeFile(path, c); err != nil {
		log.Errorf("config file parse failed, %v", err)
		return nil, err
	}
	return c, nil
}

// ToLogConfig to log config
func (c *Config) ToLogConfig() *logutil.LogConfig {
	return &logutil.LogConfig{
		Level:  c.Log.Level,
		Format: c.Log.Format,
		File: logutil.FileLogConfig{
			Filename: c.Log.LogPath,
			MaxSize:  c.Log.MaxSize,
			MaxDays:  c.Log.MaxDays,
		},
	}
}
