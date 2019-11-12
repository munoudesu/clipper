package database

import (
        "os"
        "log"
        "path/filepath"
        "github.com/pkg/errors"
        "database/sql"
        _ "github.com/mattn/go-sqlite3"
)

type DatabaseOperator struct {
        databasePath string
        db           *sql.DB
	verbose      bool
}

type Channel struct{
	ChannelId               string
	Etag                    string
	Name                    string
	CustomUrl               string
	Title                   string
	Description             string
	PublishdAt              string
	ThumbnailDefaultUrl     string
	ThumbnailDefaultWidth   int64
	ThumbnailDefaultHeight  int64
	ThumbnailHighUrl        string
	ThumbnailHighWidth      int64
	ThumbnailHighHeight     int64
	ThumbnailMediumUrl      string
	ThumbnailMediumWidth    int64
	ThumbnailMediumHeight   int64
	ResponseEtag            string
}

type Video struct {
	VideoId                      string
	Etag                         string
	Name                         string
	ChannelId                    string
	ChannelTitle                 string
	Title                        string
	Description                  string
	PublishdAt                   string
	Duration                     string
        LiveStreamActiveLiveChatId   string
        LiveStreamActualStartTime    string
        LiveStreamActualEndTime      string
        LiveStreamScheduledStartTime string
        LiveStreamScheduledEndTime   string
	ThumbnailDefaultUrl          string
	ThumbnailDefaultWidth        int64
	ThumbnailDefaultHeight       int64
	ThumbnailHighUrl             string
	ThumbnailHighWidth           int64
	ThumbnailHighHeight          int64
	ThumbnailMediumUrl           string
	ThumbnailMediumWidth         int64
	ThumbnailMediumHeight        int64
	EmbedHeight                  int64
	EmbedWidth                   int64
	EmbedHtml                    string
        StatusPrivacyStatus          string
        StatusUploadStatus           string
        StatusEmbeddable             bool
	ResponseEtag                 string
}

type CommentThread struct {
	CommentThreadId string
	Etag            string
	Name            string
	ChannelId       string
	VideoId         string
	TopLevelComment *TopLevelComment
	ReplyComments   []*ReplyComment
	ResponseEtag    string
}

type TopLevelComment struct {
	CommentId             string
	Etag                  string
	ChannelId             string
	VideoId               string
	CommentThreadId       string
	AuthorChannelUrl      string
	AuthorDisplayName     string
	AuthorProfileImageUrl string
	ModerationStatus      string
	TextDisplay           string
	TextOriginal          string
	PublishAt             string
	UpdateAt              string
}

type ReplyComment struct {
	CommentId             string
	Etag                  string
	ChannelId             string
	VideoId               string
	CommentThreadId       string
	ParentId              string
	AuthorChannelUrl      string
	AuthorDisplayName     string
	AuthorProfileImageUrl string
	ModerationStatus      string
	TextDisplay           string
	TextOriginal          string
	PublishAt             string
	UpdateAt              string
}

type CommonComment ReplyComment

type LiveChatComment struct {
	UniqueId            string
	ChannelId           string
	VideoId             string
	ClientId            string
	MessageId           string
	TimestampAt         string
	TimestampText       string
	AuthorName          string
	AuthorPhotoUrl      string
	MessageText         string
	PurchaseAmountText  string
	VideoOffsetTimeMsec string
}

type ChannelPage struct {
	ChannelId  string
	Sha1Digest string
	Dirty      int64
	TweetId    int64
}

func (d *DatabaseOperator) DeleteLiveChatCommentsByVideoId(videoId string) (error) {
	res, err := d.db.Exec(`DELETE FROM liveChatComment WHERE videoId = ?`, videoId)
        if err != nil {
                return errors.Wrap(err, "can not delete liveChatComments")
        }
        // 削除処理の結果から削除されたレコード数を取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of liveChatComment")
        }
	if d.verbose {
		log.Printf("delete liveChatComments (videoId = %v, rowsAffected = %v)", videoId, rowsAffected)
	}

        return nil
}

func (d *DatabaseOperator) UpdateLiveChatComments(liveChatComments []*LiveChatComment) (error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "can not start transaction in update live chat")
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	for _, liveChatComment := range liveChatComments {
		res, err := tx.Exec(
		    `INSERT OR REPLACE INTO liveChatComment (
			uniqueId,
			channelId,
			videoId,
			clientId,
			messageId,
			timestampAt,
			timestampText,
			authorName,
			authorPhotoUrl,
			messageText,
			purchaseAmountText,
			videoOffsetTimeMsec
		    ) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?
		    )`,
		    liveChatComment.UniqueId,
		    liveChatComment.ChannelId,
		    liveChatComment.VideoId,
		    liveChatComment.ClientId,
		    liveChatComment.MessageId,
		    liveChatComment.TimestampAt,
		    liveChatComment.TimestampText,
		    liveChatComment.AuthorName,
		    liveChatComment.AuthorPhotoUrl,
		    liveChatComment.MessageText,
		    liveChatComment.PurchaseAmountText,
		    liveChatComment.VideoOffsetTimeMsec,
		)
		if err != nil {
		        tx.Rollback()
			return errors.Wrap(err, "can not insert liveChatComment")
		}
		// 挿入処理の結果からIDを取得
		id, err := res.LastInsertId()
		if err != nil {
		        tx.Rollback()
			return errors.Wrap(err, "can not get insert id of liveChatComment")
		}
		if d.verbose {
			log.Printf("update live chat comment (uniqueId = %v, insert id = %v)", liveChatComment.UniqueId, id)
		}
	}
	tx.Commit()
	return nil
}

