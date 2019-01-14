package logutil

import (
	"errors"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultLogTimeFormat = "2006/01/02 15:04:05.000"
	// DefaultLogMaxSize is the default size of log files.
	DefaultLogMaxSize = 300 // MB
	defaultLogFormat  = "text"
	defaultLogLevel   = log.InfoLevel
)

// FileLogConfig serializes file log related config in toml/json.
type FileLogConfig struct {
	// Log filename, leave empty to disable file log.
	Filename string `toml:"filename" json:"filename"`
	// Is log rotate enabled. TODO.
	LogRotate bool `toml:"log-rotate" json:"log-rotate"`
	// Max size for a single file, in MB.
	MaxSize uint `toml:"max-size" json:"max-size"`
	// Max log keep days, default is never deleting.
	MaxDays uint `toml:"max-days" json:"max-days"`
	// Maximum number of old log files to retain.
	MaxBackups uint `toml:"max-backups" json:"max-backups"`
}

// LogConfig serializes log related config in toml/json.
type LogConfig struct {
	// Log level.
	Level string `toml:"level" json:"level"`
	// Log format. one of json, text, or console.
	Format string `toml:"format" json:"format"`
	// Disable automatic timestamps in output.
	DisableTimestamp bool `toml:"disable-timestamp" json:"disable-timestamp"`
	// File log config.
	File FileLogConfig `toml:"file" json:"file"`
}

func stringToLogLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "fatal":
		return log.FatalLevel
	case "error":
		return log.ErrorLevel
	case "warn", "warning":
		return log.WarnLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	}
	return defaultLogLevel
}

func stringToLogFormatter(format string, disableTimestamp bool) log.Formatter {
	switch strings.ToLower(format) {
	case "text":
		return &log.TextFormatter{
			DisableTimestamp: disableTimestamp,
		}
	case "json":
		return &log.JSONFormatter{
			TimestampFormat:  defaultLogTimeFormat,
			DisableTimestamp: disableTimestamp,
		}
	case "console":
		return &log.TextFormatter{
			FullTimestamp:    true,
			TimestampFormat:  defaultLogTimeFormat,
			DisableTimestamp: disableTimestamp,
		}
	case "highlight":
		return &log.TextFormatter{
			DisableTimestamp: disableTimestamp,
		}
	default:
		return &log.TextFormatter{}
	}
}

// initFileLog initializes file based logging options.
func initFileLog(cfg *FileLogConfig, logger *log.Logger) error {
	if st, err := os.Stat(cfg.Filename); err == nil {
		if st.IsDir() {
			return errors.New("can't use directory as log file name")
		}
	}
	if cfg.MaxSize == 0 {
		cfg.MaxSize = DefaultLogMaxSize
	}

	// use lumberjack to logrotate
	output := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    int(cfg.MaxSize),
		MaxBackups: int(cfg.MaxBackups),
		MaxAge:     int(cfg.MaxDays),
		LocalTime:  true,
	}

	if logger == nil {
		log.SetOutput(output)
	} else {
		logger.Out = output
	}
	return nil
}

// InitLogger initializes PD's logger.
func InitLogger(cfg *LogConfig) error {
	log.SetLevel(stringToLogLevel(cfg.Level))

	if cfg.Format == "" {
		cfg.Format = defaultLogFormat
	}
	formatter := stringToLogFormatter(cfg.Format, cfg.DisableTimestamp)
	log.SetFormatter(formatter)

	if len(cfg.File.Filename) != 0 {
		if err := initFileLog(&cfg.File, nil); err != nil {
			return err
		}
	}
	return nil
}
