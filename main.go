package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/google/go-github/v18/github"
	"golang.org/x/oauth2"
)

type licenseChange struct {
	Repo   *github.Repository
	Commit *github.RepositoryCommit
	File   *github.CommitFile
}

func tweetLicenseChange(t *anaconda.TwitterApi, c licenseChange) error {
	msg := ""
	fmt.Printf("%s\n", *c.File.Status)
	switch *c.File.Status {
	case "added":
		msg = fmt.Sprintf("HEADS UP! Someone added a license to %s! %s", *c.Repo.FullName, *c.Repo.HTMLURL)
	case "modified":
		msg = fmt.Sprintf("HEADS UP! Someone modified the license of %s! %s", *c.Repo.FullName, *c.Repo.HTMLURL)
	case "deleted":
		msg = fmt.Sprintf("HEADS UP! Someone deleted the license of %s! %s", *c.Repo.FullName, *c.Repo.HTMLURL)
	}
	fmt.Printf("%s\n", msg)
	_, err := t.PostTweet(msg, url.Values{})
	if err != nil {
		return fmt.Errorf("can't post tweet: %v", err)
	}
	fmt.Printf("tweet posted for %s\n", *c.Repo.FullName)
	return nil
}

func handleLicenseChange(ctx context.Context, g *github.Client, t *anaconda.TwitterApi, c licenseChange, errc chan error) {
	if *c.Repo.StargazersCount > -1 {
		fmt.Printf("%s (%d stars) %s %s", *c.Repo.FullName, *c.Repo.StargazersCount, *c.File.Status, *c.File.Filename)
		if err := tweetLicenseChange(t, c); err != nil {
			errc <- err
			return
		}
	}
}

func checkCommitsForLicense(ctx context.Context, g *github.Client, r *github.Repository, commits []github.PushEventCommit) (*github.RepositoryCommit, *github.Repository, *github.CommitFile, error) {
	repo, _, err := g.Repositories.GetByID(ctx, *r.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't get repository: %v", err)
	}

	for _, c := range commits {
		commit, _, err := g.Repositories.GetCommit(ctx, *repo.Owner.Login, *repo.Name, *c.SHA)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("can't get commit: %v", err)
		}
		for _, f := range commit.Files {
			matched := false
			for _, n := range []string{"LICENSE", "COPYING", "LICENSE.md"} {
				if n == *f.Filename {
					matched = true
				}
			}
			if matched {
				return commit, repo, &f, nil
			}
		}
	}

	return nil, nil, nil, nil
}

func handleEvent(ctx context.Context, g *github.Client, e *github.Event, r chan licenseChange, errc chan error) {
	if *e.Type == "PushEvent" {
		p, err := e.ParsePayload()
		if err != nil {
			errc <- fmt.Errorf("can't parse event payload: %v", err)
			return
		}

		payload, ok := p.(*github.PushEvent)
		if !ok {
			errc <- fmt.Errorf("can't get event payload")
			return
		}

		if *payload.Ref == "refs/heads/master" {
			commit, repo, file, err := checkCommitsForLicense(ctx, g, e.Repo, payload.Commits)
			if err != nil {
				errc <- fmt.Errorf("can't check commits for %s: %v", *e.Repo.Name, err)
				return
			}

			if commit != nil {
				r <- licenseChange{Repo: repo, Commit: commit, File: file}
			}
		}
	}
}

func run(ctx context.Context, g *github.Client, t *anaconda.TwitterApi) {
	r := make(chan licenseChange)
	errc := make(chan error)

	go func() {
		for change := range r {
			go handleLicenseChange(ctx, g, t, change, errc)
		}
	}()

	for {
		events, resp, err := g.Activity.ListEvents(ctx, nil)

		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				fmt.Fprintf(os.Stderr, "hit rate limit. sleeping till %v...\n", resp.Rate.Reset)
				time.Sleep(resp.Rate.Reset.Sub(time.Now()))
				fmt.Fprint(os.Stderr, "waking up...\n")
				continue
			} else {
				fmt.Fprintf(os.Stderr, "can't get events: %v\n", err)
				continue
			}
		}

		fmt.Printf("remaining: %d\n", resp.Rate.Remaining)

		for _, event := range events {
			go handleEvent(ctx, g, event, r, errc)
		}

		select {
		case err := <-errc:
			fmt.Fprintln(os.Stderr, err.Error())
		default:
			time.Sleep(time.Minute)
		}
	}
}

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	tc := oauth2.NewClient(ctx, ts)

	g := github.NewClient(tc)

	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))
	t := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

	run(ctx, g, t)
}
