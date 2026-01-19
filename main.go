package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./"))))

	mux.HandleFunc("/healthz", handlerReadiness)

	fmt.Println("Serving on port: " + port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Error Server: %v", err)
	}
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
