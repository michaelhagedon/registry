package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/stew/slice"
)

var allowedConfigs = []string{
	"ci",
	"dev",
	"docker",
	"integration",
	"production",
	"staging",
	"test",
	"travis",
	"uidemo",
}

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
	Secure        *securecookie.SecureCookie
	Domain        string
	HTTPSOnly     bool
	MaxAge        int
	SessionCookie string
	FlashCookie   string
	PrefsCookie   string
}

type LoggingConfig struct {
	File         string
	Level        zerolog.Level
	LogCaller    bool
	LogToConsole bool
	LogSql       bool
}

type TwoFactorConfig struct {
	AuthyEnabled  bool
	AuthyAPIKey   string `json:"-"`
	AWSRegion     string
	SMSEnabled    bool
	OTPExpiration time.Duration
}

type EmailConfig struct {
	AWSRegion   string
	Enabled     bool
	FromAddress string
	SesUser     string
	SesPassword string
}

type RedisConfig struct {
	URL       string
	Password  string
	DefaultDB int
}

type Config struct {
	Cookies   *CookieConfig
	DB        *DBConfig
	EnvName   string
	Logging   *LoggingConfig
	NsqUrl    string
	TwoFactor *TwoFactorConfig
	Email     *EmailConfig
	Redis     *RedisConfig
}

// Returns a new config based on APT_ENV
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
	envName := os.Getenv("APT_ENV")
	if !slice.Contains(allowedConfigs, envName) {
		PrintAndExit(fmt.Sprintf("Set APT_ENV to one of %s", strings.Join(allowedConfigs, ",")))
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
	v.AutomaticEnv() // override config file vars with ENV vars
	err := v.ReadInConfig()
	if err != nil {
		PrintAndExit(fmt.Sprintf("Fatal error config file: %v \n", err))
	}

	hashKey := []byte(v.GetString("COOKIE_HASH_KEY"))
	blockKey := []byte(v.GetString("COOKIE_BLOCK_KEY"))
	if len(hashKey) < 32 || len(blockKey) < 32 {
		PrintAndExit("COOKIE_HASH_KEY and COOKIE_BLOCK_KEY must each be >= 32 bytes")
	}
	var secureCookie = securecookie.New(hashKey, blockKey)

	nsqUrl := v.GetString("NSQ_URL")
	if !govalidator.IsURL(nsqUrl) {
		PrintAndExit("NSQ_URL is missing or invalid")
	}

	sesUser := v.GetString("AWS_SES_USER")
	sesPassword := v.GetString("AWS_SES_PWD")
	if sesUser == "" {
		fmt.Fprintln(os.Stderr, "AWS_SES_USER not set. Defaulting to AWS_ACCESS_KEY_ID for sending email and text messages.")
		sesUser = v.GetString("AWS_ACCESS_KEY_ID")
	}
	if sesPassword == "" {
		fmt.Fprintln(os.Stderr, "AWS_SES_PWD not set. Defaulting to AWS_SECRET_ACCESS_KEY for sending email and text messages.")
		sesPassword = v.GetString("AWS_SECRET_ACCESS_KEY")
	}

	return &Config{
		Logging: &LoggingConfig{
			File:         v.GetString("LOG_FILE"),
			Level:        getLogLevel(v.GetInt("LOG_LEVEL")),
			LogCaller:    v.GetBool("LOG_CALLER"),
			LogToConsole: v.GetBool("LOG_TO_CONSOLE"),
			LogSql:       v.GetBool("LOG_SQL"),
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
		EnvName: os.Getenv("APT_ENV"),
		Cookies: &CookieConfig{
			Secure:        secureCookie,
			Domain:        v.GetString("COOKIE_DOMAIN"),
			HTTPSOnly:     v.GetBool("HTTPS_COOKIES"),
			MaxAge:        v.GetInt("SESSION_MAX_AGE"),
			SessionCookie: v.GetString("SESSION_COOKIE_NAME"),
			FlashCookie:   v.GetString("FLASH_COOKIE_NAME"),
			PrefsCookie:   v.GetString("PREFS_COOKIE_NAME"),
		},
		NsqUrl: nsqUrl,
		TwoFactor: &TwoFactorConfig{
			AuthyAPIKey:   v.GetString("AUTHY_API_KEY"),
			AuthyEnabled:  v.GetBool("ENABLE_TWO_FACTOR_AUTHY"),
			AWSRegion:     v.GetString("AWS_REGION"),
			SMSEnabled:    v.GetBool("ENABLE_TWO_FACTOR_SMS"),
			OTPExpiration: v.GetDuration("OTP_EXPIRATION"),
		},
		Email: &EmailConfig{
			AWSRegion:   v.GetString("AWS_REGION"),
			Enabled:     v.GetBool("EMAIL_ENABLED"),
			FromAddress: v.GetString("EMAIL_FROM_ADDRESS"),
			SesUser:     sesUser,
			SesPassword: sesPassword,
		},
		Redis: &RedisConfig{
			DefaultDB: v.GetInt("REDIS_DEFAULT_DB"),
			Password:  v.GetString("REDIS_PASSWORD"),
			URL:       v.GetString("REDIS_URL"),
		},
	}
}

func getLogLevel(level int) zerolog.Level {
	return zerolog.Level(int8(level))
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
		path.Dir(config.Logging.File),
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

// BucketQualifier returns the S3 bucket qualifier for the current
// config. We could set this in the .env file, but we want to avoid
// the possibility of a config pointing to the wrong buckets. (For
// example, by someone carelessly copying and pasting config settings.)
// Our restrictive IAM permissions prevent the wrong environments
// from accessing the wrong buckets, but this is an extra layer of
// protection. This defaults to ".test", so if anything is misconfigured,
// we'll be reading from and writing to buckets in which we explicitly
// guarantee no permanance.
func (config *Config) BucketQualifier() string {
	if config.Cookies.Domain == "repo.aptrust.org" {
		return ""
	} else if config.Cookies.Domain == "staging.aptrust.org" {
		return ".staging"
	}
	return ".test"
}

// ToJSON serializes the config to JSON for logging purposes.
// It omits some sensitive data, such as the Pharos API key and
// AWS credentials.
func (config *Config) ToJSON() string {
	data, _ := json.Marshal(config)
	return string(data)
}

// Returns true if we're in a test or dev environment.
func (config *Config) IsTestOrDevEnv() bool {
	switch config.EnvName {
	case "dev", "test", "ci", "travis", "integration":
		return true
	}
	return false
}

// HTTPScheme returns "http" for the dev, test, ci, and travis
// environments. It returns "https" for all other environments.
func (config *Config) HTTPScheme() string {
	if config.IsTestOrDevEnv() {
		return "http"
	}
	return "https"
}
