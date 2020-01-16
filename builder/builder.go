package builder

import (
        "fmt"
        "math"
        "os"
        "log"
        "sort"
        "regexp"
        "strings"
        "strconv"
        "encoding/json"
        "html/template"
        "path/filepath"
        "crypto/sha1"
        "io/ioutil"
        "github.com/pkg/errors"
	copytool "github.com/otiai10/copy"
        "github.com/munoudesu/clipper/youtubedataapi"
        "github.com/munoudesu/clipper/database"
)

const timeRangeRegexpExpr = "[0-9]{1,2}:[0-9]{1,2}(:[0-9]{1,2})?(([ 　]*[~-→～－―][ 　]*[0-9]{1,2}:[0-9]{1,2}(:[0-9]{1,2})?)|([ 　]*@[ 　]*([0-9]+[hH])?([0-9]+[mM])?([0-9]+[sS])?))?"
const startEndSepRegexExpr = "[~-→～－―]"
const daySepRegexExpr = "[dD]"
const hourSepRegexExpr = "[hH]"
const minSepRegexExpr = "[mM]"
const secSepRegexExpr = "[sS]"

type timeRange struct {
	start         int64
	end           int64
}

type commentProperty struct {
	commentId   string
	author      string
	authorImage string
	text        string
}

type timeRangeProperty struct {
	start                int64
	end                  int64
	comments             []*commentProperty
	commentsDupCheckMap  map[string]bool
}

type videoProperty struct {
	videoId       string
	title         string
	updateAt      string
	timeRanges    []*timeRangeProperty
	twitterName   string
}

type channelProperty struct {
	videos             []*videoProperty
	videosDupCheckMap  map[string]int
}

type Builder struct {
	sourceDirPath          string
	resourceDirPath        string
	templateDirPath        string
	buildRootDirPath       string
	buildCacheDirPath      string
	channels               youtubedataapi.Channels
	databaseOperator       *database.DatabaseOperator
	verbose                bool
	timeRangeRegexp        *regexp.Regexp
	startEndSepRegexp      *regexp.Regexp
	daySepRegexp           *regexp.Regexp
	hourSepRegexp          *regexp.Regexp
	minSepRegexp           *regexp.Regexp
	secSepRegexp           *regexp.Regexp
	maxDuration            int64
	adjustStartTimeSpan    int64
	autoDetectUnitSpan     int64
	autoDetectThreshold    float64
	autoDetectRangeSec     int64
	autoDetectSkipDuration int64
	templates              *template.Template
}

type Clip struct {
	VideoId      string
	Title        string
	Start        int64
	End          int64
	Recommenders []string
	TwitterName  string
}

func (b *Builder)timeStringToSeconds(timeString string) (int64) {
	// parse 01:02, 01:02:03
	elems := strings.Split(timeString, ":")
	switch len(elems) {
	case 2:
		minString := elems[0]
		min, err := strconv.ParseInt(strings.TrimSpace(minString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse min string (minString = %v, timeString = %v)", minString, timeString)
			}
			return 0
		}
		secString := elems[1]
		sec, err := strconv.ParseInt(strings.TrimSpace(secString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse sec string (secString = %v, timeString = %v)", secString, timeString)
			}
			return 0
		}
		return sec + (min * 60)
	case 3:
		hourString := elems[0]
		hour, err := strconv.ParseInt(strings.TrimSpace(hourString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse hour string (hourString = %v, timeString = %v)", hourString, timeString)
			}
			return 0
		}
		minString := elems[1]
		min, err := strconv.ParseInt(strings.TrimSpace(minString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse min string (minString = %v, timeString = %v)", minString, timeString)
			}
			return 0
		}
		secString := elems[2]
		sec, err := strconv.ParseInt(strings.TrimSpace(secString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse sec string (secString = %v, timeString = %v)", secString, timeString)
			}
			return 0
		}
		return sec + (min * 60) + (hour * 3600)
	default:
		if b.verbose {
			log.Printf("can not parse tims string (timeString = %v, timeString = %v)", timeString)
		}
		return 0
	}
}

