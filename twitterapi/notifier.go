package twitterapi

import (
	"context"
	"strings"
	"github.com/pkg/errors"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/munoudesu/clipper/youtubedataapi"
	"github.com/munoudesu/clipper/database"
)

type User struct {
        Tags []string `toml: "tags"`
}

type Users map[string]*User

type Notifier struct {
	apiKeyAccessToken *ApiKeyAccessToken
	tweetLinkRoot     string
	tweetComment      string
	channels          youtubedataapi.Channels
	users             Users
	databaseOperator  *database.DatabaseOperator
	ctx               context.Context
	twitterClient     *twitter.Client
	verbose           bool
}

func (n *Notifier)Notify(renotify bool) (error) {
	for _, channel := range n.channels {
		channelPage, ok, err := n.databaseOperator.GetChannelPageByChannelId(channel.ChannelId)
		if err != nil {
			return errors.Wrapf(err, "can not get channel page from database (channelId = %v)", channel.ChannelId)
		}
		if !renotify && channelPage.Dirty == 0 {
			continue
		}
		tagText := ""
		user, ok := n.users[channel.Name]
		if ok {
			tagText = strings.Join(user.Tags, "\n")
		}
		tweetText := n.tweetComment + "\n" + tagText + "\n" + n.tweetLinkRoot + channel.ChannelId + ".html"
		_, res, err := n.twitterClient.Statuses.Update(tweetText, nil)
		if err != nil {
			return errors.Wrapf(err,"can not tweet (channelId = %v)", channel.ChannelId)
		}
		if res.StatusCode != 200 {
			return errors.Wrapf(err, "faild to tweet (status code = %v, channelId = %v)", res.StatusCode, channel.ChannelId)
		}
		err =  n.databaseOperator.UpdateDirtyOfChannelPage(channel.ChannelId, 0)
		if err != nil {
			return errors.Wrapf(err, "update dirty of channel page (channelId = %v)", channel.ChannelId)
		}
	}
	return nil
}

func NewNotifier(apiKeyAccessToken *ApiKeyAccessToken, tweetLinkRoot string, tweetComment string, channels youtubedataapi.Channels, users Users, databaseOperator *database.DatabaseOperator, verbose bool) (*Notifier) {
	config := oauth1.NewConfig(apiKeyAccessToken.ApiKey, apiKeyAccessToken.ApiSecretKey)
	token := oauth1.NewToken(apiKeyAccessToken.AccessToken, apiKeyAccessToken.AccessTokenSecret)
	ctx := context.Background()
	httpClient := config.Client(ctx, token)
	twitterClient := twitter.NewClient(httpClient)
	return &Notifier {
		apiKeyAccessToken: apiKeyAccessToken,
		tweetLinkRoot: tweetLinkRoot,
		tweetComment: tweetComment,
		channels: channels,
		users: users,
		databaseOperator: databaseOperator,
		ctx: ctx,
		twitterClient:  twitterClient,
		verbose: verbose,
	}
}