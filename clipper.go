package main

import (
	"fmt"
	"context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func getCommentThreads(youtubeService *youtube.Service, videoId string) {
	commentThreadsService := youtube.NewCommentThreadsService(youtubeService)
	pageToken := ""
	for {
		commentThreadsListCall := commentThreadsService.List("replies,snippet")
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
		fmt.Printf("comment thread list response etag: %v\n", commentThreadListResponse.Etag)
		for _, item := range commentThreadListResponse.Items {
			itemSnippet := item.Snippet
			itemReplies := item.Replies
			fmt.Printf("-------------------------------------------\n")
			fmt.Printf("> top level comment author: %v\n", itemSnippet.TopLevelComment.Snippet.AuthorDisplayName)
			fmt.Printf("> top level comment author image url: %v\n", itemSnippet.TopLevelComment.Snippet.AuthorProfileImageUrl)
			fmt.Printf("> top level comment text: %v\n", itemSnippet.TopLevelComment.Snippet.TextDisplay)
			fmt.Printf("> top level comment publish at: %v\n", itemSnippet.TopLevelComment.Snippet.PublishedAt)
			fmt.Printf("> top level comment update at: %v\n", itemSnippet.TopLevelComment.Snippet.UpdatedAt)
			if itemReplies == nil || itemReplies.Comments == nil {
				continue
			}
			for _, reply := range itemReplies.Comments {
				fmt.Printf("------------------------------------------\n")
				fmt.Printf(">> reply comment author: %v\n",reply.Snippet.AuthorDisplayName)
				fmt.Printf(">> reply comment author image url: %v\n", reply.Snippet.AuthorProfileImageUrl)
				fmt.Printf(">> reply comment text: %v\n", reply.Snippet.TextDisplay)
				fmt.Printf(">> reply comment publish at: %v\n", reply.Snippet.PublishedAt)
				fmt.Printf(">> reply comment update at: %v\n", reply.Snippet.UpdatedAt)
			}
		}
		if commentThreadListResponse.NextPageToken != "" {
			pageToken = commentThreadListResponse.NextPageToken
			continue
		}
		break
	}
}

func getVideos(youtubeService *youtube.Service, channelId string) {
	searchService := youtube.NewSearchService(youtubeService)
	pageToken := ""
	for {
		searchListCall := searchService.List("snippet")
		searchListCall.ChannelId(channelId)
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
			fmt.Printf("faile in search list call (%v)\n", err)
			break
		}
		fmt.Printf("search list response etag: %v\n", searchListResponse.Etag)
		for _, item := range searchListResponse.Items {
			resourceId := item.Id
			itemSnippet := item.Snippet
			thumbnailDetails := itemSnippet.Thumbnails
			thumbnailHigh := thumbnailDetails.High
			thumbnailMedium := thumbnailDetails.Medium
			fmt.Printf("==================================================================\n")
			fmt.Printf("video id: %v\n", resourceId.VideoId)
			fmt.Printf("channel title: %v\n", itemSnippet.ChannelTitle)
			fmt.Printf("title: %v\n", itemSnippet.Title)
			fmt.Printf("description: %v\n", itemSnippet.Description)
			fmt.Printf("published at: %v\n", itemSnippet.PublishedAt)
			fmt.Printf("thumbnail high url: %v\n", thumbnailHigh.Url)
			fmt.Printf("thumbnail high width: %v\n", thumbnailHigh.Width)
			fmt.Printf("thumbnail high height: %v\n", thumbnailHigh.Height)
			fmt.Printf("thumbnail medium url: %v\n", thumbnailMedium.Url)
			fmt.Printf("thumbnail medium width: %v\n", thumbnailMedium.Width)
			fmt.Printf("thumbnail medium height: %v\n", thumbnailMedium.Height)
			fmt.Printf("video url: https://www.youtube.com/embed/%v?&loop=1&autoplay=1\n", resourceId.VideoId)
			getCommentThreads(youtubeService, resourceId.VideoId)
		}
		if searchListResponse.NextPageToken != "" {
			pageToken = searchListResponse.NextPageToken
			continue
		}
		break
	}
}

func main() {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey("key"))
	if err != nil {
		fmt.Printf("faile in create youtube service\n")
		return
	}
	getVideos(youtubeService, "UC1opHUrw8rvnsadT-iGp7Cg")
}