func (b *Builder)durationStringToSeconds(durationString string) (int64) {
	// parse 1h2m3s, 1h2m 1h3s, 1m2s, 1s
	var seconds int64
	durationString = strings.TrimPrefix(durationString, "PT")
	elems := b.daySepRegexp.Split(durationString, 2)
	if len(elems) == 2 {
		dayString := elems[0]
		day, err := strconv.ParseInt(strings.TrimSpace(dayString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse day string (dayString = %v, durationString = %v)", dayString, durationString)
			}
			return 0
		}
		seconds += day * 86400
		durationString = elems[1]
	}
	elems = b.hourSepRegexp.Split(durationString, 2)
	if len(elems) == 2 {
		hourString := elems[0]
		hour, err := strconv.ParseInt(strings.TrimSpace(hourString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse hour string (hourString = %v, durationString = %v)", hourString, durationString)
			}
			return 0
		}
		seconds += hour * 3600
		durationString = elems[1]
	}
	elems = b.minSepRegexp.Split(durationString, 2)
	if len(elems) == 2 {
		minString := elems[0]
		min, err := strconv.ParseInt(strings.TrimSpace(minString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse min string (minString = %v, durationString = %v)", minString, durationString)
			}
			return 0
		}
		seconds += min * 60
		durationString = elems[1]
	}
	elems = b.secSepRegexp.Split(durationString, 2)
	if len(elems) == 2 {
		secString := elems[0]
		sec, err := strconv.ParseInt(strings.TrimSpace(secString), 10, 64)
		if err != nil {
			if b.verbose {
				log.Printf("can not parse sec string (minString = %v, durationString = %v)", secString, durationString)
			}
			return 0
		}
		seconds += sec
	}
	return seconds
}

func (b *Builder)parseTimeRange(timeRangeString string) (*timeRange) {
	elems := b.startEndSepRegexp.Split(timeRangeString, 2)
	if len(elems) == 2 {
		startTimeString := elems[0]
		startSeconds := b.timeStringToSeconds(startTimeString)
		endTimeString := elems[1]
		endSeconds := b.timeStringToSeconds(endTimeString)
		if endSeconds < startSeconds  {
			endSeconds = 0
		}
		if endSeconds > startSeconds + b.maxDuration {
			endSeconds = startSeconds + b.maxDuration
		}
		return &timeRange {
			start: startSeconds,
			end: endSeconds,
		}
	}
	elems = strings.SplitN(timeRangeString, "@", 2)
	if len(elems) == 2 {
		startTimeString := elems[0]
		startSeconds := b.timeStringToSeconds(startTimeString)
		timeRange := &timeRange{
			start: startSeconds,
			end: 0,
		}
		durationString := elems[1]
		durationSeconds := b.durationStringToSeconds(durationString)
		if durationSeconds > b.maxDuration {
			timeRange.end = startSeconds + b.maxDuration
		} else if durationSeconds > 0 {
			timeRange.end = startSeconds +  durationSeconds
		}
		return timeRange
	}
	startSeconds := b.timeStringToSeconds(timeRangeString)
	return &timeRange {
		start: startSeconds,
		end: 0,
	}
}

func (b *Builder)parseTimeRanges(textOriginal string) ([]*timeRange) {
	timeRanges := make([]*timeRange, 0)
	timeRageStringList := b.timeRangeRegexp.FindAllString(textOriginal, -1)
	for _, timeRangeString := range timeRageStringList {
		timeRange := b.parseTimeRange(timeRangeString)
		if timeRange.start == 0 {
			continue
		}
		timeRanges = append(timeRanges, timeRange)
	}
	return timeRanges
}

