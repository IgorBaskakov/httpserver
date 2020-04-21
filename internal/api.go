package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
)

// StartServer start http server.
func StartServer(port string) {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("All OK!"))
	})
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

type hookBody struct {
	EventType string `json:"event_type"`
	Project   struct {
		ID uint `json:"id"`
	} `json:"project"`
	ObjAttributes struct {
		IID          uint   `json:"iid"`
		TargetBranch string `json:"target_branch"`
		State        string `json:"state"`
	} `json:"object_attributes"`
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read request body error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hbody := hookBody{}
	if err = json.Unmarshal(body, &hbody); err != nil {
		log.Printf("read request body error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go handleHookBody(hbody)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hook ok"))
}

func handleHookBody(hb hookBody) {
	var commits []commit
	var err error

	if hb.EventType == "merge_request" && hb.ObjAttributes.State == "merged" {
		commits, err = getCommits(hb.Project.ID, hb.ObjAttributes.IID)
		if err != nil {
			log.Printf("get commits error: %v", err)
			return
		}
	}

	ids, err := getTaskNumbers(commits)
	if err != nil {
		log.Printf("get number tasks error: %v", err)
		return
	}

	// ids = []uint{13380, 13362}
	tasks, err := getTasks(ids)
	if err != nil {
		log.Printf("get tasks error: %v", err)
		return
	}

	fmt.Println("len tasks:", len(tasks))
}
