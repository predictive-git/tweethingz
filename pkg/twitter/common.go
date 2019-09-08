package twitter

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mchmarny/tweethingz/pkg/data"
	"github.com/mchmarny/tweethingz/pkg/config"
)

var (
	logger = log.New(os.Stdout, "twitter: ", 0)
)


func getClient(byUser *data.AuthedUser) (client *twitter.Client, err error) {

	c, e := config.GetTwitterConfig()
	if e != nil {
		return nil, errors.Wrap(e, "Error getting Twitter config")
	}

	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(byUser.AccessTokenKey, byUser.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient), nil
}
