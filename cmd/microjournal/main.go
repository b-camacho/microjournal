package main

import (
	"context"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/config"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/b-camacho/microjournal/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	conf := config.New()
	conf.Init()
	log.Println(conf)

	store := db.Init(conf.DbUri)
	authProvider := auth.Init(store, []byte("very secure"), []byte("much safety"))
	// This is the domain the server should accept connections for.
	handler := server.NewRouter(conf, authProvider, store)
	srv := &http.Server{
		Addr:         conf.Port,
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
