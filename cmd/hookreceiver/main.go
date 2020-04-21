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
	"strconv"
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
	}

	w.Header().Add("Content-Type", "application/json")

	if len(commits) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hook ok"))
	}

	res, err := json.Marshal(commits)
	if err != nil {
		log.Printf("marshal result data error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("commits: %#v\n", commits)
	ids, err := getTasks(commits)
	if err != nil {
		log.Printf("get number tasks error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("ids:", ids)

	w.WriteHeader(http.StatusOK)
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

func getTasks(commits []commit) ([]uint, error) {
	uniq := make(map[uint]bool)
	tasks := make([]uint, 0, len(commits))

	for _, c := range commits {
		id, err := parseTitle(c.Title)
		if err != nil {
			return nil, err
		}

		if id == 0 {
			continue
		}

		if !uniq[id] {
			tasks = append(tasks, id)
		}

		uniq[id] = true
	}

	return tasks, nil
}

func parseTitle(title string) (uint, error) {
	if !strings.HasPrefix(title, "#") {
		return 0, nil
	}

	var num string
	for _, char := range title {
		if isNumber(char) {
			num += string(char)
		}
	}

	res, err := strconv.ParseUint(num, 10, 0)
	if err != nil {
		return 0, err
	}

	return uint(res), nil
}

func isNumber(char rune) bool {
	chars := [...]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

	for _, c := range chars {
		if char == c {
			return true
		}
	}

	return false
}

// tasks: 13380, 13362
