package builder

import (
        "os"
        "log"
        "sort"
        "regexp"
        "strings"
        "strconv"
        "github.com/pkg/errors"
        "github.com/munoudesu/clipper/youtubedataapi"
        "github.com/munoudesu/clipper/database"
)

const timeRangeRegexpExpr = "[0-9]{1,2}:[0-9]{1,2}(:[0-9]{1,2})?(([ 　]*[~-～－―][ 　]*[0-9]{1,2}:[0-9]{1,2}(:[0-9]{1,2})?)|([ 　]*@[ 　]*([0-9]+[hH])?([0-9]+[mM])?([0-9]+[sS])?))?"
const startEndSepRegexExpr = "[~-～－―]"

type timeRange struct {
	start         int64
	end           int64
}

type commentProperty struct {
	author      string
	authorImage string
	text        string
}

type timeRangeProperty struct {
	start    int64
	end      int64
	comments map[commentProperty]bool
}

type videoProperty struct {
	timeRangeList           []*timeRangeProperty
}

type pageProperty struct {
	videos map[string]*videoProperty
}

type Builder struct {
	buildDirPath        string
	channels            youtubedataapi.Channels
	databaseOperator    *database.DatabaseOperator
	timeRangeRegexp     *regexp.Regexp
	startEndSepRegexp   *regexp.Regexp
	maxDuration         int64
	adjustStartTimeSpan int64
}

func (b *Builder)timeStringToSeconds(timeString string) (int64) {
	// parse 01:02, 01:02:03
	elems := strings.Split(timeString, ":")
	switch len(elems) {
	case 2:
		minString := elems[0]
		min, err := strconv.ParseInt(minString, 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v)", minString)
			return 0
		}
		secString := elems[1]
		sec, err := strconv.ParseInt(secString, 10, 64)
		if err != nil {
			log.Printf("can not parse sec string (secString = %v)", secString)
			return 0
		}
		return sec + (min * 60)
	case 3:
		hourString := elems[0]
		hour, err := strconv.ParseInt(hourString, 10, 64)
		if err != nil {
			log.Printf("can not parse hour string (hourString = %v)", hourString)
			return 0
		}
		minString := elems[1]
		min, err := strconv.ParseInt(minString, 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v)", minString)
			return 0
		}
		secString := elems[2]
		sec, err := strconv.ParseInt(secString, 10, 64)
		if err != nil {
			log.Printf("can not parse sec string (secString = %v)", secString)
			return 0
		}
		return sec + (min * 60) + (hour * 3600)
	default:
		log.Printf("can not parse tims string (timeString = %v)", timeString)
		return 0
	}
}

