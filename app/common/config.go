package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

type DBConfig struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     int
	Driver   string
	UseSSL   bool
}

type CookieConfig struct {
	Domain        string
	Secure        bool
	MaxAge        int
	SessionCookie string
}

type LoggingConfig struct {
	File  string
	Level logging.Level
}

type Config struct {
	AESKey  []byte
	Cookies *CookieConfig
	DB      *DBConfig
	Logging *LoggingConfig
}

var logLevels = map[string]logging.Level{
	"CRITICAL": logging.CRITICAL,
	"ERROR":    logging.ERROR,
	"WARNING":  logging.WARNING,
	"NOTICE":   logging.NOTICE,
	"INFO":     logging.INFO,
	"DEBUG":    logging.DEBUG,
}

// Returns a new config based on T2M_ENV
func NewConfig() *Config {
	config := loadConfig()
	config.expandPaths()
	config.makeDirs()
	return config
}

// This returns the default config directory and file.
// In most cases, that will be the .env file in the
// current working directory. When running automated tests,
// however, go changes into the subdirectories that contain
// the test files, so this resolves configDir to the project
// root directory.
func configDirAndFile() (configDir string, configFile string) {
	configDir, _ = os.Getwd()
	envName := os.Getenv("T2M_ENV")
	if envName == "" {
		PrintAndExit("Set T2M_ENV: dev, test, or production")
	}
	configFile = ".env"
	if envName != "" {
		configFile = ".env." + envName
	}
	if TestsAreRunning() {
		configDir = ProjectRoot()
	}
	return configDir, configFile
}

func loadConfig() *Config {
	configDir, configFile := configDirAndFile()
	v := viper.New()
	v.AddConfigPath(configDir)
	v.SetConfigName(configFile)
	v.SetConfigType("env")
	err := v.ReadInConfig()
	if err != nil {
		PrintAndExit(fmt.Sprintf("Fatal error config file: %v \n", err))
	}
	aesKey := v.GetString("AES_KEY")
	if len(aesKey) != 32 {
		PrintAndExit(fmt.Sprintf("Invalid AES Key."))
	}
	return &Config{
		AESKey: []byte(aesKey),
		Logging: &LoggingConfig{
			File:  v.GetString("LOG_FILE"),
			Level: getLogLevel(v.GetString("LOG_LEVEL")),
		},
		DB: &DBConfig{
			Host:     v.GetString("DB_HOST"),
			Name:     v.GetString("DB_NAME"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Port:     v.GetInt("DB_PORT"),
			Driver:   v.GetString("DB_DRIVER"),
			UseSSL:   v.GetBool("DB_USE_SSL"),
		},
		Cookies: {
			Domain:        v.GetString("COOKIE_DOMAIN"),
			Secure:        v.GetBool("SECURE_COOKIES"),
			MaxAge:        v.GetInt("SESSION_MAX_AGE"),
			SessionCookie: v.GetString("SESSION_COOKIE_NAME"),
		},
	}
}

func getLogLevel(level string) logging.Level {
	if level == "" {
		level = "INFO"
	}
	return logLevels[level]
}

// Expand ~ to home dir in path settings.
func (config *Config) expandPaths() {
	config.Logging.File = expandPath(config.Logging.File)
}

func expandPath(dirName string) string {
	dir, err := ExpandTilde(dirName)
	if err != nil {
		PrintAndExit(err.Error())
	}
	if dir == dirName && strings.HasPrefix(dirName, ".") {
		// dirName didn't change
		absPath, err := filepath.Abs(path.Join(ProjectRoot(), dirName))
		if err == nil && absPath != "" {
			dir = absPath
		}
	}
	return dir
}

func (config *Config) makeDirs() error {
	dirs := []string{
		config.Logging.File,
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err == nil || os.IsExist(err) {
			return nil
		} else {
			PrintAndExit(err.Error())
		}
	}
	return nil
}

// ToJSON serializes the config to JSON for logging purposes.
// It omits some sensitive data, such as the Pharos API key and
// AWS credentials.
func (config *Config) ToJSON() string {
	data, _ := json.Marshal(config)
	return string(data)
}