func (b *Builder)adjustTimRanges(timeRanges []*timeRangeProperty) ([]*timeRangeProperty) {
	// start時間がadjustStartTimeSpan以内のずれなら一つにまとめる、この時startとendは最大になるようにする
	for {
		var prevTimeRange *timeRangeProperty
		var lastTimeRange *timeRangeProperty
		adjustIdx := -1
		for idx, timeRange := range timeRanges {
			if lastTimeRange == nil {
				lastTimeRange = timeRange
				continue
			}
			prevTimeRange = lastTimeRange
			lastTimeRange = timeRange
			if prevTimeRange.start + b.adjustStartTimeSpan >= lastTimeRange.start  {
				if prevTimeRange.end < lastTimeRange.end {
					prevTimeRange.end = lastTimeRange.end
				}
				for _, comment := range lastTimeRange.comments {
					prevTimeRange.comments = append(prevTimeRange.comments, comment)
				}
				adjustIdx = idx
				break
			}
		}
		if adjustIdx == -1 {
			break
		}
		timeRanges = append(timeRanges[:adjustIdx], timeRanges[adjustIdx+1:]...)
	}
	// 任意の要素のend時間が次の要素のstartよりも大きい場合はendを次の要素のstartにする
	var prevTimeRange *timeRangeProperty
	var lastTimeRange *timeRangeProperty
	for _, timeRange := range timeRanges {
		if lastTimeRange == nil {
			lastTimeRange = timeRange
			continue
		}
		prevTimeRange = lastTimeRange
		lastTimeRange = timeRange
		if prevTimeRange.end > lastTimeRange.start  {
			prevTimeRange.end = lastTimeRange.start
		}
	}
	return timeRanges
}

func (b *Builder)computeStandardDeviationThreshold(counts[]float64) (float64) {
	var total float64
	var n float64 = (float64)(len(counts))
	for _, count := range counts {
		total += count
	}
	average := total / n
	var powTotal float64
	for _, count := range counts {
		powTotal += math.Pow(average - count, 2)
	}
	return average + (math.Sqrt(powTotal/n) * b.autoDetectThreshold)
}

func (b *Builder)containsPatterns(text string) (bool) {
	patterns := [...] string{"w", "W", "W", "ｗ", "lol", "LOL", "草", "くさ", "笑", "ワロ", "さす", "かっこいい", "ナイス", "ないす"}
	for _, s := range patterns {
		if strings.Contains(text, s) == true {
			return true
		}
	}
	return false
}

