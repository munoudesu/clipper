package builder

import (
        "os"
        //"log"
        "github.com/pkg/errors"
        "github.com/munoudesu/clipper/youtubedataapi"
        "github.com/munoudesu/clipper/database"
)

type Builder struct {
	buildDirPath     string
	channels         youtubedataapi.Channels
	databaseOperator *database.DatabaseOperator
}

func (b *Builder) buildPage(channel *youtubedataapi.Channel) (string, error) {
	return "", nil
}

func (b *Builder) Build() (error) {
	for _, channel := range b.channels {
		lastChannelPage, ok, err := b.databaseOperator.GetChannelPageByChannelId(channel.ChannelId)
                if err != nil {
                        return  errors.Errorf("can not get channel page from database (channelId = %v)", channel.ChannelId)
                }
		newPageHash, err := b.buildPage(channel)
		if ok && lastChannelPage.PageHash == newPageHash {
			continue
		}
		err = b.databaseOperator.UpdatePageHashAndDirtyOfChannelPage(channel.ChannelId, newPageHash, 1)
                if err != nil {
                        return  errors.Errorf("can not update page hash and dirty of channelPage (channelId = %v, newPageHash = %v)", channel.ChannelId, newPageHash)
                }
	}
	return nil
}

func NewBuilder(buildDirPath string, channels youtubedataapi.Channels, databaseOperator *database.DatabaseOperator) (*Builder, error)  {
        if buildDirPath == "" {
                return nil, errors.New("no build directory path")
        }
        _, err := os.Stat(buildDirPath)
        if err != nil {
                err := os.MkdirAll(buildDirPath, 0755)
                if err != nil {
                        return nil, errors.Errorf("can not create directory (%v)", buildDirPath)
                }
        }
	return &Builder {
		buildDirPath: buildDirPath,
		channels: channels,
		databaseOperator: databaseOperator,
	}, nil
}

