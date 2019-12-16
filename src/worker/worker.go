package worker

import (
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
)

var (
	logger         = log.New(os.Stdout, "worker: ", 0)
	consumerKey    = env.MustGetEnvVar("TW_KEY", "")
	consumerSecret = env.MustGetEnvVar("TW_SECRET", "")
)

func getClient(byUser *store.AuthedUser) (client *twitter.Client, err error) {

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(byUser.AccessTokenKey, byUser.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient), nil
}

// returns items from b that are NOT in a
func getArrayDiff(a, b []int64) (diff []int64) {

	m := make(map[int64]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