func (b *Builder)makeChannelProperty(channel *database.Channel) (*channelProperty, error) {
	// create channel property
	channelProp := &channelProperty{
		videos: make([]*videoProperty, 0),
		videosDupCheckMap: make(map[string]int),
	}
	// comments
	comments, err := b.databaseOperator.GetAllCommentsByChannelIdAndLikeText(channel.ChannelId, "%:%")
	if err != nil {
		return nil, errors.Wrapf(err, "can not get comments from database (channelId = %v)", channel.ChannelId)
	}
	for _, comment := range comments {
		video, ok, err := b.databaseOperator.GetVideoByVideoId(comment.VideoId)
		if err != nil {
			return nil, errors.Wrapf(err, "can not get video from database (videoId = %v, commentId = %v)", comment.VideoId, comment.CommentId)
		}
		if !ok {
			if b.verbose {
				log.Printf("skip comment not found video (videoId = %v, commentId = %v)", comment.VideoId, comment.CommentId)
			}
			continue
		}
		if !video.StatusEmbeddable {
			if b.verbose {
				log.Printf("skip comment because unembeddable video (videoId = %v, commentId = %v)", comment.VideoId, comment.CommentId)
			}
			continue
		}
		duration := b.durationStringToSeconds(video.Duration)
		if b.verbose {
			log.Printf("videoId = %v, video duration = %v, commentId = %v", video.VideoId, duration, comment.CommentId)
		}
		idx, ok := channelProp.videosDupCheckMap[comment.VideoId]
		if !ok {
			videoProp := &videoProperty{
				videoId: comment.VideoId,
				title: video.Title,
				updateAt: comment.UpdateAt,
				timeRanges: make([]*timeRangeProperty, 0),
				twitterName: channel.TwitterName,
			}
			idx = len(channelProp.videos)
			channelProp.videos = append(channelProp.videos, videoProp)
			channelProp.videosDupCheckMap[comment.VideoId] = idx
		}
		videoProp := channelProp.videos[idx]
		// convert text to time ranges
		timeRanges := b.parseTimeRanges(comment.TextOriginal)
		if len(timeRanges) > 0 && videoProp.updateAt < comment.UpdateAt {
			videoProp.updateAt = comment.UpdateAt
		}
		for _, timeRange := range timeRanges {
			if timeRange.start > duration || timeRange.start < 0 {
				if b.verbose {
					log.Printf("time range start over duration (videoId = %v, commentId = %v, comment start = %v, video duration = %v)",
						comment.VideoId, comment.CommentId, timeRange.start, duration)
				}
				timeRange.start = duration
			}
			if timeRange.end > duration || timeRange.end < 0 {
				if b.verbose {
					log.Printf("time range end over duration (videoId = %v, commentId = %v, comment end = %v, video  duration = %v)",
						comment.VideoId, comment.CommentId, timeRange.end, duration)
				}
				timeRange.end = duration
			}
			timeRangeProp := &timeRangeProperty{
				start: timeRange.start,
				end: timeRange.end,
				comments: make([]*commentProperty, 0),
				commentsDupCheckMap: make(map[string]bool),
			}
			commentProperty := &commentProperty{
				commentId: comment.CommentId,
				author: comment.AuthorDisplayName,
				authorImage: comment.AuthorProfileImageUrl,
				text: comment.TextOriginal,
			}
			_, ok := timeRangeProp.commentsDupCheckMap[comment.CommentId]
			if !ok {
				timeRangeProp.comments = append(timeRangeProp.comments, commentProperty)
				timeRangeProp.commentsDupCheckMap[comment.CommentId] = true
			}
			videoProp.timeRanges = append(videoProp.timeRanges, timeRangeProp)
		}
	}
	// check live chat comments
	videos, err := b.databaseOperator.GetVideosByChannelId(channel.ChannelId)
	if err != nil {
		return nil, errors.Wrapf(err, "can not get videos from database (channelId = %v)", channel.ChannelId)
	}
	for _, video := range videos {
		if !video.StatusEmbeddable {
			if b.verbose {
				log.Printf("skip live chat comments because unembeddable video (channelId = %v, videoId = %v)", channel.ChannelId, video.VideoId)
			}
			continue
		}
		duration := b.durationStringToSeconds(video.Duration)
		if b.verbose {
			log.Printf("videoId = %v, video duration = %v", video.VideoId, duration)
		}
		liveChatComments, err := b.databaseOperator.GetLiveChatCommentsByVideoId(video.VideoId)
		if err != nil {
			return nil, errors.Wrapf(err, "can not get live chat comment from database (channelId = %v, videoId = %v)", channel.ChannelId, video.VideoId)
		}
		size := (duration / b.autoDetectUnitSpan) + 1
		counts := make([]float64, size, size)
		for _, liveChatComment := range liveChatComments {
			offsetMsec, err := strconv.ParseInt(liveChatComment.VideoOffsetTimeMsec, 10, 64)
			if err != nil {
				log.Printf("can not parse videoOffsetTimeMsec (%v)", liveChatComment.VideoOffsetTimeMsec)
				continue
			}
			offset := offsetMsec / 1000
			if offset < b.autoDetectSkipDuration || offset > duration {
				continue
			}
			idx := offset / b.autoDetectUnitSpan
			if liveChatComment.PurchaseAmountText != "" {
				counts[idx] += 2
			} else if b.containsPatterns(liveChatComment.MessageText) {
				counts[idx] += 2
			} else {
				counts[idx] += 1
			}
		}
		// get threashold from standard deviation
		threshold := b.computeStandardDeviationThreshold(counts[(b.autoDetectSkipDuration/b.autoDetectUnitSpan):])
		if b.verbose {
			log.Printf("standard deviation threshold = %v", threshold)
		}
		if threshold != 0 {
			for i, c := range counts {
				if c < threshold {
					continue
				}
				if b.verbose {
					log.Printf("idx = %v score = %v", i, c)
				}
				idx, ok := channelProp.videosDupCheckMap[video.VideoId]
				if !ok {
					videoProp := &videoProperty{
						videoId: video.VideoId,
						title: video.Title,
						updateAt: video.PublishdAt,
						timeRanges: make([]*timeRangeProperty, 0),
					}
					idx = len(channelProp.videos)
					channelProp.videos = append(channelProp.videos, videoProp)
					channelProp.videosDupCheckMap[video.VideoId] = idx
				}
				videoProp := channelProp.videos[idx]
				if videoProp.updateAt < video.PublishdAt {
					videoProp.updateAt = video.PublishdAt
				}
				start := ((int64)(i) * b.autoDetectUnitSpan) - b.autoDetectRangeSec
				if start < 0 {
					start = 0;
				}
				if start > duration {
					start = duration;
				}
				end := ((int64)(i) * b.autoDetectUnitSpan) + (b.autoDetectRangeSec / 3 * 2)
				if end > duration {
					end = duration;
				}
				timeRangeProp := &timeRangeProperty{
					start: start,
					end: end,
					comments: make([]*commentProperty, 0),
					commentsDupCheckMap: make(map[string]bool),
				}
				commentId := fmt.Sprintf("AD.%v.%v.%v.%v.%v", channel.ChannelId, video.VideoId, i, c, b.autoDetectUnitSpan)
				commentProperty := &commentProperty{
					commentId: commentId,
					author: "Automatic detection by clipper",
					authorImage: "",
					text: fmt.Sprintf("Automatic detection score %v", c),
				}
				_, ok = timeRangeProp.commentsDupCheckMap[commentId]
				if !ok {
					timeRangeProp.comments = append(timeRangeProp.comments, commentProperty)
					timeRangeProp.commentsDupCheckMap[commentId] = true
				}
				videoProp.timeRanges = append(videoProp.timeRanges, timeRangeProp)
			}
		}
	}
	// sort and adjust
	sort.Slice(channelProp.videos, func(i, j int) bool {
		return channelProp.videos[i].updateAt > channelProp.videos[j].updateAt
	})
	for _, videoProp := range channelProp.videos {
                sort.Slice(videoProp.timeRanges, func(i, j int) bool {
                        return videoProp.timeRanges[i].start < videoProp.timeRanges[j].start
                })
                videoProp.timeRanges = b.adjustTimRanges(videoProp.timeRanges)
	}
	return channelProp, nil
}

