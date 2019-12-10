package config

import (
	"log"
	"os"
	"reflect"
)

type Config struct {
	Url            string `env:"URL"`
	Environment    string `env:"ENVIRONMENT"`
	Port           string `env:"PORT"`
	DbUri          string `env:"DBURI"`
	CookieBlockKey string `env:"CKBLOCKKEY"`
	CookieHashKey  string `env:"CKHASHKEY"`
}

func (c *Config) Init() {
	valC := reflect.ValueOf(c).Elem()
	typeC := reflect.TypeOf(*c)
	for i := 0; i < typeC.NumField(); i += 1 {
		configField := typeC.Field(i)
		configVal := valC.Field(i)
		envName := configField.Tag.Get("env")
		if v, ok := os.LookupEnv(envName); !ok {
			log.Fatalf("Required env var %s is not set", envName)
		} else {
			configVal.SetString(v)
		}
	}
}

func New() Config { // populates config from envvars
	c := Config{}
	return c

}
