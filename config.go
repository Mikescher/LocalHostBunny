package bunny

import (
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/confext"
	"os"
	"time"
)

const APILevel = 1

type Config struct {
	Namespace       string
	GinDebug        bool          `env:"GINDEBUG"`
	ReturnRawErrors bool          `env:"RETURNERRORS"`
	Custom404       bool          `env:"CUSTOM404"`
	LogLevel        zerolog.Level `env:"LOGLEVEL"`
	ServerIP        string        `env:"IP"`
	ServerPort      string        `env:"PORT"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT"`
	Cors            bool          `env:"CORS"`
}

type MailConfig struct {
	Host     string `env:"HOST"`
	Port     int    `env:"PORT"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	Sender   string `env:"SENDER"`
}

var Conf Config

var configLocHost = func() Config {
	return Config{
		Namespace:       "local",
		GinDebug:        true,
		ServerIP:        "0.0.0.0",
		ServerPort:      "80",
		Custom404:       true,
		ReturnRawErrors: true,
		RequestTimeout:  16 * time.Second,
		LogLevel:        zerolog.DebugLevel,
		Cors:            true,
	}
}

var configLocDocker = func() Config {
	return Config{
		Namespace:       "local-docker",
		GinDebug:        true,
		ServerIP:        "0.0.0.0",
		ServerPort:      "80",
		Custom404:       true,
		ReturnRawErrors: true,
		RequestTimeout:  16 * time.Second,
		LogLevel:        zerolog.DebugLevel,
		Cors:            true,
	}
}

var configDev = func() Config {
	return Config{
		Namespace:       "develop",
		GinDebug:        true,
		ServerIP:        "0.0.0.0",
		ServerPort:      "80",
		Custom404:       false,
		ReturnRawErrors: false,
		RequestTimeout:  16 * time.Second,
		LogLevel:        zerolog.DebugLevel,
		Cors:            false,
	}
}

var configStag = func() Config {
	return Config{
		Namespace:       "staging",
		GinDebug:        true,
		ServerIP:        "0.0.0.0",
		ServerPort:      "80",
		Custom404:       false,
		ReturnRawErrors: false,
		RequestTimeout:  16 * time.Second,
		LogLevel:        zerolog.DebugLevel,
		Cors:            false,
	}
}

var configProd = func() Config {
	return Config{
		Namespace:       "production",
		GinDebug:        false,
		ServerIP:        "0.0.0.0",
		ServerPort:      "80",
		Custom404:       false,
		ReturnRawErrors: false,
		RequestTimeout:  16 * time.Second,
		LogLevel:        zerolog.InfoLevel,
		Cors:            false,
	}
}

var allConfig = map[string]func() Config{
	"local-host":   configLocHost,
	"local-docker": configLocDocker,
	"develop":      configDev,
	"staging":      configStag,
	"production":   configProd,
}

var instID xid.ID

func InstanceID() string {
	return instID.String()
}

func getConfig(ns string) (Config, bool) {
	if ns == "" {
		ns = "local-host"
	}
	if cfn, ok := allConfig[ns]; ok {
		c := cfn()
		err := confext.ApplyEnvOverrides("BUNNY_", &c, "_")
		if err != nil {
			panic(err)
		}
		return c, true
	}
	return Config{}, false
}

func init() {
	instID = xid.New()

	ns := os.Getenv("CONF_NS")

	cfg, ok := getConfig(ns)
	if !ok {
		log.Fatal().Str("ns", ns).Msg("Unknown config-namespace")
	}

	Conf = cfg
}