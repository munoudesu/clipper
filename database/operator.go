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
}

type Video struct {
	VideoId                string
	Etag                   string
	Name                   string
	ChannelId              string
	ChannelTitle           string
	Title                  string
	Description            string
	PublishdAt             string
	ThumbnailDefaultUrl    string
	ThumbnailDefaultWidth  int64
	ThumbnailDefaultHeight int64
	ThumbnailHighUrl       string
	ThumbnailHighWidth     int64
	ThumbnailHighHeight    int64
	ThumbnailMediumUrl     string
	ThumbnailMediumWidth   int64
	ThumbnailMediumHeight  int64
	EmbedHeight            int64
	EmbedWidth             int64
	EmbedHtml              string
}

type CommentThread struct {
	CommentThreadId string
	Etag            string
	Name            string
	ChannelId       string
	VideoId         string
	TopLevelComment *TopLevelComment
	ReplyComments   []*ReplyComment
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

type ChannelPage struct {
	ChannelId string
	PageHash  string
	Dirty     int64
}

func (d *DatabaseOperator) updateReplyComments(replyComments []*ReplyComment) (error) {
	for _, replyComment := range replyComments {
		res, err := d.db.Exec(
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
		log.Printf("update reply comment (commentId = %v, insert id = %v)", replyComment.CommentId, id)
	}

	return nil
}

func (d *DatabaseOperator) updateTopLevelComment(topLevelComment *TopLevelComment) (error) {
	res, err := d.db.Exec(
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
	log.Printf("update top level comment (commentId = %v, insert id = %v)", topLevelComment.CommentId, id)

        return nil
}

func (d *DatabaseOperator) UpdateCommentThread(commentThread *CommentThread) (error) {
	res, err := d.db.Exec(
            `INSERT OR REPLACE INTO commentThread (
                commentThreadId,
                etag,
                name,
                channelId,
                videoId
            ) VALUES (
                ?, ?, ?, ?, ?
            )`,
	    commentThread.CommentThreadId,
	    commentThread.Etag,
	    commentThread.Name,
	    commentThread.ChannelId,
	    commentThread.VideoId,
        )
        if err != nil {
                return errors.Wrap(err, "can not insert commentThread")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of commentThread")
        }
	log.Printf("update comment thread (commentThreadId = %v, insert id = %v)", commentThread.CommentThreadId, id)
	err = d.updateTopLevelComment(commentThread.TopLevelComment)
	if err != nil {
                return errors.Wrap(err, "can not update topLevelComment")
	}
	err = d.updateReplyComments(commentThread.ReplyComments)
	if err != nil {
                return errors.Wrap(err, "can not update replayComments")
	}

        return nil
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
                )
                if err != nil {
                        return nil, errors.Wrap(err, "can not scan videos from database")
                }
		videos = append(videos, video)
        }
        return videos, nil
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
                embedHtml
            ) VALUES (
                ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
                ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
            )`,
	    video.VideoId,
	    video.Etag,
	    video.Name,
	    video.ChannelId,
	    video.ChannelTitle,
	    video.Title,
	    video.Description,
	    video.PublishdAt,
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
        )
        if err != nil {
                return errors.Wrap(err, "can not insert tradeContext")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of tradeContext")
        }
	log.Printf("update video (videoId = %v, insert id = %v)", video.VideoId, id)

        return nil
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
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan video by videoId from database")
                }
		return video, true, nil
        }
        return nil, false, nil
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
                    &channelPage.PageHash,
                    &channelPage.Dirty,
                )
                if err != nil {
                        return nil, false, errors.Wrap(err, "can not scan channelPage by channelId from database")
                }
		return channelPage, true, nil
        }
        return nil, false, nil

}

func (d *DatabaseOperator) UpdatePageHashAndDirtyOfChannelPage(channelId string, pageHash string, dirty int64) (error) {
	res, err := d.db.Exec(
            `INSERT OR REPLACE INTO channelPage (
                channelId,
                pageHash,
                dirty
            ) VALUES (
                ?, ?, ?
            )`,
	    channelId,
	    pageHash,
	    dirty,
        )
        if err != nil {
                return errors.Wrap(err, "can not insert channelPage")
        }
        // 挿入処理の結果からIDを取得
        id, err := res.LastInsertId()
        if err != nil {
                return errors.Wrap(err, "can not get insert id of channelPage")
        }
	log.Printf("update channel page (channelId = %v, insert id = %v)", channelId, id)

	return nil
}

func (d *DatabaseOperator) UpdateDirtyOfChannelPage(channelId string, dirty int64) (error) {
	res, err := d.db.Exec( `UPDATE channelPage SET dirty = ? WHERE channelId = ?` , dirty, channelId)
        if err != nil {
                return errors.Wrap(err, "can not update channelPage")
        }
        // 更新処理の結果からIDを取得
        rowsAffected, err := res.RowsAffected()
        if err != nil {
                return errors.Wrap(err, "can not get rowsAffected of channelPage")
        }
        log.Printf("update channel page (channelId = %v, rowsAffected = %v)", channelId, rowsAffected)

	return nil
}

func (d *DatabaseOperator) createTables() (error) {
        videoTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS Video (
                videoId                TEXT PRIMARY KEY,
                etag                   TEXT NOT NULL,
                name                   TEXT NOT NULL,
                channelId              TEXT NOT NULL,
                channelTitle           TEXT NOT NULL,
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
		embedHeight            INTEGER NOT NULL,
		embedWidth             INTEGER NOT NULL,
		embedHtml              TEXT NOT NULL
	)`
	_, err := d.db.Exec(videoTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create video table")
	}

        commentThreadTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS commentThread (
                commentThreadId       TEXT PRIMARY KEY,
                etag                  TEXT NOT NULL,
		name                  TEXT NOT NULL,
		channelId             TEXT NOT NULL,
		videoId               TEXT NOT NULL
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

        channelPageTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS channelPage (
                channelId TEXT PRIMARY KEY,
                pageHash  TEXT NOT NULL,
                dirty     INTEGER NOT NULL
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

func NewDatabaseOperator(databasePath string) (*DatabaseOperator, error) {
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
        }, nil
}
