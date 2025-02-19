package main

import (
	"fmt"
	"log"
	"md-api-gateway/config"
	"md-api-gateway/proxy"
	"md-api-gateway/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.LoadAuthConfig()
	services := make(map[string]proxy.ServiceConfig)
	for key, val := range config.AuthSettings.Services {
		services[key] = proxy.ServiceConfig{
			Target: val.Target,
			Routes: val.Routes,
		}
	}
	proxy.LoadConfig(services)

	r := router.NewRouter()

	var httpAddr = os.Getenv("HTTP_ADDR")

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println("Server is running " + httpAddr)

	go func() {
		server := &http.Server{
			Addr:    httpAddr,
			Handler: r,
		}
		errs <- server.ListenAndServe()
	}()

	log.Println("exit", <-errs)
}