func (b Builder)contains(s []string, v string) (bool) {
	for _, sv := range s {
		if sv == v {
			return true
		}
	}
	return false
}

func (b *Builder)buildClips(channel *database.Channel) ([]*Clip, error) {
	channelProp, err := b.makeChannelProperty(channel)
	if err != nil {
		return nil, errors.Wrapf(err, "can not make clip property (channelId = %v)", channel.ChannelId)
	}
	clips := make([]*Clip, 0)
	for _, video := range channelProp.videos {
		for _, timeRange := range video.timeRanges {
			authors := make([]string, 0, len(timeRange.comments))
			for _, comment := range timeRange.comments {
				if b.contains(authors, comment.author) {
					continue
				}
				authors = append(authors, comment.author)
			}
			clip := &Clip {
				VideoId: video.videoId,
				Title: video.title,
				Start: timeRange.start,
				End: timeRange.end,
				Recommenders: authors,
				TwitterName: video.twitterName,
			}
			clips = append(clips, clip)
		}
	}
	return clips, nil
}

func (b *Builder)Build(rebuild bool) (error) {
	dbChannels := make([]*database.Channel, 0)
	for _, channel := range b.channels {
		dbChannel, ok, err := b.databaseOperator.GetChannelByChannelId(channel.ChannelId)
		if err != nil {
			return errors.Wrapf(err, "can not get chennal by channelId from database (channelId = %v)", channel.ChannelId)
		}
		if !ok {
			continue
		}
		dbChannels = append(dbChannels, dbChannel)
	}
	// create channel page
	for _, dbChannel := range dbChannels {
		pageHtmlPath := filepath.Join(b.buildRootDirPath, dbChannel.ChannelId + ".html")
		pageHtmlFile, err := os.OpenFile(pageHtmlPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return errors.Wrapf(err, "can not open page html file (path = %v)", pageHtmlPath)
		}
		defer pageHtmlFile.Close()
		err = b.templates.ExecuteTemplate(pageHtmlFile, "page.tmpl", dbChannel)
		if err != nil {
			return errors.Wrapf(err, "can not write to page html file (path = %v)", pageHtmlPath)
		}
	}
	// create index.html
	indexHtmlPath := filepath.Join(b.buildRootDirPath, "index.html")
	indexHtmlFile, err := os.OpenFile(indexHtmlPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "can not open index html file (path = %v)", indexHtmlPath)
	}
	defer indexHtmlFile.Close()
	err = b.templates.ExecuteTemplate(indexHtmlFile, "index.tmpl", dbChannels)
	if err != nil {
		return errors.Wrapf(err, "can not write to index html file (path = %v)", indexHtmlPath)
	}
	// create channel page in database
	somethingModified := false
	for _, dbChannel := range dbChannels {
		lastChannelPage, lastChannelPageOk, err := b.databaseOperator.GetChannelPageByChannelId(dbChannel.ChannelId)
                if err != nil {
                        return  errors.Wrapf(err, "can not get channel page from database (channelId = %v)", dbChannel.ChannelId)
                }
		clips, err := b.buildClips(dbChannel)
		if err != nil {
			return errors.Wrapf(err, "can not get build page (channelId = %v)", dbChannel.ChannelId)
		}
		clipsJsonBytes, err := json.Marshal(clips)
		if err != nil {
			return errors.Wrapf(err, "can not marshal json (channelId = %v)", dbChannel.ChannelId)
		}
		newChannelPageSha1Digest := fmt.Sprintf("%x", sha1.Sum(clipsJsonBytes))
		if !rebuild && lastChannelPageOk && lastChannelPage.Sha1Digest == newChannelPageSha1Digest {
			if b.verbose {
				log.Printf("skip because same sha1 digest of channel page (oldSha1Digest = %v, newSha1Digest = %v)", lastChannelPage.Sha1Digest, newChannelPageSha1Digest)
			}
			continue
		}
		somethingModified = true
		channelPageJsonPath := filepath.Join(b.buildCacheDirPath, dbChannel.ChannelId + ".json")
		err = ioutil.WriteFile(channelPageJsonPath, clipsJsonBytes, 0644)
		if err != nil {
			return errors.Wrapf(err, "can not write json to file (channelId = %v, path = %v)", dbChannel.ChannelId, channelPageJsonPath)
		}
		channelPageJsonSha1DigestPath := filepath.Join(b.buildCacheDirPath, dbChannel.ChannelId + ".json.sha1")
		err = ioutil.WriteFile(channelPageJsonSha1DigestPath, []byte(newChannelPageSha1Digest), 0644)
		if err != nil {
			return errors.Wrapf(err, "can not write sha1 digest of json to file (channelId = %v, path = %v)", dbChannel.ChannelId, channelPageJsonSha1DigestPath)
		}
		var tweetId int64
		if lastChannelPageOk {
			tweetId = lastChannelPage.TweetId
		} else {
			tweetId = -1
		}
		err = b.databaseOperator.UpdateSha1DigestAndDirtyOfChannelPage(dbChannel.ChannelId, newChannelPageSha1Digest, 1, tweetId)
                if err != nil {
                        return  errors.Wrapf(err, "can not update sha1 digest and dirty of channelPage (channelId = %v, newChannelPageSha1Digest = %v)", dbChannel.ChannelId, newChannelPageSha1Digest)
                }
	}
	// create index page in database
	lastChannelPage, lastChannelPageOk, err := b.databaseOperator.GetChannelPageByChannelId("index")
        if err != nil {
		return  errors.Wrapf(err, "can not get channel page from database (channelId = index)")
	}
	if !rebuild && lastChannelPageOk && somethingModified == false {
		if b.verbose {
			log.Printf("skip because same sha1 digest of all channel pages")
		}
		return nil
	}
	var tweetId int64
	if lastChannelPageOk {
		tweetId = lastChannelPage.TweetId
	} else {
		tweetId = -1
	}
	err = b.databaseOperator.UpdateSha1DigestAndDirtyOfChannelPage("index", "", 1, tweetId)
        if err != nil {
		return  errors.Wrapf(err, "can not update sha1 digest and dirty of channelPage (channelId = index)")
	}
	return nil
}