func (d *DatabaseOperator) GetLiveChatCommentsByChannelId(channelId string) ([]*LiveChatComment, error) {
	liveChatComments := make([]*LiveChatComment, 0)
        liveChatCommentRows, err := d.db.Query(`SELECT * FROM liveChatComment Where channelId = ?`, channelId)
        if err != nil {
                return nil, errors.Wrap(err, "can not get liveChatComment by channelId from database")
        }
        defer liveChatCommentRows.Close()
        for liveChatCommentRows.Next() {
                liveChatComment := &LiveChatComment{}
                // カーソルから値を取得
                err := liveChatCommentRows.Scan(
		    &liveChatComment.UniqueId,
		    &liveChatComment.ChannelId,
		    &liveChatComment.VideoId,
		    &liveChatComment.ClientId,
		    &liveChatComment.MessageId,
		    &liveChatComment.TimestampAt,
		    &liveChatComment.TimestampText,
		    &liveChatComment.AuthorName,
		    &liveChatComment.AuthorPhotoUrl,
		    &liveChatComment.MessageText,
		    &liveChatComment.PurchaseAmountText,
		    &liveChatComment.VideoOffsetTimeMsec,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan liveChatComment by channelId from database")
                }
		liveChatComments = append(liveChatComments, liveChatComment)
        }
        return liveChatComments, nil
}

func (d *DatabaseOperator) GetLiveChatCommentsByVideoId(videoId string) ([]*LiveChatComment, error) {
	liveChatComments := make([]*LiveChatComment, 0)
        liveChatCommentRows, err := d.db.Query(`SELECT * FROM liveChatComment Where videoId = ?`, videoId)
        if err != nil {
                return nil, errors.Wrap(err, "can not get liveChatComment by videoId from database")
        }
        defer liveChatCommentRows.Close()
        for liveChatCommentRows.Next() {
                liveChatComment := &LiveChatComment{}
                // カーソルから値を取得
                err := liveChatCommentRows.Scan(
		    &liveChatComment.UniqueId,
		    &liveChatComment.ChannelId,
		    &liveChatComment.VideoId,
		    &liveChatComment.ClientId,
		    &liveChatComment.MessageId,
		    &liveChatComment.TimestampAt,
		    &liveChatComment.TimestampText,
		    &liveChatComment.AuthorName,
		    &liveChatComment.AuthorPhotoUrl,
		    &liveChatComment.MessageText,
		    &liveChatComment.PurchaseAmountText,
		    &liveChatComment.VideoOffsetTimeMsec,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan liveChatComment by videoId from database")
                }
		liveChatComments = append(liveChatComments, liveChatComment)
        }
        return liveChatComments, nil
}

func (d *DatabaseOperator) DeleteReplyCommentsByVideoId(videoId string) (error) {
	res, err := d.db.Exec(`DELETE FROM replyComment WHERE videoId = ?`, videoId)
        if err != nil {
                return errors.Wrap(err, "can not delete replyComments")
        }
        // 削除処理の結果から削除されたレコード数を取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of replyComment")
        }
	if d.verbose {
		log.Printf("delete replyComments (videoId = %v, rowsAffected = %v)", videoId, rowsAffected)
	}

        return nil
}

func (d *DatabaseOperator) DeleteTopLevelCommentsByVideoId(videoId string) (error) {
	res, err := d.db.Exec(`DELETE FROM topLevelComment WHERE videoId = ?`, videoId)
        if err != nil {
                return errors.Wrap(err, "can not delete topLevelComments")
        }
        // 削除処理の結果から削除されたレコード数を取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of topLevelComment")
        }
	if d.verbose {
		log.Printf("delete topLevelComments (videoId = %v, rowsAffected = %v)", videoId, rowsAffected)
	}

        return nil
}

func (d *DatabaseOperator) DeleteCommentThreadsByVideoId(videoId string) (error) {
	res, err := d.db.Exec(`DELETE FROM commentThread WHERE videoId = ?`, videoId)
        if err != nil {
                return errors.Wrap(err, "can not delete commentThreads")
        }
        // 削除処理の結果から削除されたレコード数を取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of commentThread")
        }
	if d.verbose {
		log.Printf("delete commentThreads (videoId = %v, rowsAffected = %v)", videoId, rowsAffected)
	}

        return nil
}

