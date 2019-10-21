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

type Channels []*Channel

type Searcher struct {
	apiKeys            []string
	maxVideos          int64
	channels           Channels
	ctxs               []context.Context
	youtubeServices    []*youtube.Service
	youtubeServicesIdx int
	databaseOperator   *database.DatabaseOperator
}

func (s *Searcher)getYoutubeService() (*youtube.Service) {
	youtubeService := s.youtubeServices[s.youtubeServicesIdx]
	s.youtubeServicesIdx += 1
	if s.youtubeServicesIdx >= len(s.youtubeServices) {
		s.youtubeServicesIdx = 0
	}
	return youtubeService
}

func (s *Searcher)getCommentThreadByCommentThreadId(video *database.Video, commentThreadId string, etag string) (*database.CommentThread, bool, bool, error) {
        commentThreadsService := youtube.NewCommentThreadsService(s.getYoutubeService())
	commentThreadsListCall := commentThreadsService.List("id,replies,snippet")
	commentThreadsListCall.MaxResults(2)
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
		ChannelId: video.ChannelId,
		VideoId: item.Snippet.VideoId,
		ResponseEtag: commentThreadListResponse.Etag,
		TopLevelComment: &database.TopLevelComment {
			CommentId: item.Snippet.TopLevelComment.Id,
			Etag: item.Snippet.TopLevelComment.Etag,
			ChannelId: video.ChannelId,
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
				ChannelId: video.ChannelId,
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


func (s *Searcher)searchCommentThreadsByVideo(video *database.Video, checkModified  bool) (error) {
        commentThreadsService := youtube.NewCommentThreadsService(s.getYoutubeService())
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
			commentThread, ok, err := s.databaseOperator.GetCommentThreadByCommentThreadId(item.Id)
			if err != nil {
				return errors.Wrapf(err, "can not get commentThread by commentThreadId from database (commentThreadId = %v)", item.Id)
			}
			if ok {
				if !checkModified {
					log.Printf("skipped because commentThread is already exists in database (commentThreadId = %v)", commentThread.CommentThreadId)
					continue
				}
				newCommentThread, notModified, notFound, err := s.getCommentThreadByCommentThreadId(video, commentThread.CommentThreadId, commentThread.ResponseEtag)
				if err != nil {
					return errors.Wrapf(err, "can not get commentThread by commentThreadId with api (commentThreadIdId = %v)", commentThread.CommentThreadId)
				}
				if notFound {
					log.Printf("skipped because not found commentThread resource (commentThreadIdId = %v)", commentThread.CommentThreadId)
					continue
				}
				if notModified {
					log.Printf("skipped because commentThread resource is not modified (commentThreadIdId = %v, responseEtag = %v)", commentThread.CommentThreadId, commentThread.ResponseEtag)
					continue
				}
/*
				if commentThread.Etag == newCommentThread.Etag {
					log.Printf("skipped because commentThread resource have same etag (commentThreadIdId = %v, oldEtag = %v, newEtag = %v)", newCommentThread.CommentThreadId, commentThread.Etag, newCommentThread.Etag)
					continue
				}
*/
				err = s.databaseOperator.UpdateCommentThread(newCommentThread)
				if err != nil {
					return errors.Wrapf(err, "can not update commentThread (commentThreadId = %v)", newCommentThread.CommentThreadId)
				}
			} else {
				newCommentThread, _, notFound,  err := s.getCommentThreadByCommentThreadId(video, item.Id, "")
				if err != nil {
					return errors.Wrapf(err, "can not get commentThread by commentThreadId with api (commentThreadIdId = %v)", item.Id)
				}
				if notFound {
					log.Printf("skipped because not found commentThread resource (commentThreadIdId = %v)", item.Id)
					continue
				}
				err = s.databaseOperator.UpdateCommentThread(newCommentThread)
				if err != nil {
					return errors.Wrapf(err, "can not update commentThread (commentThreadId = %v)", newCommentThread.CommentThreadId)
				}
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

func (s *Searcher)getVideoByVideoId(channel *Channel, videoId string, etag string) (*database.Video, bool, bool, error) {
	videoService :=	youtube.NewVideosService(s.getYoutubeService())
	videosListCall := videoService.List("id,snippet,player,status")
	videosListCall.Id(videoId)
	videosListCall.MaxResults(2)
	videosListCall.PageToken("")
	videosListCall.IfNoneMatch(etag)
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
		StatusUploadStatus: item.Status.UploadStatus,
		StatusEmbeddable : item.Status.Embeddable,
		ResponseEtag: videoListResponse.Etag,
	}
	return video, false, false, nil
}

func (s *Searcher)searchVideosByChannel(channel *Channel, checkModified bool) (error) {
	// search new videos
        searchService := youtube.NewSearchService(s.getYoutubeService())
        pageToken := ""
        for loop := s.maxVideos; loop > 0; loop -= 50{
		maxResults := loop
		if maxResults > 50 {
			maxResults = 50
		}
                searchListCall := searchService.List("id")
                searchListCall.ChannelId(channel.ChannelId)
                searchListCall.EventType("completed")
                searchListCall.MaxResults(maxResults)
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
			video, ok, err := s.databaseOperator.GetVideoByVideoId(item.Id.VideoId)
			if err != nil {
				return errors.Wrapf(err, "can not get video by videoId from database (videoId = %v)", item.Id.VideoId)
			}
			if ok {
				if !checkModified {
					log.Printf("skipped because video is already exists in database (videoId = %v)", video.VideoId)
					continue
				}
				newVideo, notModified, notFound, err := s.getVideoByVideoId(channel, video.VideoId, video.ResponseEtag)
				if err != nil {
					return errors.Wrapf(err, "can not get video by videoId with api (videoId = %v)", video.VideoId)
				}
				if notFound {
					log.Printf("skipped because not found video resource (videoId = %v)", video.VideoId)
					continue
				}
				if notModified {
					log.Printf("skipped because video resource is not modified (videoId = %v, responseEtag = %v)", video.VideoId, video.ResponseEtag)
					continue
				}
				if video.Etag == newVideo.Etag {
					log.Printf("skipped because video resource have same etag (videoId = %v, oldEtag = %v, newEtag = %v)", newVideo.VideoId, video.Etag, newVideo.Etag,)
					continue
				}
				err = s.databaseOperator.UpdateVideo(newVideo)
				if err != nil {
					return errors.Wrapf(err, "can not update video (videoId = %v)", newVideo.VideoId)
				}
			} else {
				newVideo, _, notFound, err := s.getVideoByVideoId(channel, item.Id.VideoId, "")
				if err != nil {
					return errors.Wrapf(err, "can not get video by videoId with api (videoId = %v)", item.Id.VideoId)
				}
				if notFound {
					log.Printf("skipped because not found video resource (videoId = %v)", item.Id.VideoId)
					continue
				}
				err = s.databaseOperator.UpdateVideo(newVideo)
				if err != nil {
					return errors.Wrapf(err, "can not update video (videoId = %v)", newVideo.VideoId)
				}
			}
                }
                if searchListResponse.NextPageToken != "" {
                        pageToken = searchListResponse.NextPageToken
			continue
                }
                break
        }
	// delete old videos
	videos, err := s.databaseOperator.GetOldVideosByChannelIdAndOffset(channel.ChannelId, s.maxVideos)
	if err != nil {
		return errors.Wrapf(err, "can not get old videos (channelId = %v, maxVideos = %vv)", channel.ChannelId, s.maxVideos)
	}
	for _, video := range videos {
		err := s.databaseOperator.DeleteVideoByVideoId(video.VideoId)
		if err != nil {
			return errors.Wrapf(err, "can not delete old videos (videoId = %vv)", video.VideoId)
		}
		err = s.databaseOperator.DeleteCommentThreadsByVideoId(video.VideoId)
		if err != nil {
			return errors.Wrapf(err, "can not delete old comment threads (videoId = %vv)", video.VideoId)
		}
		err = s.databaseOperator.DeleteTopLevelCommentsByVideoId(video.VideoId)
		if err != nil {
			return errors.Wrapf(err, "can not delete old topLevelComments (videoId = %vv)", video.VideoId)
		}
		err = s.databaseOperator.DeleteReplyCommentsByVideoId(video.VideoId)
		if err != nil {
			return errors.Wrapf(err, "can not delete old replyComments (videoId = %vv)", video.VideoId)
		}
	}
	return nil
}

func (s *Searcher)getChannelByChannelId(name string, channelId string, etag string) (*database.Channel, bool, bool, error) {
	channelService := youtube.NewChannelsService(s.getYoutubeService())
	channelListCall := channelService.List("id,snippet")
	channelListCall.Id(channelId)
	channelListCall.MaxResults(2)
	channelListCall.PageToken("")
	channelListCall.IfNoneMatch(etag)
	channelListResponse, err := channelListCall.Do()
	if err != nil {
		if googleapi.IsNotModified(err) {
			return nil, true, false, nil
		} else {
			return nil, false, false, errors.Wrapf(err, "can not get channel by channelId with api (channelId = %v)", channelId)
		}
	}
	if len(channelListResponse.Items) != 1 {
		log.Printf("not found channel or found many channel (channelId = %v): %v", err)
		return nil, false, true, nil
	}
	item := channelListResponse.Items[0]
	channel := &database.Channel{
		ChannelId: item.Id,
		Etag: item.Etag,
		Name: name,
		CustomUrl: item.Snippet.CustomUrl,
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
		ResponseEtag: channelListResponse.Etag,
	}
	return channel, false, false, nil
}

func (s *Searcher)searchChannelByChannelId(name string, channelId string, checkModified bool) (error) {
	channel, ok, err := s.databaseOperator.GetChannelByChannelId(channelId)
	if err != nil {
		return errors.Wrapf(err, "can not get chennal by channelId from database (channelId = %v)", channelId)
	}
	if ok {
		if !checkModified {
			log.Printf("skipped because channel is already exists in database (channelId = %v)", channel.ChannelId)
			return nil
		}
		newChannel, notModified, notFound, err := s.getChannelByChannelId(name, channel.ChannelId, channel.ResponseEtag)
		if err != nil {
			return errors.Wrapf(err, "can not get channel by channelId with api (channelId = %v)", channel.ChannelId)
		}
		if notFound {
			log.Printf("skipped because not found channel resource (channelId = %v)", channel.ChannelId)
			return nil
		}
		if notModified {
			log.Printf("skipped because channel resource is not modified (channelId = %v, responseEtag = %v)", channel.ChannelId, channel.ResponseEtag)
			return nil
		}
		if channel.Etag == newChannel.Etag {
			log.Printf("skipped because channel resource have same etag (channelId = %v, oldEtag = %v, newEtag = %v)", newChannel.ChannelId, channel.Etag, newChannel.Etag,)
			return nil
		}
		err = s.databaseOperator.UpdateChannel(newChannel)
		if err != nil {
			return errors.Wrapf(err, "can not update channel (channelId = %v, etag = %v)", newChannel.ChannelId, newChannel.Etag)
		}
	} else {
		newChannel, _, notFound, err := s.getChannelByChannelId(name, channelId, "")
		if err != nil {
			return errors.Wrapf(err, "can not get channel by channelId with api (channelId = %v)", channelId)
		}
		if notFound {
			log.Printf("skipped because not found channel resource (channelId = %v)", channelId)
			return nil
		}
		err = s.databaseOperator.UpdateChannel(newChannel)
		if err != nil {
			return errors.Wrapf(err, "can not update channel (channelId = %v)", newChannel.ChannelId)
		}
	}
	return nil
}

func (s *Searcher)Search(searchChannel bool, searchVideo bool, searchComment bool, checkChannelModified bool, checkVideoModified bool, checkCommentModified bool) (error) {
	for _, channel := range s.channels {
		if searchChannel {
			err := s.searchChannelByChannelId(channel.Name, channel.ChannelId, checkChannelModified)
			if err != nil {
				return errors.Wrapf(err, "can not search channel by channelId (name = %v, channelId = %v)", channel.Name, channel.ChannelId)
			}
		}
		if searchVideo {
			err := s.searchVideosByChannel(channel, checkVideoModified)
			if err != nil {
				return errors.Wrapf(err, "can not search videos by channel (name = %v, channelId = %v)", channel.Name, channel.ChannelId)
			}
		}
		if searchComment {
			videos, err := s.databaseOperator.GetVideosByChannelId(channel.ChannelId)
			if err != nil {
				return errors.Wrapf(err, "can not get videos from database")
			}
			for _, video := range videos {
				if !video.StatusEmbeddable {
					log.Printf("skip get comment because unembeddable video (videoId = %v)", video.VideoId)
					continue
				}
				err := s.searchCommentThreadsByVideo(video, checkCommentModified)
				if err != nil {
					return errors.Wrapf(err, "can not search comment threads by video (neme = %v, videoId = %v)", video.Name, video.VideoId)
				}
			}
		}
	}
	return nil
}

func NewSearcher(apiKeys []string, maxVideos int64, channels []*Channel, databaseOperator *database.DatabaseOperator) (*Searcher, error) {
	ctxs := make([]context.Context, 0, len(apiKeys))
	youtubeServices := make([]*youtube.Service, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		ctx := context.Background()
		youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			return nil, errors.Wrapf(err, "can not create youtube service")
		}
		ctxs = append(ctxs, ctx)
		youtubeServices = append(youtubeServices, youtubeService)
	}
	return &Searcher{
		apiKeys: apiKeys,
		maxVideos: maxVideos,
		channels: channels,
		ctxs: ctxs,
		youtubeServices: youtubeServices,
		youtubeServicesIdx: 0,
		databaseOperator: databaseOperator,
	}, nil
}
