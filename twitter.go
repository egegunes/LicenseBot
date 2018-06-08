package main

import "os"
import "net/url"
import "github.com/ChimeraCoder/anaconda"

var TWITTER_CONSUMER_KEY string = os.Getenv("TWITTER_CONSUMER_KEY")
var TWITTER_CONSUMER_SECRET string = os.Getenv("TWITTER_CONSUMER_SECRET")
var TWITTER_ACCESS_TOKEN string = os.Getenv("TWITTER_ACCESS_TOKEN")
var TWITTER_ACCESS_TOKEN_SECRET string = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

func PushTweet(status string) {
	var v url.Values
	anaconda.SetConsumerKey(TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(TWITTER_ACCESS_TOKEN, TWITTER_ACCESS_TOKEN_SECRET)
	api.PostTweet(status, v)
}
