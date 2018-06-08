package main

import "fmt"
import "regexp"
import "time"

func AnalyzeEvents() map[string]string {
	events, err := FetchEvents()

	if err != nil {
		panic(err)
	}

	var m map[string]string
	m = make(map[string]string)

	for _, event := range events {
		stars, err := FetchStars(event.Repository.URL)

		if err != nil {
			panic(err)
		}

		if stars > 100 {
			for _, commit := range event.Payload.Commits {
				commit, err := FetchCommit(commit.URL)

				if err != nil {
					panic(err)
				}

				for _, f := range commit.Files {
					matched, _ := regexp.MatchString("(?i)license", f.Filename)
					if matched {
						fmt.Sprintf("%s, license file %s", event.Repository.Name, f.Status)
						m[event.Repository.Name] = f.Status
					}
				}
			}
		}
	}

	return m
}

func main() {
	for {
		remaining, reset := CheckRateLimit()
		timeToReset := time.Unix(reset, 0).Sub(time.Now())
		if remaining > 0 {
			m := AnalyzeEvents()
			for k, _ := range m {
				PushTweet(fmt.Sprintf("HEADS UP! There is a license change in %s", k))
			}
		} else {
			fmt.Println(timeToReset, "left to rate limit reset")
			time.Sleep(timeToReset)
		}
	}
}
