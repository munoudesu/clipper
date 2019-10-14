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
	VideoId               string
	Etag                  string
	Name                  string
	ChannelId             string
	ChannelTitle          string
	Title                 string
	Description           string
	PublishdAt            string
	ThumbnailHighUrl      string
	ThumbnailHighWidth    int64
	ThumbnailHighHeight   int64
	ThumbnailMediumUrl    string
	ThumbnailMediumWidth  int64
	ThumbnailMediumHeight int64
	EmbedHeight           int64
	EmbedWidth            int64
	EmbedHtml             string
}

func (d *DatabaseOperator) GetVideo(videoId string) (*Video, bool, error) {
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
                ?, ?, ?, ?, ?, ?, ?
            )`,
	    video.VideoId,
	    video.Etag,
	    video.Name,
	    video.ChannelId,
	    video.ChannelTitle,
	    video.Title,
	    video.Description,
	    video.PublishdAt,
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
        log.Printf("insert id = %v", id)

        return nil
}

func (d *DatabaseOperator) createTables() (error) {
        videoTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS Video (
                videoId               TEXT PRIMARY KEY,
                etag                  TEXT NOT NULL,
                name                  TEXT NOT NULL,
                channelId             TEXT NOT NULL,
                channelTitle          TEXT,
                title                 TEXT,
                description           TEXT,
		publishdAt            TEXT,
		thumbnailHighUrl      TEXT,
		thumbnailHighWidth    INTEGER,
		thumbnailHighHeight   INTEGER,
		thumbnailMediumUrl    TEXT,
		thumbnailMediumWidth  INTEGER,
		thumbnailMediumHeight INTEGER,
		embedHeight           INTEGER,
		embedWidth            INTEGER,
		embedHtml             TEXT
	)`
	_, err := d.db.Exec(videoTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create video table")
	}

/*
        commentTableCreateQuery := `
            CREATE TABLE IF NOT EXISTS Video (
                id                    INTEGER PRIMARY KEY AUTOINCREMENT,
                commentId             TEXT NOT NULL UNIQUE,
                etag                  TEXT NOT NULL,
                name                  TEXT NOT NULL,
                channelId             TEXT NOT NULL,
                channelTitle          TEXT,
                title                 TEXT,
                description           TEXT,
		publishdAt            INTEGER,
		thumbnailHighUrl      TEXT,
		thumbnailHighWidth    TEXT,
		thumbnailHighHeight   TEXT,
		thumbnailMediumUrl    TEXT,
		thumbnailMediumWidth  TEXT,
		thumbnailMediumHeight TEXT,
		tweetId               TEXT                                 
	)`
	_, err := dbo.db.Exec(videoTableCreateQuery);
	if  err != nil {
		return errors.Wrap(err, "can not create video table")
	}
        // videoIdでvideoを検索するのに使う
        videoVideoIdIndexCreateQuery := `
            CREATE UNIQUE INDEX IF NOT EXISTS video_videoId_index ON video(videoId)`
        _, err = dbo.db.Exec(videoVideoIdIndexCreateQuery);
        if  err != nil {
                return errors.Wrap(err, "can not create video_videoId_index index")
        }
*/

	return nil
}

func (d *DatabaseOperator) Open() (error) {
        db, err := sql.Open("sqlite3", d.databasePath)
        if err != nil {
                return errors.Errorf("can not open database")
        }
        d.db = db
        err = d.createTables()
        if err != nil {
                return errors.Errorf("can not create table of database")
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
