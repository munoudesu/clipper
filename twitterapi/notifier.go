package twitterapi

import (
	"log"
	"context"
	"strings"
	"github.com/pkg/errors"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/munoudesu/clipper/youtubedataapi"
	"github.com/munoudesu/clipper/database"
)

type User struct {
        Tags       []string `toml: "tags"`
	Comment    string   `toml: "comment"`
	SkipNotify bool     `toml: "skipNotify"`
}

type Users map[string]*User

type Notifier struct {
	apiKeyAccessToken *ApiKeyAccessToken
	tweetLinkRoot     string
	channels          youtubedataapi.Channels
	users             Users
	databaseOperator  *database.DatabaseOperator
	ctx               context.Context
	twitterClient     *twitter.Client
	verbose           bool
}

func (n *Notifier)Notify(renotify bool) (error) {
	// tweet of channel
	for _, channel := range n.channels {
		channelPage, ok, err := n.databaseOperator.GetChannelPageByChannelId(channel.ChannelId)
		if err != nil {
			return errors.Wrapf(err, "can not get channel page from database (channelId = %v)", channel.ChannelId)
		}
		if !renotify && channelPage.Dirty == 0 {
			continue
		}
		// 過去のtweetを消す
		if channelPage.TweetId != -1 {
			_, res, err := n.twitterClient.Statuses.Destroy(channelPage.TweetId, nil)
			if err != nil {
				if n.verbose {
					log.Printf("can not delete previous tweet (channelId = %v): %v", channel.ChannelId, err)
				}
			}
			if res.StatusCode != 200 {
				if n.verbose {
					log.Printf("delete tweet response error (status code = %v, channelId = %v)", res.StatusCode, channel.ChannelId)
				}
			}
			if n.verbose {
				log.Printf("delete tweet done (channelId = %v, tweetId = %v)", channel.ChannelId, channelPage.TweetId)
			}
		}
		// 新しいtweetをする
		tagText := ""
		user, ok := n.users[channel.Name]
		if !ok {
			return errors.Wrapf(err, "can not get user setting of twitter (channelId = %v, channelName = %v)", channel.ChannelId, channel.Name)
		}
		if user.SkipNotify {
			if n.verbose {
				log.Printf("skip notify (channelId = %v, channelName = %v, tweetId = %v)", channel.ChannelId, channel.Name, channelPage.TweetId)
			}
			err = n.databaseOperator.UpdateDirtyAndTweetIdOfChannelPage(channel.ChannelId, 0, -1)
			if err != nil {
				return errors.Wrapf(err, "update dirty of channel page (channelId = %v)", channel.ChannelId)
			}
			continue
		}
		tagText = strings.Join(user.Tags, "\n")
		tweetText := user.Comment + "\n" + tagText + "\n" + n.tweetLinkRoot + channel.ChannelId + ".html"
		tweet, res, err := n.twitterClient.Statuses.Update(tweetText, nil)
		if err != nil {
			return errors.Wrapf(err,"can not post new tweet (channelId = %v)", channel.ChannelId)
		}
		if res.StatusCode != 200 {
			return errors.Wrapf(err, "post tweet response error (status code = %v, channelId = %v)", res.StatusCode, channel.ChannelId)
		}
		err = n.databaseOperator.UpdateDirtyAndTweetIdOfChannelPage(channel.ChannelId, 0, tweet.ID)
		if err != nil {
			return errors.Wrapf(err, "update dirty of channel page (channelId = %v)", channel.ChannelId)
		}
	}
	// tweet of index
	channelPage, ok, err := n.databaseOperator.GetChannelPageByChannelId("index")
	if err != nil {
		return errors.Wrapf(err, "can not get channel page from database (channelId = index)")
	}
	if !renotify && channelPage.Dirty == 0 {
		return nil
	}
	// 過去のtweetを消す
	if channelPage.TweetId != -1 {
		_, res, err := n.twitterClient.Statuses.Destroy(channelPage.TweetId, nil)
		if err != nil {
			if n.verbose {
				log.Printf("can not delete previous tweet (channelId = index): %v", err)
			}
		}
		if res.StatusCode != 200 {
			if n.verbose {
				log.Printf("delete tweet response error (status code = %v, channelId = index)", res.StatusCode)
			}
		}
		if n.verbose {
			log.Printf("delete tweet done (channelId = index, tweetId = %v)", channelPage.TweetId)
		}
	}
	// 新しいtweetをする
	tagText := ""
	user, ok := n.users["index"]
	if !ok {
		return errors.Wrapf(err, "can not get user setting of twitter (channelId = index)")
	}
	if user.SkipNotify {
		if n.verbose {
			log.Printf("skip notify (channelId = index, tweetId = %v)", channelPage.TweetId)
		}
		err = n.databaseOperator.UpdateDirtyAndTweetIdOfChannelPage("index", 0, -1)
		if err != nil {
			return errors.Wrapf(err, "update dirty of channel page (channelId = index)")
		}
		return nil
	}
	tagText = strings.Join(user.Tags, "\n")
	tweetText := user.Comment + "\n" + tagText + "\n" + n.tweetLinkRoot + "index.html"
	tweet, res, err := n.twitterClient.Statuses.Update(tweetText, nil)
	if err != nil {
		return errors.Wrapf(err,"can not post new tweet (channelId = index)")
	}
	if res.StatusCode != 200 {
		return errors.Wrapf(err, "post tweet response error (status code = %v, channelId = index)", res.StatusCode)
	}
	err = n.databaseOperator.UpdateDirtyAndTweetIdOfChannelPage("index", 0, tweet.ID)
	if err != nil {
		return errors.Wrapf(err, "update dirty of channel page (channelId = index)")
	}
	return nil
}

func NewNotifier(apiKeyAccessToken *ApiKeyAccessToken, tweetLinkRoot string, channels youtubedataapi.Channels, users Users, databaseOperator *database.DatabaseOperator, verbose bool) (*Notifier) {
	config := oauth1.NewConfig(apiKeyAccessToken.ApiKey, apiKeyAccessToken.ApiSecretKey)
	token := oauth1.NewToken(apiKeyAccessToken.AccessToken, apiKeyAccessToken.AccessTokenSecret)
	ctx := context.Background()
	httpClient := config.Client(ctx, token)
	twitterClient := twitter.NewClient(httpClient)
	return &Notifier {
		apiKeyAccessToken: apiKeyAccessToken,
		tweetLinkRoot: tweetLinkRoot,
		channels: channels,
		users: users,
		databaseOperator: databaseOperator,
		ctx: ctx,
		twitterClient:  twitterClient,
		verbose: verbose,
	}
}