func (b *Builder)durationStringToSeconds(durationString string) (int64) {
	// parse 1h2m3s, 1h2m 1h3s, 1m2s, 1s
	var seconds int64
	elems := strings.SplitN(durationString, "h", 2)
	if len(elems) == 2 {
		hourString := elems[0]
		hour, err := strconv.ParseInt(hourString, 10, 64)
		if err != nil {
			log.Printf("can not parse hour string (hourString = %v)", hourString)
			return 0
		}
		seconds += hour * 3600
		durationString = elems[1]
	}
	elems = strings.SplitN(durationString, "m", 2)
	if len(elems) == 2 {
		minString := elems[0]
		min, err := strconv.ParseInt(minString, 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v)", minString)
			return 0
		}
		seconds += min * 60
		durationString = elems[1]
	}
	elems = strings.SplitN(durationString, "s", 2)
	if len(elems) == 2 {
		secString := elems[0]
		sec, err := strconv.ParseInt(secString, 10, 64)
		if err != nil {
			log.Printf("can not parse sec string (minString = %v)", secString)
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

func (b *Builder)parseTimeRangeList(textOriginal string) ([]*timeRange) {
	timeRangeList := make([]*timeRange, 0)
	timeRageStringList := b.timeRangeRegexp.FindAllString(textOriginal, -1)
	for _, timeRangeString := range timeRageStringList {
		timeRange := b.parseTimeRange(timeRangeString)
		if timeRange.start == 0 {
			continue
		}
		timeRangeList = append(timeRangeList, timeRange)
	}
	return timeRangeList
}

func (b *Builder)adjustTimRangeList(timeRangeList []*timeRangeProperty) ([]*timeRangeProperty) {
	// start時間がadjustStartTimeSpan以内のずれなら一つにまとめる、この時startとendは最大になるようにする
	for {
		var prevTimeRange *timeRangeProperty
		var lastTimeRange *timeRangeProperty
		adjustIdx := -1
		for idx, timeRange := range timeRangeList {
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
				for comment, _ := range lastTimeRange.comments {
					prevTimeRange.comments[comment] = true
				}
				adjustIdx = idx
				break
			}
		}
		if adjustIdx == -1 {
			break
		}
		timeRangeList = append(timeRangeList[:adjustIdx], timeRangeList[adjustIdx+1:]...)
	}
	// 任意の要素のend時間が次の要素のstartよりも大きい場合はendを次の要素のstartにする
	var prevTimeRange *timeRangeProperty
	var lastTimeRange *timeRangeProperty
	for _, timeRange := range timeRangeList {
		if lastTimeRange == nil {
			lastTimeRange = timeRange
			continue
		}
		prevTimeRange = lastTimeRange
		lastTimeRange = timeRange
		if prevTimeRange.end >= lastTimeRange.start  {
			prevTimeRange.end = lastTimeRange.start
		}
	}
	return timeRangeList
}

func (b *Builder)makePageProperty(channel *youtubedataapi.Channel) (*pageProperty, error) {
	comments, err := b.databaseOperator.GetAllCommentsByChannelIdAndLikeText(channel.ChannelId, "%:%")
	if err != nil {
		return nil, errors.Wrapf(err, "can not get comments from database (channelId = %v)", channel.ChannelId)
	}
	pageProp := &pageProperty{
		videos: make(map[string]*videoProperty),
	}
	for _, comment := range comments {
		videoProp, ok := pageProp.videos[comment.VideoId]
		if !ok {
			videoProp = &videoProperty{
				timeRangeList: make([]*timeRangeProperty, 0),
			}
			pageProp.videos[comment.VideoId] = videoProp
		}
		timeRangeList := b.parseTimeRangeList(comment.TextOriginal)
		for _, timeRange := range timeRangeList {
			timeRangeProp := &timeRangeProperty{
				start: timeRange.start,
				end: timeRange.end,
				comments: make(map[commentProperty]bool),
			}
			commentProperty := commentProperty{
				author: comment.AuthorDisplayName,
				authorImage: comment.AuthorProfileImageUrl,
				text: comment.TextOriginal,
			}
			timeRangeProp.comments[commentProperty] = true
			videoProp.timeRangeList = append(videoProp.timeRangeList, timeRangeProp)
			continue
		}
		sort.Slice(videoProp.timeRangeList, func(i, j int) bool {
			return videoProp.timeRangeList[i].start < videoProp.timeRangeList[j].start
		})
		videoProp.timeRangeList = b.adjustTimRangeList(videoProp.timeRangeList)
	}
	return pageProp, nil
}

func (b *Builder)buildPage(channel *youtubedataapi.Channel) (string, error) {
	pageProp, err := b.makePageProperty(channel)
	if err != nil {
		return "", errors.Wrapf(err, "can not make page property (channelId = %v)", channel.ChannelId)
	}

	for videoId, videoProp := range pageProp.videos {
		for _, timeRangeProp := range videoProp.timeRangeList {
			log.Printf("===========")
			log.Printf("channelId = %v, videoId = %v, start = %vs, end = %vs", channel.ChannelId, videoId, timeRangeProp.start, timeRangeProp.end)
			authors := ""
			for commentProp, _ := range timeRangeProp.comments {
				if authors == "" {
					authors = commentProp.author
				} else {
					authors += ", " + commentProp.author
				}
			}
			log.Printf("authors = %v", authors)
		}
	}

	return "", nil
}

func (b *Builder)Build() (error) {
	for _, channel := range b.channels {
		lastChannelPage, ok, err := b.databaseOperator.GetChannelPageByChannelId(channel.ChannelId)
                if err != nil {
                        return  errors.Wrapf(err, "can not get channel page from database (channelId = %v)", channel.ChannelId)
                }
log.Printf("%v", lastChannelPage)
		newPageHash, err := b.buildPage(channel)
		if err != nil {
			return errors.Wrapf(err, "can not get build page (channelId = %v)", channel.ChannelId)
		}
		if ok && lastChannelPage.PageHash == newPageHash {
			continue
		}
		err = b.databaseOperator.UpdatePageHashAndDirtyOfChannelPage(channel.ChannelId, newPageHash, 1)
                if err != nil {
                        return  errors.Wrapf(err, "can not update page hash and dirty of channelPage (channelId = %v, newPageHash = %v)", channel.ChannelId, newPageHash)
                }
	}
	return nil
}

func NewBuilder(buildDirPath string, maxDuration int64, adjustStartTimeSpan int64, channels youtubedataapi.Channels, databaseOperator *database.DatabaseOperator) (*Builder, error)  {
        if buildDirPath == "" {
                return nil, errors.New("no build directory path")
        }
	if maxDuration == 0 {
                return nil, errors.New("no max time range")
	}
        _, err := os.Stat(buildDirPath)
        if err != nil {
                err := os.MkdirAll(buildDirPath, 0755)
                if err != nil {
                        return nil, errors.Errorf("can not create directory (%v)", buildDirPath)
                }
        }
        timeRangeRegexp, err := regexp.Compile(timeRangeRegexpExpr)
	if err != nil {
		return nil, errors.Errorf("can not compile timeRangeRegexpExpr (%v)", timeRangeRegexpExpr)
	}
	startEndSepRegexp, err := regexp.Compile(startEndSepRegexExpr)
	if err != nil {
		return nil, errors.Errorf("can not compile startEndSepRegexExpr (%v)", startEndSepRegexExpr)
	}
	return &Builder {
		buildDirPath: buildDirPath,
		channels: channels,
		databaseOperator: databaseOperator,
		timeRangeRegexp: timeRangeRegexp,
		startEndSepRegexp: startEndSepRegexp,
		maxDuration: maxDuration,
		adjustStartTimeSpan: adjustStartTimeSpan,
	}, nil
}

