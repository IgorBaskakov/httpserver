package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const redmineURL = "http://rm.pixel.alx/issues/%d.json"

func getTasks(ids []uint) ([]*task, error) {
	tasks := make([]*task, 0, len(ids))
	for _, id := range ids {
		t, err := getTask(id)
		if err != nil {
			return nil, err
		}

		if t == nil {
			continue
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

type task struct {
	Issue struct {
		Tracker struct {
			Name string `json:"name"`
		} `json:"tracker"`
		Assigned struct {
			Name string `json:"name"`
		} `json:"assigned_to"`
		Subject         string    `json:"subject"`
		Description     string    `json:"description"`
		TotalSpentHours float32   `json:"total_spent_hours"`
		Created         time.Time `json:"created_on"`
	} `json:"issue"`
}

func getTask(id uint) (*task, error) {
	url := fmt.Sprintf(redmineURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	key := os.Getenv("APIKEY")
	if key == "" {
		return nil, errors.New("$APIKEY must be set")
	}

	req.Header.Add("X-Redmine-API-Key", key)

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

	t := task{}
	json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}

	if t.Issue.Subject == "" {
		return nil, nil
	}

	return &t, nil
}