func (d *DatabaseOperator) updateReplyComments(tx *sql.Tx, replyComments []*ReplyComment) (error) {
	for _, replyComment := range replyComments {
		res, err := tx.Exec(
		    `INSERT OR REPLACE INTO replyComment (
			commentId,
			etag,
			channelId,
			videoId,
			commentThreadId,
			parentId,
			authorChannelUrl,
			authorDisplayName,
			authorProfileImageUrl,
			moderationStatus,
			textDisplay,
			textOriginal,
			publishAt,
			updateAt
		    ) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?
		    )`,
		    replyComment.CommentId,
		    replyComment.Etag,
		    replyComment.ChannelId,
		    replyComment.VideoId,
		    replyComment.CommentThreadId,
		    replyComment.ParentId,
		    replyComment.AuthorChannelUrl,
		    replyComment.AuthorDisplayName,
		    replyComment.AuthorProfileImageUrl,
		    replyComment.ModerationStatus,
		    replyComment.TextDisplay,
		    replyComment.TextOriginal,
		    replyComment.PublishAt,
		    replyComment.UpdateAt,
		)
		if err != nil {
			return errors.Wrap(err, "can not insert replyComment")
		}
		// 挿入処理の結果からIDを取得
		id, err := res.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "can not get insert id of replyComment")
		}
		if d.verbose {
			log.Printf("update reply comment (commentId = %v, insert id = %v)", replyComment.CommentId, id)
		}
	}
	return nil
}

func (d *DatabaseOperator) updateTopLevelComment(tx *sql.Tx, topLevelComment *TopLevelComment) (error) {
	res, err := tx.Exec(
            `INSERT OR REPLACE INTO topLevelComment (
                commentId,
                etag,
                channelId,
                videoId,
                commentThreadId,
                authorChannelUrl,
                authorDisplayName,
                authorProfileImageUrl,
                moderationStatus,
                textDisplay,
                textOriginal,
                publishAt,
                updateAt
            ) VALUES (
                ?, ?, ?, ?, ?,
                ?, ?, ?, ?, ?,
                ?, ?, ?
            )`,
	    topLevelComment.CommentId,
	    topLevelComment.Etag,
	    topLevelComment.ChannelId,
	    topLevelComment.VideoId,
	    topLevelComment.CommentThreadId,
	    topLevelComment.AuthorChannelUrl,
	    topLevelComment.AuthorDisplayName,
	    topLevelComment.AuthorProfileImageUrl,
	    topLevelComment.ModerationStatus,
	    topLevelComment.TextDisplay,
	    topLevelComment.TextOriginal,
	    topLevelComment.PublishAt,
	    topLevelComment.UpdateAt,
        )
        if err != nil {
                return errors.Wrap(err, "can not insert topLevelComment")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of topLevelComment")
        }
	if d.verbose {
		log.Printf("update top level comment (commentId = %v, insert id = %v)", topLevelComment.CommentId, id)
	}
        return nil
}

func (d *DatabaseOperator) UpdateCommentThread(commentThread *CommentThread) (error) {
        tx, err := d.db.Begin()
        if err != nil {
                return errors.Wrap(err, "can not start transaction in update reply comment")
        }
        defer func() {
                if p := recover(); p != nil {
                        tx.Rollback()
                        panic(p)
                }
        }()
	res, err := tx.Exec(
            `INSERT OR REPLACE INTO commentThread (
                commentThreadId,
                etag,
                name,
                channelId,
                videoId,
		responseEtag
            ) VALUES (
                ?, ?, ?, ?, ?, ?
            )`,
	    commentThread.CommentThreadId,
	    commentThread.Etag,
	    commentThread.Name,
	    commentThread.ChannelId,
	    commentThread.VideoId,
	    commentThread.ResponseEtag,
        )
        if err != nil {
                tx.Rollback()
                return errors.Wrap(err, "can not insert commentThread")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                tx.Rollback()
                return errors.Wrap(err, "can not get insert id of commentThread")
        }
	if d.verbose {
		log.Printf("update comment thread (commentThreadId = %v, insert id = %v)", commentThread.CommentThreadId, id)
	}
	err = d.updateTopLevelComment(tx, commentThread.TopLevelComment)
	if err != nil {
                tx.Rollback()
                return errors.Wrap(err, "can not update topLevelComment")
	}
	err = d.updateReplyComments(tx, commentThread.ReplyComments)
	if err != nil {
                tx.Rollback()
                return errors.Wrap(err, "can not update replayComments")
	}
        tx.Commit()
        return nil
}

