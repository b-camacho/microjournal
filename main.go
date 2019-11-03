package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/b-camacho/microjournal/server"
)

type Config struct {
	Urls []string
	Env string
	Port string
}

func ReadConfig() Config{
	configFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()
	bytes, _ := ioutil.ReadAll(configFile)
	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
	return config
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
