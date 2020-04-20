package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := chi.NewRouter()
	r.Get("/", rootHandler)

	r.Post("/", hookHandler)

	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("start http server at port %s...\n", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("server error listen: %s", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	srv.Shutdown(ctx)
	cancel()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("start default router")

	log.Printf("get request: %#v\n", r)
	log.Print("method:", r.Method)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "New changes coming to server!")
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("start hook router")

	log.Printf("post request: %#v\n", r)
	log.Print("method:", r.Method)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`{"result":"hook handler done"}`))
}