func (d *DatabaseOperator) GetAllCommentsByChannelIdAndLikeText(channelId string, likeText string) ([]*CommonComment, error) {
	commonComments := make([]*CommonComment, 0)
        topLevelCommentRows, err := d.db.Query(`SELECT * FROM topLevelComment Where channelId = ? AND textOriginal like ?`, channelId, likeText)
        if err != nil {
                return nil, errors.Wrap(err, "can not get topLevelComment by channelId and likeText from database")
        }
        defer topLevelCommentRows.Close()
        for topLevelCommentRows.Next() {
                commonComment := &CommonComment{}
                // カーソルから値を取得
                err := topLevelCommentRows.Scan(
                    &commonComment.CommentId,
                    &commonComment.Etag,
                    &commonComment.ChannelId,
                    &commonComment.VideoId,
                    &commonComment.CommentThreadId,
                    &commonComment.AuthorChannelUrl,
                    &commonComment.AuthorDisplayName,
                    &commonComment.AuthorProfileImageUrl,
                    &commonComment.ModerationStatus,
                    &commonComment.TextDisplay,
                    &commonComment.TextOriginal,
                    &commonComment.PublishAt,
                    &commonComment.UpdateAt,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan topLevelComment by channelId and likeText from database")
                }
		commonComments = append(commonComments, commonComment)
        }
        replyCommentRows, err := d.db.Query(`SELECT * FROM replyComment Where channelId = ? AND textOriginal like ?`, channelId, likeText)
        if err != nil {
                return nil, errors.Wrap(err, "can not get replyComment by channelId and likeText from database")
        }
        defer replyCommentRows.Close()
        for replyCommentRows.Next() {
                commonComment := &CommonComment{}
                // カーソルから値を取得
                err := replyCommentRows.Scan(
                    &commonComment.CommentId,
                    &commonComment.Etag,
                    &commonComment.ChannelId,
                    &commonComment.VideoId,
                    &commonComment.CommentThreadId,
                    &commonComment.ParentId,
                    &commonComment.AuthorChannelUrl,
                    &commonComment.AuthorDisplayName,
                    &commonComment.AuthorProfileImageUrl,
                    &commonComment.ModerationStatus,
                    &commonComment.TextDisplay,
                    &commonComment.TextOriginal,
                    &commonComment.PublishAt,
                    &commonComment.UpdateAt,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan replyComment by channelId and likeText from database")
                }
		commonComments = append(commonComments, commonComment)
        }
        return commonComments, nil
}

func (d *DatabaseOperator) getReplyComments(commentThreadId string) ([]*ReplyComment, error) {
        rows, err := d.db.Query(`SELECT * FROM replyComment WHERE commentThreadId = ?`, commentThreadId)
        if err != nil {
                return nil, errors.Wrap(err, "can not get replyComment by commentThreadId from database")
        }
        defer rows.Close()
	replyComments := make([]*ReplyComment, 0)
        for rows.Next() {
                replyComment := &ReplyComment{}
                // カーソルから値を取得
                err := rows.Scan(
                    &replyComment.CommentId,
                    &replyComment.Etag,
                    &replyComment.ChannelId,
                    &replyComment.VideoId,
                    &replyComment.CommentThreadId,
                    &replyComment.ParentId,
                    &replyComment.AuthorChannelUrl,
                    &replyComment.AuthorDisplayName,
                    &replyComment.AuthorProfileImageUrl,
                    &replyComment.ModerationStatus,
                    &replyComment.TextDisplay,
                    &replyComment.TextOriginal,
                    &replyComment.PublishAt,
                    &replyComment.UpdateAt,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan replyComment by commentThreadId from database")
                }
		replyComments = append(replyComments, replyComment)
        }
        return replyComments, nil
}

func (d *DatabaseOperator) getTopLevelComment(commentThreadId string) (*TopLevelComment, bool, error) {
        rows, err := d.db.Query(`SELECT * FROM topLevelComment WHERE commentThreadId = ?`, commentThreadId)
        if err != nil {
                return nil, false, errors.Wrap(err, "can not get topLevelComment by commentThreadId from database")
        }
        defer rows.Close()
        for rows.Next() {
                topLevelComment := &TopLevelComment{}
                // カーソルから値を取得
                err := rows.Scan(
                    &topLevelComment.CommentId,
                    &topLevelComment.Etag,
                    &topLevelComment.ChannelId,
                    &topLevelComment.VideoId,
                    &topLevelComment.CommentThreadId,
                    &topLevelComment.AuthorChannelUrl,
                    &topLevelComment.AuthorDisplayName,
                    &topLevelComment.AuthorProfileImageUrl,
                    &topLevelComment.ModerationStatus,
                    &topLevelComment.TextDisplay,
                    &topLevelComment.TextOriginal,
                    &topLevelComment.PublishAt,
                    &topLevelComment.UpdateAt,
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan topLevelComment by commentThreadId from database")
                }
		return topLevelComment, true, nil
        }
        return nil, false, nil
}

func (d *DatabaseOperator) GetCommentThreadByCommentThreadId(commentThreadId string) (*CommentThread, bool, error) {
        rows, err := d.db.Query(`SELECT * FROM commentThread WHERE commentThreadId = ?`, commentThreadId)
        if err != nil {
                return nil, false, errors.Wrap(err, "can not get commentThread by commentThreadId from database")
        }
        defer rows.Close()
        for rows.Next() {
                commentThread := &CommentThread{}
                // カーソルから値を取得
                err := rows.Scan(
                    &commentThread.CommentThreadId,
                    &commentThread.Etag,
                    &commentThread.Name,
                    &commentThread.ChannelId,
                    &commentThread.VideoId,
                    &commentThread.ResponseEtag,
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan commentThread by commentThreadId from database")
                }
		topLevelComment, ok, err := d.getTopLevelComment(commentThread.CommentThreadId)
		if err != nil {
                        return nil, false, errors.Wrap(err, "can not get top level comment by commentThreadId from database")
		}
		if !ok {
                        return nil, false, errors.Wrap(err, "not found top level comment by commentThreadId from database")
		}
		replyComments, err := d.getReplyComments(commentThread.CommentThreadId)
		if err != nil {
                        return nil, false, errors.Wrap(err, "can not get reply comments by commentThreadId from database")
		}
		commentThread.TopLevelComment = topLevelComment
		commentThread.ReplyComments = replyComments
		return commentThread, true, nil
        }
        return nil, false, nil
}

