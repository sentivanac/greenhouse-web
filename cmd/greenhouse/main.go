package main

import (
	"log"
	"net/http"
	"time"

	"greenhouse/internal/api"
	"greenhouse/internal/model"
	"greenhouse/internal/mqtt"
	"greenhouse/internal/storage"
)

func main() {
	store, err := storage.NewSQLite("greenhouse.db")
	if err != nil {
		log.Fatal(err)
	}

	api := api.New(store)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/api/latest", api.Latest)
	http.HandleFunc("/api/recent", api.Recent)
	http.HandleFunc("/api/range/combined", api.RangeCombined)

	go func() {
		log.Println("HTTP listening on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	mqtt.New(
		"tcp://broker.emqx.io:1883",
		"ghpep-backend",
		"ghpep/params",
		func(m model.Measurement) {
			m.Ts = time.Now().UnixMilli()
			if err := store.Insert(m); err != nil {
				log.Println("db insert failed:", err)
			}
		},
	)
	select {}
}
