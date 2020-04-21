package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const gitlabURL = "http://gitlab.alx/api/v4/projects/%d/merge_requests/%d/commits"

type commit struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
}

func getCommits(id, iid uint) ([]commit, error) {
	url := fmt.Sprintf(gitlabURL, id, iid)
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

func getTaskNumbers(commits []commit) ([]uint, error) {
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