func (d *DatabaseOperator) DeleteVideoByVideoId(videoId string) (error) {
	res, err := d.db.Exec(`DELETE FROM video WHERE videoId = ?`, videoId)
        if err != nil {
                return errors.Wrap(err, "can not delete video")
        }
        // 削除処理の結果から削除されたレコード数を取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of video")
        }
	if d.verbose {
		log.Printf("delete video (videoId = %v, rowsAffected = %v)", videoId, rowsAffected)
	}

        return nil
}

func (d *DatabaseOperator) UpdateVideo(video *Video) (error) {
	res, err := d.db.Exec(
            `INSERT OR REPLACE INTO video (
                videoId,
                etag,
                name,
                channelId,
                channelTitle,
                title,
                description,
                publishdAt,
                duration,
                liveStreamActiveLiveChatId,
                liveStreamActualStartTime,
                liveStreamActualEndTime,
                liveStreamScheduledStartTime,
                liveStreamScheduledEndTime,
                thumbnailDefaultUrl,
                thumbnailDefaultWidth,
                thumbnailDefaultHeight,
                thumbnailHighUrl,
                thumbnailHighWidth,
                thumbnailHighHeight,
                thumbnailMediumUrl,
                thumbnailMediumWidth,
                thumbnailMediumHeight,
                embedHeight,
                embedWidth,
                embedHtml,
                statusUploadStatus,
                statusEmbeddable,
		responseEtag
            ) VALUES (
                ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
                ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?
            )`,
	    video.VideoId,
	    video.Etag,
	    video.Name,
	    video.ChannelId,
	    video.ChannelTitle,
	    video.Title,
	    video.Description,
	    video.PublishdAt,
	    video.Duration,
            video.LiveStreamActiveLiveChatId,
            video.LiveStreamActualStartTime,
            video.LiveStreamActualEndTime,
            video.LiveStreamScheduledStartTime,
            video.LiveStreamScheduledEndTime,
	    video.ThumbnailDefaultUrl,
	    video.ThumbnailDefaultWidth,
	    video.ThumbnailDefaultHeight,
	    video.ThumbnailHighUrl,
	    video.ThumbnailHighWidth,
	    video.ThumbnailHighHeight,
	    video.ThumbnailMediumUrl,
	    video.ThumbnailMediumWidth,
	    video.ThumbnailMediumHeight,
	    video.EmbedHeight,
	    video.EmbedWidth,
	    video.EmbedHtml,
	    video.StatusPrivacyStatus,
	    video.StatusUploadStatus,
            video.StatusEmbeddable,
	    video.ResponseEtag,
        )
        if err != nil {
                return errors.Wrap(err, "can not insert video")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of video")
        }
	if d.verbose {
		log.Printf("update video (videoId = %v, insert id = %v)", video.VideoId, id)
	}

        return nil
}

func (d *DatabaseOperator) GetOldVideosByChannelIdAndOffset(channelId string, offset int64) ([]*Video, error) {
        rows, err := d.db.Query(`select * from video where channelId = ? order by publishdAt desc limit ?,(select count(videoId) from video where channelId = ?);`, channelId, offset, channelId)
        if err != nil {
                return nil, errors.Wrap(err, "can not get videos from database")
        }
        defer rows.Close()
	videos := make([]*Video, 0)
        for rows.Next() {
                video := &Video{}
                // カーソルから値を取得
                err := rows.Scan(
                    &video.VideoId,
                    &video.Etag,
                    &video.Name,
                    &video.ChannelId,
                    &video.ChannelTitle,
                    &video.Title,
                    &video.Description,
                    &video.PublishdAt,
                    &video.Duration,
                    &video.LiveStreamActiveLiveChatId,
                    &video.LiveStreamActualStartTime,
                    &video.LiveStreamActualEndTime,
                    &video.LiveStreamScheduledStartTime,
                    &video.LiveStreamScheduledEndTime,
                    &video.ThumbnailDefaultUrl,
                    &video.ThumbnailDefaultWidth,
                    &video.ThumbnailDefaultHeight,
                    &video.ThumbnailHighUrl,
                    &video.ThumbnailHighWidth,
                    &video.ThumbnailHighHeight,
                    &video.ThumbnailMediumUrl,
                    &video.ThumbnailMediumWidth,
                    &video.ThumbnailMediumHeight,
		    &video.EmbedHeight,
		    &video.EmbedWidth,
		    &video.EmbedHtml,
                    &video.StatusPrivacyStatus,
                    &video.StatusUploadStatus,
                    &video.StatusEmbeddable,
		    &video.ResponseEtag,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan videos from database")
                }
		videos = append(videos, video)
        }
        return videos, nil
}

