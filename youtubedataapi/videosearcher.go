package youtubedataapi


import (
	"log"
	"context"
	"github.com/pkg/errors"
	//"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"github.com/munoudesu/clipper/database"
)

type Channel struct {
	Name      string `toml:"name"`
	ChannelId string `toml:"channelId"`
}

type VideoSearcher struct {
	apiKey           string
	channels         []*Channel
	ctx              context.Context
	youtubeService   *youtube.Service
	databaseOperator *database.DatabaseOperator
}


/*
func (v *VideoSearcher)searchCommentThreads(videoId string) (error) {
        commentThreadsService := youtube.NewCommentThreadsService(v.youtubeService)
        pageToken := ""
        for {
                commentThreadsListCall := commentThreadsService.List("id")
                commentThreadsListCall.MaxResults(100)
                commentThreadsListCall.VideoId(videoId)
                commentThreadsListCall.Order("time")
                commentThreadsListCall.PageToken(pageToken)
                commentThreadsListCall.TextFormat("plainText")
                commentThreadListResponse, err := commentThreadsListCall.Do()
                if err != nil {
                        fmt.Printf("faile in comment thread list call (%v)\n", err)
                        break
                }
                for _, item := range commentThreadListResponse.Items {
                        fmt.Printf("-----------------------------------------\n")
                        fmt.Printf("comment thread etag: %v\n", item.Etag)
                        fmt.Printf("comment thread id: %v\n", item.Id)
			ok, etag, err := databaseOperator.ExistsCommentThread(itemId)
			if ok && etag == item.Etag {
				continue
			}
			commentThread, err := v.getCommentThread(resourceId.VideoId)
			databaseOperator.updateCommentthread(comment)
			for _, replay := range commentthread.replay {
				databaseOperator.updateReply(video)
			}
                }
                if commentThreadListResponse.NextPageToken != "" {
                        pageToken = commentThreadListResponse.NextPageToken
                        continue
                }
                break
        }
	return nil
}
*/

func (v *VideoSearcher)getVideoByVideoId(channel *Channel, videoId string) (*database.Video, error) {
	videoService :=	youtube.NewVideosService(v.youtubeService)
	videosListCall := videoService.List("id,snippet,player")
	videosListCall.Id(videoId)
	videosListCall.MaxResults(50)
	videosListCall.PageToken("")
	videoListResponse, err := videosListCall.Do()
	if err != nil {
		return nil, errors.Wrapf(err, "can not get video by videoId with api (videoId = %v)", videoId)
	}
	if len(videoListResponse.Items) != 1 {
		return nil, errors.Wrapf(err, "not found video or found many video (videoId = %v)", videoId)
	}
	item := videoListResponse.Items[0]
	video := &database.Video{
		VideoId: item.Id,
		Etag: item.Etag,
		Name: channel.Name,
		ChannelId: item.Snippet.ChannelId,
		Title: item.Snippet.Title,
		Description: item.Snippet.Description,
		PublishdAt: item.Snippet.PublishedAt,
		ThumbnailHighUrl: item.Snippet.Thumbnails.High.Url,
		ThumbnailHighWidth: item.Snippet.Thumbnails.High.Width,
		ThumbnailHighHeight: item.Snippet.Thumbnails.High.Height,
		ThumbnailMediumUrl: item.Snippet.Thumbnails.Medium.Url,
		ThumbnailMediumWidth: item.Snippet.Thumbnails.Medium.Width,
		ThumbnailMediumHeight: item.Snippet.Thumbnails.Medium.Height,
		EmbedWidth: item.Player.EmbedWidth,
		EmbedHeight: item.Player.EmbedHeight,
		EmbedHtml: item.Player.EmbedHtml,
	}
	return video, nil
}

func (v *VideoSearcher)searchVideosByChannel(channel *Channel, checkModified bool) (error) {
        searchService := youtube.NewSearchService(v.youtubeService)
        pageToken := ""
        for {
                searchListCall := searchService.List("id")
                searchListCall.ChannelId(channel.ChannelId)
                searchListCall.EventType("completed")
                searchListCall.MaxResults(50)
                searchListCall.Order("date")
                searchListCall.PageToken(pageToken)
                searchListCall.SafeSearch("none")
                searchListCall.Type("video")
                searchListCall.VideoCaption("any")
                searchListCall.VideoDefinition("any")
                searchListCall.VideoDimension("any")
                searchListCall.VideoDuration("any")
                searchListCall.VideoEmbeddable("any")
                searchListCall.VideoLicense("any")
                searchListCall.VideoSyndicated("any")
                searchListCall.VideoType("any")
                searchListResponse, err := searchListCall.Do()
                if err != nil {
			return errors.Wrapf(err, "can not do search list call (channelId = %v)", channel.ChannelId)
                }
                for _, item := range searchListResponse.Items {
			video, ok, err := v.databaseOperator.GetVideo(item.Id.VideoId)
			if err != nil {
				return errors.Wrapf(err, "can not get video by videoId from database (channelId = %v)", item.Id.VideoId)
			}
			if ok {
				// 更新チェックもする場合
				if !checkModified {
					log.Printf("skipped because video is already exists in database. (videoId = %v, etag = %v)", video.VideoId, video.Etag)
					continue
				}
				newVideo, err := v.getVideoByVideoId(channel, video.VideoId)
				if err != nil {
					return errors.Wrapf(err, "can not get video by videoId with api (channelId = %v)", video.VideoId)
				}
				if video.Etag == newVideo.Etag {
					log.Printf("skipped because video resource is not modified. (videoId = %v, etag = %v)", newVideo.VideoId, newVideo.Etag)
					continue
				}
				err = v.databaseOperator.UpdateVideo(newVideo)
				if err != nil {
					return errors.Wrapf(err, "can not update video (videoId = %v, etag = %v)", newVideo.VideoId, newVideo.Etag)
				}
			} else {
				// DB上にレコードがまだないので新規に情報を取得して追加
				newVideo, err := v.getVideoByVideoId(channel, item.Id.VideoId)
				if err != nil {
					return errors.Wrapf(err, "can not get video by videoId with api (channelId = %v)", item.Id.VideoId)
				}
				err = v.databaseOperator.UpdateVideo(newVideo)
				if err != nil {
					return errors.Wrapf(err, "can not update video (videoId = %v, etag = %v)", video.VideoId, video.Etag)
				}
			}
                }
                if searchListResponse.NextPageToken != "" {
                        pageToken = searchListResponse.NextPageToken
			break
                        //continue
                }
                break
        }
	return nil
}

func (v *VideoSearcher)Search(checkModified bool) (error) {
	for _, channel := range v.channels {
		err := v.searchVideosByChannel(channel, checkModified)
		if err != nil {
			return errors.Wrapf(err, "can not search videos by channelId (name = %v, channelId = %v)", channel.Name, channel.ChannelId)
		}
	}

/*
	videos databaseOperator.getVideos()
	for _, video := range videos {
		err := v.searchCommentThreads(resourceId.VideoId)
		if err != nil {
			fmt.Printf("faile in get comment threads (%v)\n", err)
			break
		}
	}
*/


	return nil
}

func NewVideoSearcher(apiKey string, channels []*Channel, databaseOperator *database.DatabaseOperator) (*VideoSearcher, error) {
        ctx := context.Background()
        youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
        if err != nil {
		return nil, errors.Wrapf(err, "can not create youtube service")
	}
	return &VideoSearcher{
		channels: channels,
		ctx: ctx,
		youtubeService: youtubeService,
		databaseOperator: databaseOperator,
	}, nil
}