func NewBuilder(
	sourceDirPath string,
	buildDirPath string,
	maxDuration int64,
	adjustStartTimeSpan int64,
	autoDetectUnitSpan int64,
	autoDetectThreshold float64,
	autoDetectRangeSec int64,
	autoDetectSkipDuration int64,
	channels youtubedataapi.Channels,
	databaseOperator *database.DatabaseOperator,
	verbose bool) (*Builder, error)  {
        if buildDirPath == "" {
                return nil, errors.New("no build directory path")
        }
	if maxDuration == 0 {
                return nil, errors.New("no max time range")
	}
        resourceDirPath := filepath.Join(sourceDirPath, "resource")
        templateDirPath := filepath.Join(sourceDirPath, "template")
        _, err := os.Stat(buildDirPath)
        if err != nil {
                err := os.MkdirAll(buildDirPath, 0755)
                if err != nil {
                        return nil, errors.Wrapf(err, "can not create directory (%v)", buildDirPath)
                }
        }
        timeRangeRegexp, err := regexp.Compile(timeRangeRegexpExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "can not compile timeRangeRegexpExpr (%v)", timeRangeRegexpExpr)
	}
	startEndSepRegexp, err := regexp.Compile(startEndSepRegexExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "can not compile startEndSepRegexExpr (%v)", startEndSepRegexExpr)
	}
	daySepRegexp, err := regexp.Compile(daySepRegexExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "can not compile daySepRegexExpr (%v)", daySepRegexExpr)
	}
	hourSepRegexp, err := regexp.Compile(hourSepRegexExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "can not compile hourSepRegexExpr (%v)", hourSepRegexExpr)
	}
	minSepRegexp, err := regexp.Compile(minSepRegexExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "can not compile minSepRegexExpr (%v)", minSepRegexExpr)
	}
	secSepRegexp, err := regexp.Compile(secSepRegexExpr)
	if err != nil {
		return nil, errors.Wrapf(err, "can not compile secSepRegexExpr (%v)", secSepRegexExpr)
	}
	templatePattern := filepath.Join(templateDirPath, "*.tmpl")
	templates, err := template.ParseGlob(templatePattern)
	if err != nil {
		return nil, errors.Wrapf(err, "can not parse templates (template pattern = %v)", templatePattern)
	}
        buildRootDirPath := filepath.Join(buildDirPath, "root")
         _, err = os.Stat(buildRootDirPath)
	if err != nil {
                err := os.MkdirAll(buildRootDirPath, 0755)
                if err != nil {
                        return nil, errors.Wrapf(err, "can not create directory (%v)", buildRootDirPath)
                }
        }
        buildCacheDirPath := filepath.Join(buildDirPath, "cache")
         _, err = os.Stat(buildCacheDirPath)
	if err != nil {
                err := os.MkdirAll(buildCacheDirPath, 0755)
                if err != nil {
                        return nil, errors.Wrapf(err, "can not create directory (%v)", buildCacheDirPath)
                }
        }
	err = copytool.Copy(resourceDirPath, buildDirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "can not copy resource to build (%v -> %v)", resourceDirPath, buildDirPath)
	}
	return &Builder {
		sourceDirPath: sourceDirPath,
		resourceDirPath: resourceDirPath,
		templateDirPath: templateDirPath,
		buildRootDirPath: buildRootDirPath,
		buildCacheDirPath: buildCacheDirPath,
		channels: channels,
		databaseOperator: databaseOperator,
		verbose: verbose,
		timeRangeRegexp: timeRangeRegexp,
		startEndSepRegexp: startEndSepRegexp,
		daySepRegexp: daySepRegexp,
		hourSepRegexp: hourSepRegexp,
		minSepRegexp: minSepRegexp,
		secSepRegexp: secSepRegexp,
		maxDuration: maxDuration,
		adjustStartTimeSpan: adjustStartTimeSpan,
		autoDetectUnitSpan: autoDetectUnitSpan,
		autoDetectThreshold: autoDetectThreshold,
		autoDetectRangeSec: autoDetectRangeSec,
		autoDetectSkipDuration: autoDetectSkipDuration,
		templates: templates,
	}, nil
}