func (d *DatabaseOperator) GetVideosByChannelId(channelId string) ([]*Video, error) {
        rows, err := d.db.Query(`SELECT * FROM video WHERE channelId = ?`, channelId)
        if err != nil {
                return nil, errors.Wrap(err, "can not get videos from database")
        }
        defer rows.Close()
	videos := make([]*Video, 0)
        for rows.Next() {
                video := &Video{}
                // カーソルから値を取得
                err := rows.Scan(
                    &video.VideoId,
                    &video.Etag,
                    &video.Name,
                    &video.ChannelId,
                    &video.ChannelTitle,
                    &video.Title,
                    &video.Description,
                    &video.PublishdAt,
                    &video.Duration,
                    &video.LiveStreamActiveLiveChatId,
                    &video.LiveStreamActualStartTime,
                    &video.LiveStreamActualEndTime,
                    &video.LiveStreamScheduledStartTime,
                    &video.LiveStreamScheduledEndTime,
                    &video.ThumbnailDefaultUrl,
                    &video.ThumbnailDefaultWidth,
                    &video.ThumbnailDefaultHeight,
                    &video.ThumbnailHighUrl,
                    &video.ThumbnailHighWidth,
                    &video.ThumbnailHighHeight,
                    &video.ThumbnailMediumUrl,
                    &video.ThumbnailMediumWidth,
                    &video.ThumbnailMediumHeight,
		    &video.EmbedHeight,
		    &video.EmbedWidth,
		    &video.EmbedHtml,
                    &video.StatusPrivacyStatus,
                    &video.StatusUploadStatus,
                    &video.StatusEmbeddable,
		    &video.ResponseEtag,
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan videos from database")
                }
		videos = append(videos, video)
        }
        return videos, nil
}

func (d *DatabaseOperator) GetVideoByVideoId(videoId string) (*Video, bool, error) {
        rows, err := d.db.Query(`SELECT * FROM video WHERE videoId = ?`, videoId)
        if err != nil {
                return nil, false, errors.Wrap(err, "can not get video by videoId from database")
        }
        defer rows.Close()
        for rows.Next() {
                video := &Video{}
                // カーソルから値を取得
                err := rows.Scan(
                    &video.VideoId,
                    &video.Etag,
                    &video.Name,
                    &video.ChannelId,
                    &video.ChannelTitle,
                    &video.Title,
                    &video.Description,
                    &video.PublishdAt,
                    &video.Duration,
                    &video.LiveStreamActiveLiveChatId,
                    &video.LiveStreamActualStartTime,
                    &video.LiveStreamActualEndTime,
                    &video.LiveStreamScheduledStartTime,
                    &video.LiveStreamScheduledEndTime,
                    &video.ThumbnailDefaultUrl,
                    &video.ThumbnailDefaultWidth,
                    &video.ThumbnailDefaultHeight,
                    &video.ThumbnailHighUrl,
                    &video.ThumbnailHighWidth,
                    &video.ThumbnailHighHeight,
                    &video.ThumbnailMediumUrl,
                    &video.ThumbnailMediumWidth,
                    &video.ThumbnailMediumHeight,
		    &video.EmbedHeight,
		    &video.EmbedWidth,
		    &video.EmbedHtml,
                    &video.StatusPrivacyStatus,
                    &video.StatusUploadStatus,
                    &video.StatusEmbeddable,
		    &video.ResponseEtag,
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan video by videoId from database")
                }
		return video, true, nil
        }
        return nil, false, nil
}

func (d *DatabaseOperator) UpdateChannel(channel *Channel) (error) {
	res, err := d.db.Exec(
            `INSERT OR REPLACE INTO channel (
                channelId,
                etag,
                name,
		customUrl,
                title,
                description,
                publishdAt,
                thumbnailDefaultUrl,
                thumbnailDefaultWidth,
                thumbnailDefaultHeight,
                thumbnailHighUrl,
                thumbnailHighWidth,
                thumbnailHighHeight,
                thumbnailMediumUrl,
                thumbnailMediumWidth,
                thumbnailMediumHeight,
		responseEtag
            ) VALUES (
                ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
                ?, ?, ?, ?, ?, ?, ?
            )`,
	    channel.ChannelId,
	    channel.Etag,
	    channel.Name,
	    channel.CustomUrl,
	    channel.Title,
	    channel.Description,
	    channel.PublishdAt,
	    channel.ThumbnailDefaultUrl,
	    channel.ThumbnailDefaultWidth,
	    channel.ThumbnailDefaultHeight,
	    channel.ThumbnailHighUrl,
	    channel.ThumbnailHighWidth,
	    channel.ThumbnailHighHeight,
	    channel.ThumbnailMediumUrl,
	    channel.ThumbnailMediumWidth,
	    channel.ThumbnailMediumHeight,
	    channel.ResponseEtag,
        )
        if err != nil {
                return errors.Wrap(err, "can not insert channel")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of channel")
        }
	if d.verbose {
		log.Printf("update channel (channelId = %v, insert id = %v)", channel.ChannelId, id)
	}

        return nil
}

