package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/b-camacho/microjournal/server"
)

type Config struct {
	Url string `env:"URL"`
	Environment string `env:"ENVIRONMENT"`
	Port string `env:"PORT"`
	DbUri string `env:"DBURI"`
}

func ReadConfig() Config{
	c := Config{}
	valC := reflect.ValueOf(&c).Elem()
	typeC := reflect.TypeOf(c)
	for i:=0; i< typeC.NumField(); i+=1 {
		configField := typeC.Field(i)
		configVal := valC.Field(i)
		envName := configField.Tag.Get("env")
		if v, ok :=  os.LookupEnv(envName); !ok {
			log.Fatalf("Required env var %s is not set", envName)
		} else {
			configVal.SetString(v)
		}
	}
	return c
}

func main() {
	config := ReadConfig()
	log.Println(config)
	// This is the domain the server should accept connections for.
	handler := server.NewRouter()
	srv := &http.Server{
		Addr:         config.Port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Start the server
	go func() {
		err := srv.ListenAndServe()
		//err := srv.Serve(autocert.NewListener(config.Urls...)) disable tls for now
		if err != nil {
			panic(err)
		}
	}()

	// Wait for an interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Attempt a graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
