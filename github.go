package main

import "os"
import "encoding/json"
import "net/http"

type Commit struct {
	Sha     string `json:"sha,omitempty"`
	URL     string `json:"url,omitempty"`
	HtmlURL string `json:"html_url,omitempty"`
	Files   []File `json:"files,omitempty"`
}

type File struct {
	Filename string `json:"filename,omitempty"`
	Status   string `json:"status,omitempty"`
}

type Repository struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Stars int    `json:"stargazers_count,omitempty"`
}

type Event struct {
	Repository Repository `json:"repo,omitempty"`
	Payload    struct {
		Commits []Commit `json:"commits,omitempty"`
	} `json:"payload,omitempty"`
}

type RateLimit struct {
	Rate struct {
		Limit     int   `json:"limit,omitempty"`
		Remaining int   `json:"remaining,omitempty"`
		Reset     int64 `json:"reset,omitempty"`
	} `json:"rate,omitempty"`
}

var GITHUB_TOKEN string = os.Getenv("GITHUB_TOKEN")

func MakeRequest(url string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "token "+GITHUB_TOKEN)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	return resp, err
}

func CheckRateLimit() (int, int64) {
	resp, err := MakeRequest("https://api.github.com/rate_limit")

	if err != nil {
		panic(err)
	}

	var rateLimit RateLimit
	err = json.NewDecoder(resp.Body).Decode(&rateLimit)

	if err == nil {
		resp.Body.Close()
	}

	return rateLimit.Rate.Remaining, rateLimit.Rate.Reset
}

func FetchEvents() ([]Event, error) {
	resp, err := MakeRequest("https://api.github.com/events")

	if err != nil {
		panic(err)
	}

	var events []Event
	err = json.NewDecoder(resp.Body).Decode(&events)

	if err == nil {
		resp.Body.Close()
	}

	return events, err
}

func FetchCommit(url string) (Commit, error) {
	resp, err := MakeRequest(url)

	if err != nil {
		panic(err)
	}

	var commit Commit
	err = json.NewDecoder(resp.Body).Decode(&commit)

	if err == nil {
		resp.Body.Close()
	}

	return commit, err
}

func FetchStars(url string) (int, error) {
	resp, err := MakeRequest(url)

	if err != nil {
		panic(err)
	}

	var repo Repository
	err = json.NewDecoder(resp.Body).Decode(&repo)

	if err == nil {
		resp.Body.Close()
	}

	return repo.Stars, err
}