func (d *DatabaseOperator) GetChannelByChannelId(channelId string) (*Channel, bool, error) {
        rows, err := d.db.Query(`SELECT * FROM channel WHERE channelId = ?`, channelId)
        if err != nil {
                return nil, false, errors.Wrap(err, "can not get videos from database")
        }
        defer rows.Close()
        for rows.Next() {
                channel := &Channel{}
                // カーソルから値を取得
                err := rows.Scan(
                    &channel.ChannelId,
                    &channel.Etag,
                    &channel.Name,
                    &channel.CustomUrl,
                    &channel.Title,
                    &channel.Description,
                    &channel.PublishdAt,
                    &channel.ThumbnailDefaultUrl,
                    &channel.ThumbnailDefaultWidth,
                    &channel.ThumbnailDefaultHeight,
                    &channel.ThumbnailHighUrl,
                    &channel.ThumbnailHighWidth,
                    &channel.ThumbnailHighHeight,
                    &channel.ThumbnailMediumUrl,
                    &channel.ThumbnailMediumWidth,
                    &channel.ThumbnailMediumHeight,
		    &channel.ResponseEtag,
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan channel from database")
                }
		return channel, true, nil
        }
        return nil, false, nil
}

func (d *DatabaseOperator) UpdateSha1DigestAndDirtyOfChannelPage(channelId string, pageHash string, dirty int64, tweetId int64) (error) {
	res, err := d.db.Exec(
            `INSERT OR REPLACE INTO channelPage (
                channelId,
                sha1Digest,
                dirty,
		tweetId
            ) VALUES (
                ?, ?, ?, ?
            )`,
	    channelId,
	    pageHash,
	    dirty,
	    tweetId,
        )
        if err != nil {
                return errors.Wrap(err, "can not insert channelPage")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of channelPage")
        }
	if d.verbose {
		log.Printf("update channel page (channelId = %v, insert id = %v)", channelId, id)
	}

	return nil
}

func (d *DatabaseOperator) UpdateDirtyAndTweetIdOfChannelPage(channelId string, dirty int64, tweetId int64) (error) {
	res, err := d.db.Exec( `UPDATE channelPage SET dirty = ?, tweetId = ? WHERE channelId = ?` , dirty, tweetId, channelId)
        if err != nil {
                return errors.Wrap(err, "can not update channelPage")
        }
        // 更新処理の結果からIDを取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of channelPage")
        }
	if d.verbose {
		log.Printf("update channel page (channelId = %v, rowsAffected = %v)", channelId, rowsAffected)
	}

	return nil
}

func (d *DatabaseOperator) GetChannelPageByChannelId(channelId string) (*ChannelPage, bool, error) {
        rows, err := d.db.Query(`SELECT * FROM channelPage WHERE channelId = ?`, channelId)
        if err != nil {
                return nil, false, errors.Wrap(err, "can not get channelPage by chanelId from database")
        }
        defer rows.Close()
        for rows.Next() {
                channelPage := &ChannelPage{}
                // カーソルから値を取得
                err := rows.Scan(
                    &channelPage.ChannelId,
                    &channelPage.Sha1Digest,
                    &channelPage.Dirty,
		    &channelPage.TweetId,
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan channelPage by channelId from database")
                }
		return channelPage, true, nil
        }
        return nil, false, nil

}

