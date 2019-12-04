package worker

import (
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/data"
)

var (
	logger         = log.New(os.Stdout, "worker: ", 0)
	consumerKey    = env.MustGetEnvVar("TW_KEY", "")
	consumerSecret = env.MustGetEnvVar("TW_SECRET", "")
)

func getClient(byUser *data.AuthedUser) (client *twitter.Client, err error) {

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(byUser.AccessTokenKey, byUser.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient), nil
}
