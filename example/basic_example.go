package main

import (
	"database/sql"
	"github.com/go-redis/redis/v7"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"shutdown"
	"sync"
	"time"
)

var (
	db          *sql.DB
	redisClient *redis.Client
)

func init() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:         "localhost:6379",
			Password:     "", // no password set
			DB:           0,  // use default DB,
			MinIdleConns: 5,
			PoolSize:     10,
		})
		log.Println()
	}()
	go func() {
		var err error
		db, err = sql.Open("postgres", "postgres://postgres:@localhost:5432/postgres?sslmode=disable")
		log.Println()
		if err != nil {
			log.Fatal("postgres error", err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func router() http.Handler {
	return http.DefaultServeMux
}

func main() {

	srv := http.Server{
		Addr:    ":8080",
		Handler: router(),
	}

	shutdown.Manage(5*time.Second).
		Listener("server", srv.ListenAndServe, srv.Shutdown).
		PingCloseService("redis", func() error {
			if err := redisClient.Ping().Err(); err != nil {
				return err
			}
			log.Println("oi")
			return nil
		}, redisClient.Close).
		PingCloseService("postgres", db.Ping, db.Close).
		Listen()
}
