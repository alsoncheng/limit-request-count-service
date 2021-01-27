package main

import (
	"strconv"
	"log"
	"net/http"
	"os"
	"time"
	"net"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	RedisClient *redis.Client
}


func (a *App) Initialize() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_URL", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	a.RedisClient = redisClient

	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/page", Page(a.RedisClient))
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func Page(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip1 := r.Header.Get("X-Forwarded-For")
		ip2, _, _ := net.SplitHostPort(r.RemoteAddr)

		var ip string
		if ip1 != "" {
			ip = ip1
		} else if ip2 != "" {
			ip = ip2
		} else {
			ip = "default"
		}

		incr := redisClient.Incr("prefix_" + ip)
		result := incr.Val()
		if result == int64(1) {
			redisClient.Expire("prefix_" + ip, 1 * time.Minute)
		}
		if result > int64(60) {
			w.Write([]byte("Error"))	
			return
		}
		
		w.Write([]byte(strconv.FormatInt(result, 10)))
	}
}