func (d *DatabaseOperator) createTables() (error) {
        channelTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS channel (
                channelId              TEXT PRIMARY KEY,
                etag                   TEXT NOT NULL,
                name                   TEXT NOT NULL,
                customUrl              TEXT NOT NULL,
                title                  TEXT NOT NULL,
                description            TEXT NOT NULL,
		publishdAt             TEXT NOT NULL,
		thumbnailDefaultUrl    TEXT NOT NULL,
		thumbnailDefaultWidth  INTEGER NOT NULL,
		thumbnailDefaultHeight INTEGER NOT NULL,
		thumbnailHighUrl       TEXT NOT NULL,
		thumbnailHighWidth     INTEGER NOT NULL,
		thumbnailHighHeight    INTEGER NOT NULL,
		thumbnailMediumUrl     TEXT NOT NULL,
		thumbnailMediumWidth   INTEGER NOT NULL,
		thumbnailMediumHeight  INTEGER NOT NULL,
		responseEtag           TEXT NOT NULL
	)`
	_, err := d.db.Exec(channelTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create channel table")
	}

        videoTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS video (
                videoId                      TEXT PRIMARY KEY,
                etag                         TEXT NOT NULL,
                name                         TEXT NOT NULL,
                channelId                    TEXT NOT NULL,
                channelTitle                 TEXT NOT NULL,
                title                        TEXT NOT NULL,
                description                  TEXT NOT NULL,
		publishdAt                   TEXT NOT NULL,
		duration                     TEXT NOT NULL,
                liveStreamActiveLiveChatId   TEXT NOT NULL,
                liveStreamActualStartTime    TEXT NOT NULL,
                liveStreamActualEndTime      TEXT NOT NULL,
                liveStreamScheduledStartTime TEXT NOT NULL, 
                liveStreamScheduledEndTime   TEXT NOT NULL,
		thumbnailDefaultUrl          TEXT NOT NULL,
		thumbnailDefaultWidth        INTEGER NOT NULL,
		thumbnailDefaultHeight       INTEGER NOT NULL,
		thumbnailHighUrl             TEXT NOT NULL,
		thumbnailHighWidth           INTEGER NOT NULL,
		thumbnailHighHeight          INTEGER NOT NULL,
		thumbnailMediumUrl           TEXT NOT NULL,
		thumbnailMediumWidth         INTEGER NOT NULL,
		thumbnailMediumHeight        INTEGER NOT NULL,
		embedHeight                  INTEGER NOT NULL,
		embedWidth                   INTEGER NOT NULL,
		embedHtml                    TEXT NOT NULL,
		statusPrivacyStatus          TEXT NOT NULL,
                statusUploadStatus           TEXT NOT NULL,
                statusEmbeddable             INTEGER NOT NULL,
		responseEtag                 TEXT NOT NULL
	)`
	_, err = d.db.Exec(videoTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create video table")
	}

        commentThreadTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS commentThread (
                commentThreadId       TEXT PRIMARY KEY,
                etag                  TEXT NOT NULL,
		name                  TEXT NOT NULL,
		channelId             TEXT NOT NULL,
		videoId               TEXT NOT NULL,
		responseEtag          TEXT NOT NULL
	)`
	_, err = d.db.Exec(commentThreadTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create commentThread table")
	}

        topLevelCommentTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS topLevelComment (
                commentId             TEXT PRIMARY KEY,
                etag                  TEXT NOT NULL,
		channelId             TEXT NOT NULL,
		videoId               TEXT NOT NULL,
		commentThreadId       TEXT NOT NULL,
		authorChannelUrl      TEXT NOT NULL,
		authorDisplayName     TEXT NOT NULL,
		authorProfileImageUrl TEXT NOT NULL,
		moderationStatus      TEXT NOT NULL,
		textDisplay           TEXT NOT NULL,
		textOriginal          TEXT NOT NULL,
		publishAt             TEXT NOT NULL,
		updateAt              TEXT NOT NULL
	)`
	_, err = d.db.Exec(topLevelCommentTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create topLevelComment table")
	}

        replyCommentTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS replyComment (
                commentId             TEXT PRIMARY KEY,
                etag                  TEXT NOT NULL,
		channelId             TEXT NOT NULL,
		videoId               TEXT NOT NULL,
		commentThreadId       TEXT NOT NULL,
		parentId              TEXT NOT NULL, 
		authorChannelUrl      TEXT NOT NULL,
		authorDisplayName     TEXT NOT NULL,
		authorProfileImageUrl TEXT NOT NULL,
		moderationStatus      TEXT NOT NULL,
		textDisplay           TEXT NOT NULL,
		textOriginal          TEXT NOT NULL,
		publishAt             TEXT NOT NULL,
		updateAt              TEXT NOT NULL
	)`
	_, err = d.db.Exec(replyCommentTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create replyComment table")
	}

        liveChatCommentTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS liveChatComment (
                uniqueId            TEXT PRIMARY KEY,
		channelId           TEXT NOT NULL,
		videoId             TEXT NOT NULL,
		clientId            TEXT NOT NULL,
		messageId           TEXT NOT NULL, 
		timestampAt         TEXT NOT NULL,
		timestampText       TEXT NOT NULL,
		authorName          TEXT NOT NULL,
		authorPhotoUrl      TEXT NOT NULL,
		messageText         TEXT NOT NULL,
		purchaseAmountText  TEXT NOT NULL,
		videoOffsetTimeMsec TEXT NOT NULL
	)`
	_, err = d.db.Exec(liveChatCommentTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create liveChatComment table")
	}
        liveChatCommentVideoIdIndexQuery := `CREATE INDEX IF NOT EXISTS liveChatComment_videoId_index ON liveChatComment(videoId)`
	_, err = d.db.Exec(liveChatCommentVideoIdIndexQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create liveChatComment_videoId_index index")
	}
        liveChatCommentChannelIdIndexQuery := `CREATE INDEX IF NOT EXISTS liveChatComment_channelId_index ON liveChatComment(channelId)`
	_, err = d.db.Exec(liveChatCommentChannelIdIndexQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create liveChatComment_channelId_index index")
	}

        channelPageTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS channelPage (
                channelId   TEXT PRIMARY KEY,
                sha1Digest  TEXT NOT NULL,
                dirty       INTEGER NOT NULL,
                tweetId     INTEGER NOT NULL
	)`
	_, err = d.db.Exec(channelPageTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create channelPage table")
	}

	return nil
}

func (d *DatabaseOperator) Open() (error) {
        db, err := sql.Open("sqlite3", d.databasePath)
        if err != nil {
                return errors.Wrapf(err, "can not open database")
        }
        d.db = db
        err = d.createTables()
        if err != nil {
                return errors.Wrapf(err, "can not create table of database")
        }
        return nil
}

func (d *DatabaseOperator) Close()  {
        d.db.Close()
}

func NewDatabaseOperator(databasePath string, verbose bool) (*DatabaseOperator, error) {
        if databasePath == "" {
                return nil, errors.New("no database path")
        }
        dirname := filepath.Dir(databasePath)
        _, err := os.Stat(dirname)
        if err != nil {
                err := os.MkdirAll(dirname, 0755)
                if err != nil {
                        return nil, errors.Errorf("can not create directory (%v)", dirname)
                }
        }
        return &DatabaseOperator{
                databasePath: databasePath,
                db: nil,
		verbose: verbose,
        }, nil
}
