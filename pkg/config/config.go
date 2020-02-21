package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config is a struct that encapsulates configuration variables for our application.
type Config struct {
	Db struct {
		Name string `required:"true"`
		URI  string `required:"true"`
	}
	Server struct {
		Port          string `envconfig:"port" required:"true"`
		OriginAllowed string `default:"*" split_words:"true"`
		Domain        string `required:"true"`
	}
}

var conf *Config = &Config{}

func init() {
	godotenv.Load("../../.env")
	if err := envconfig.Process("", conf); err != nil {
		log.Fatal(err)
	}
}

// GetConf returns a configuration struct (used for setting up DB, server, etc.)
func GetConf() *Config {
	return conf
}
