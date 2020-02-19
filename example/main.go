package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IvanDmytrenko/rls"
	"github.com/go-redis/redis/v7"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Error handshaking redis server: %v", err)
		os.Exit(1)
	}

	sendMsglimiter := rls.NewRateLimiter(client, rls.Options{
		Limit:    1,
		Key:      "send_message",
		Duration: time.Second,
	})

	onDayDuration, err := time.ParseDuration("24h")
	if err != nil {
		log.Printf("Failed to parse duration: %v", err)
		os.Exit(1)
	}

	//registrationLimiter use for signup rate limiting
	registrationLimiter := rls.NewRateLimiter(client, rls.Options{
		Limit:    20,
		Key:      "signup",
		Duration: onDayDuration,
	})

	//wrongActionLimiter use for attempts rate limiting
	wrongActionLimiter := rls.NewRateLimiter(client, rls.Options{
		Limit:    3,
		Key:      "card_pay_attempt",
		Duration: onDayDuration,
	})

	_ = registrationLimiter
	_ = wrongActionLimiter

	http.HandleFunc("/sendMsg", RateLimitMiddleware(Signup, sendMsglimiter))
	http.ListenAndServe(":5555", nil)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test!")
}

func RateLimitMiddleware(h http.HandlerFunc, limitter *rls.RateLimiter) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//We can use IP or for example extracted actor
		allowed, err := limitter.Allow(r.RemoteAddr)
		if err != nil {
			log.Printf("Failed to check ratelimiting allowance: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}
