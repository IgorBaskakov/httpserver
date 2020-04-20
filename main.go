package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	log.Print("method:", r.Method)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "New changes coming to server!")
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
	log.Print("start hook router")
	log.Print("method:", r.Method)

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
	fmt.Printf("hook body: %+v\n", hbody)
	fmt.Println(strings.Repeat("-", 100))

	var commits []commit
	if hbody.EventType == "merge_request" && hbody.ObjAttributes.State == "merged" {
		commits, err = getCommits(hbody.Project.ID, hbody.ObjAttributes.IID)
		if err != nil {
			log.Printf("get commits error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Printf("commits: %#v\n", commits)
	}

	w.Header().Add("Content-Type", "application/json")

	res := []byte("hook ok")
	w.WriteHeader(http.StatusOK)
	if len(commits) > 0 {
		res, err = json.Marshal(commits)
		if err != nil {
			log.Printf("marshal result data error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Write(res)
}

type commit struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
}

func getCommits(id, iid uint) ([]commit, error) {
	url := fmt.Sprintf("http://gitlab.alx/api/v4/projects/%d/merge_requests/%d/commits", id, iid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		return nil, errors.New("$TOKEN must be set")
	}

	req.Header.Add("Private-Token", token)

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	commits := []commit{}
	json.Unmarshal(body, &commits)
	if err != nil {
		return nil, err
	}

	return commits, nil
}
