package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

const port = 8090

func main() {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("start default router")

		fmt.Printf("request: %#v\n", r)
		fmt.Println("url:", r.URL.String())

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "done")
	})

	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: r}
	go func() {
		fmt.Printf("start http server at port %d...\n", port)
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
