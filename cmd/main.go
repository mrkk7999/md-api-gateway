package main

import (
	"fmt"
	redis "md-api-gateway/caches/redis"
	"md-api-gateway/config"
	"md-api-gateway/proxy"
	"md-api-gateway/router"
	"md-api-gateway/utils/token"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {

	// Logger
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	err := godotenv.Load("../.env")
	if err != nil {
		log.Error("Error loading .env file")
	}

	var (
		redisUrl = os.Getenv("REDIS_URL")
	)

	cache, err := redis.NewRedis(redisUrl)
	if err != nil {
		log.Warn("Failed to connect to Redis")
	}

	config.LoadAuthConfig()

	token.InitJWKS()

	services := make(map[string]proxy.ServiceConfig)

	for key, val := range config.AuthSettings.Services {
		services[key] = proxy.ServiceConfig{
			Target: val.Target,
			Routes: val.Routes,
		}
	}

	proxy.LoadConfig(services)

	r := router.NewRouter(cache, log)

	var httpAddr = os.Getenv("HTTP_ADDR")

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	log.Info("Server is running " + httpAddr)

	go func() {
		server := &http.Server{
			Addr:    httpAddr,
			Handler: r,
		}
		errs <- server.ListenAndServe()
	}()

	log.Error("exit", <-errs)

}
