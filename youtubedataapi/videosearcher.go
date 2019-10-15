package youtubedataapi


import (
	"log"
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
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

func (v *VideoSearcher)getCommentThreadByCommentThreadId(video *database.Video, commentThreadId string, etag string) (*database.CommentThread, bool, bool, error) {
        commentThreadsService := youtube.NewCommentThreadsService(v.youtubeService)
	commentThreadsListCall := commentThreadsService.List("id,replies,snippet")
	commentThreadsListCall.MaxResults(100)
	commentThreadsListCall.Id(commentThreadId)
	commentThreadsListCall.Order("time")
	commentThreadsListCall.PageToken("")
	commentThreadsListCall.TextFormat("plainText")
	commentThreadsListCall.IfNoneMatch(etag)
	commentThreadListResponse, err := commentThreadsListCall.Do()
	if err != nil {
		if googleapi.IsNotModified(err) {
			return nil, true, false, nil
		} else {
			return nil, false, false, errors.Wrapf(err, "can not do comment thread list call")
		}
	}
	if len(commentThreadListResponse.Items) != 1 {
		log.Printf("not found commentThread or found many commentThread (commentThreadId = %v): %v", commentThreadId, err)
		return nil, false, true, nil
	}
	item := commentThreadListResponse.Items[0]
	commentThread := &database.CommentThread {
		CommentThreadId: item.Id,
		Etag: item.Etag,
		Name: video.Name,
		ChannelId: item.Snippet.ChannelId,
		VideoId: item.Snippet.VideoId,
		TopLevelComment: &database.TopLevelComment {
			CommentId: item.Snippet.TopLevelComment.Id,
			Etag: item.Snippet.TopLevelComment.Etag,
			ChannelId: item.Snippet.ChannelId,
			VideoId: item.Snippet.TopLevelComment.Snippet.VideoId,
			CommentThreadId: commentThreadId,
			AuthorChannelUrl: item.Snippet.TopLevelComment.Snippet.AuthorChannelUrl,
			AuthorDisplayName: item.Snippet.TopLevelComment.Snippet.AuthorDisplayName,
			AuthorProfileImageUrl: item.Snippet.TopLevelComment.Snippet.AuthorProfileImageUrl,
			ModerationStatus: item.Snippet.TopLevelComment.Snippet.ModerationStatus,
			TextDisplay: item.Snippet.TopLevelComment.Snippet.TextDisplay,
			TextOriginal: item.Snippet.TopLevelComment.Snippet.TextOriginal,
			PublishAt: item.Snippet.TopLevelComment.Snippet.PublishedAt,
			UpdateAt: item.Snippet.TopLevelComment.Snippet.UpdatedAt,
		},
	}
	replyComments := make([]*database.ReplyComment, 0)
	if item.Replies != nil {
		for _, r := range item.Replies.Comments {
			replyComment := &database.ReplyComment {
				CommentId: r.Id,
				Etag: r.Etag,
				ChannelId: item.Snippet.ChannelId,
				VideoId: r.Snippet.VideoId,
				CommentThreadId: commentThreadId,
				ParentId: r.Snippet.ParentId,
				AuthorChannelUrl: r.Snippet.AuthorChannelUrl,
				AuthorDisplayName: r.Snippet.AuthorDisplayName,
				AuthorProfileImageUrl: r.Snippet.AuthorProfileImageUrl,
				ModerationStatus: r.Snippet.ModerationStatus,
				TextDisplay: r.Snippet.TextDisplay,
				TextOriginal: r.Snippet.TextOriginal,
				PublishAt: r.Snippet.PublishedAt,
				UpdateAt: r.Snippet.UpdatedAt,
			}
			replyComments = append(replyComments, replyComment)
		}
	}
	commentThread.ReplyComments = replyComments
	return commentThread, false, false, nil
}


func (v *VideoSearcher)searchCommentThreadsByVideo(video *database.Video, checkModified  bool) (error) {
        commentThreadsService := youtube.NewCommentThreadsService(v.youtubeService)
        pageToken := ""
        for {
                commentThreadsListCall := commentThreadsService.List("id")
                commentThreadsListCall.MaxResults(100)
                commentThreadsListCall.VideoId(video.VideoId)
                commentThreadsListCall.Order("time")
                commentThreadsListCall.PageToken(pageToken)
                commentThreadsListCall.TextFormat("plainText")
                commentThreadListResponse, err := commentThreadsListCall.Do()
                if err != nil {
			return  errors.Wrapf(err, "can not do comment thread list call")
                }
                for _, item := range commentThreadListResponse.Items {
			commentThread, ok, err := v.databaseOperator.GetCommentThreadByCommentThreadId(item.Id)
			if err != nil {
				return errors.Wrapf(err, "can not get commentThread by commentThreadId from database (commentThreadId = %v)", item.Id)
			}
			if ok {
				if !checkModified {
					log.Printf("skipped because video is already exists in database (commentThreadId = %v, etag = %v)", commentThread.CommentThreadId, commentThread.Etag)
					continue
				}
				// 更新チェックもする場合
				newCommentThread, notModified, notFound, err := v.getCommentThreadByCommentThreadId(video, commentThread.CommentThreadId, commentThread.Etag)
				if err != nil {
					return errors.Wrapf(err, "can not get commentThread by commentThreadId with api (commentThreadIdId = %v)", commentThread.CommentThreadId)
				}
				if notFound {
					log.Printf("skipped because not found commentThread resource (commentThreadIdId = %v)", commentThread.CommentThreadId)
					continue
				}
				if notModified || commentThread.Etag == newCommentThread.Etag {
					log.Printf("skipped because commentThread resource is not modified (commentThreadIdId = %v, etag = %v)", newCommentThread.CommentThreadId, newCommentThread.Etag)
					continue
				}
				err = v.databaseOperator.UpdateCommentThread(newCommentThread)
				if err != nil {
					return errors.Wrapf(err, "can not update commentThread (commentThreadId = %v, etag = %v)", newCommentThread.CommentThreadId, newCommentThread.Etag)
				}
			} else {
				// DB上にレコードがまだないので新規に情報を取得して追加
				newCommentThread, _, notFound,  err := v.getCommentThreadByCommentThreadId(video, item.Id, "")
				if err != nil {
					return errors.Wrapf(err, "can not get commentThread by commentThreadId with api (commentThreadIdId = %v)", item.Id)
				}
				if notFound {
					log.Printf("skipped because not found commentThread resource (commentThreadIdId = %v)", item.Id)
					continue
				}
				err = v.databaseOperator.UpdateCommentThread(newCommentThread)
				if err != nil {
					return errors.Wrapf(err, "can not update commentThread (commentThreadId = %v, etag = %v)", newCommentThread.CommentThreadId, newCommentThread.Etag)
				}
			}

                }
                if commentThreadListResponse.NextPageToken != "" {
                        pageToken = commentThreadListResponse.NextPageToken
                        //continue
			break
                }
                break
        }
	return nil
}

func (v *VideoSearcher)getVideoByVideoId(channel *Channel, videoId string, etag string) (*database.Video, bool, bool, error) {
	videoService :=	youtube.NewVideosService(v.youtubeService)
	videosListCall := videoService.List("id,snippet,player")
	videosListCall.Id(videoId)
	videosListCall.MaxResults(50)
	videosListCall.PageToken("")
	videoListResponse, err := videosListCall.Do()
	if err != nil {

		if googleapi.IsNotModified(err) {
			return nil, true, false, nil
		} else {
			return nil, false, false, errors.Wrapf(err, "can not get video by videoId with api (videoId = %v)", videoId)
		}
	}
	if len(videoListResponse.Items) != 1 {
		log.Printf("not found video or found many video (videoId = %v): %v", err)
		return nil, false, true, nil
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
		ThumbnailDefaultUrl: item.Snippet.Thumbnails.Default.Url,
		ThumbnailDefaultWidth: item.Snippet.Thumbnails.Default.Width,
		ThumbnailDefaultHeight: item.Snippet.Thumbnails.Default.Height,
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
	return video, false, false, nil
}

func (v *VideoSearcher)searchVideosByChannel(channel *Channel, checkModified bool) (error) {
	log.Printf("search video of channel %v", channel.ChannelId)
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
			video, ok, err := v.databaseOperator.GetVideoByVideoId(item.Id.VideoId)
			if err != nil {
				return errors.Wrapf(err, "can not get video by videoId from database (videoId = %v)", item.Id.VideoId)
			}
			if ok {
				if !checkModified {
					log.Printf("skipped because video is already exists in database (videoId = %v, etag = %v)", video.VideoId, video.Etag)
					continue
				}
				// 更新チェックもする場合
				newVideo, notModified, notFound, err := v.getVideoByVideoId(channel, video.VideoId, video.Etag)
				if err != nil {
					return errors.Wrapf(err, "can not get video by videoId with api (videoId = %v)", video.VideoId)
				}
				if notFound {
					log.Printf("skipped because not found video resource (videoId = %v)", video.VideoId)
					continue
				}
				if notModified || video.Etag == newVideo.Etag {
					log.Printf("skipped because video resource is not modified (videoId = %v, etag = %v)", newVideo.VideoId, newVideo.Etag)
					continue
				}
				err = v.databaseOperator.UpdateVideo(newVideo)
				if err != nil {
					return errors.Wrapf(err, "can not update video (videoId = %v, etag = %v)", newVideo.VideoId, newVideo.Etag)
				}
			} else {
				// DB上にレコードがまだないので新規に情報を取得して追加
				newVideo, _, notFound, err := v.getVideoByVideoId(channel, item.Id.VideoId, "")
				if err != nil {
					return errors.Wrapf(err, "can not get video by videoId with api (videoId = %v)", item.Id.VideoId)
				}
				if notFound {
					log.Printf("skipped because not found video resource (videoId = %v)", item.Id.VideoId)
					continue
				}
				err = v.databaseOperator.UpdateVideo(newVideo)
				if err != nil {
					return errors.Wrapf(err, "can not update video (videoId = %v, etag = %v)", newVideo.VideoId, newVideo.Etag)
				}
			}
                }
                if searchListResponse.NextPageToken != "" {
                        pageToken = searchListResponse.NextPageToken
			continue
                }
                break
        }
	return nil
}

func (v *VideoSearcher)Search(searchVideo bool, searchComment bool, checkVideoModified bool, checkCommentModified bool) (error) {
	if searchVideo {
		for _, channel := range v.channels {
			err := v.searchVideosByChannel(channel, checkVideoModified)
			if err != nil {
				return errors.Wrapf(err, "can not search videos by channel (name = %v, channelId = %v)", channel.Name, channel.ChannelId)
			}
		}
	}
	if searchComment {
		videos, err := v.databaseOperator.GetVideos()
		if err != nil {
			return errors.Wrapf(err, "can not get videos from database")
		}
		for _, video := range videos {
			err := v.searchCommentThreadsByVideo(video, checkCommentModified)
			if err != nil {
				return errors.Wrapf(err, "can not search comment threads by video (neme = %v, videoId = %v)", video.Name, video.VideoId)
			}
		}
	}
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
