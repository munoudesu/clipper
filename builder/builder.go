package builder

import (
        "os"
        "log"
        "regexp"
        "strings"
        "strconv"
        "github.com/pkg/errors"
        "github.com/munoudesu/clipper/youtubedataapi"
        "github.com/munoudesu/clipper/database"
)

const timeRangeRegexpExpr = "([0-9]{2}:[0-9]{2}(:[0-9]{2})?|[0-9]:[0-9](:[0-9])?)(([~-～－―]([0-9]{2}:[0-9]{2}(:[0-9]{2})?|[0-9]:[0-9](:[0-9])?))|(@([0-9]+[hH])?([0-9]+[mM])?([0-9]+[sS])?))?"
const startEndSepRegexExpr = "[~-～－―]"

type timeRange struct {
	start         int64
	end           int64
}

type timeRangeComments struct {
	start         int64
	end           int64
	textOriginals map[string]bool
}

type pageVideo struct {
	timeRangeCommentsMap map[int64]*timeRangeComments
}

type pageSource struct {
	videos map[string]*pageVideo
}

type Builder struct {
	buildDirPath      string
	channels          youtubedataapi.Channels
	databaseOperator  *database.DatabaseOperator
	timeRangeRegexp   *regexp.Regexp
	startEndSepRegexp *regexp.Regexp
	maxTimeRange      int64
	adjustStartTime   int64
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

func (b *Builder)periodicStringToSeconds(periodicString string) (int64) {
	// parse 1h2m3s, 1h2m 1h3s, 1m2s, 1s
	var seconds int64
	elems := strings.SplitN(periodicString, "h", 2)
	if len(elems) == 2 {
		hourString := elems[0]
		hour, err := strconv.ParseInt(hourString, 10, 64)
		if err != nil {
			log.Printf("can not parse hour string (hourString = %v)", hourString)
			return 0
		}
		seconds += hour * 3600
		periodicString = elems[1]
	}
	elems = strings.SplitN(periodicString, "m", 2)
	if len(elems) == 2 {
		minString := elems[0]
		min, err := strconv.ParseInt(minString, 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v)", minString)
			return 0
		}
		seconds += min * 60
		periodicString = elems[1]
	}
	elems = strings.SplitN(periodicString, "s", 2)
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
		if endSeconds > startSeconds + b.maxTimeRange {
			endSeconds = startSeconds + b.maxTimeRange
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
		periodicString := elems[1]
		periodicSeconds := b.periodicStringToSeconds(periodicString)
		if periodicSeconds > b.maxTimeRange {
			timeRange.end = startSeconds + b.maxTimeRange
		} else if periodicSeconds > 0 {
			timeRange.end = startSeconds +  periodicSeconds
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

func (b *Builder)buildPage(channel *youtubedataapi.Channel) (string, error) {
log.Printf("---")
	comments, err := b.databaseOperator.GetAllCommentsByChannelIdAndLikeText(channel.ChannelId, "%:%")
	if err != nil {
		return "", errors.Wrapf(err, "can not get comments from database (channelId = %v)", channel.ChannelId)
	}
log.Printf("%v", len(comments))
	pageSource := &pageSource{
		videos: make(map[string]*pageVideo),
	}
	for _, comment := range comments {
		timeRangeList := b.parseTimeRangeList(comment.TextOriginal)
		pVideo, ok := pageSource.videos[comment.VideoId]
		if !ok {
			pVideo = &pageVideo{
				timeRangeCommentsMap: make(map[int64]*timeRangeComments),
			}
			pageSource.videos[comment.VideoId] = pVideo
		}
		for _, timeRange := range timeRangeList {
			adjustTime := timeRange.start/b.adjustStartTime
			tComments, ok := pVideo.timeRangeCommentsMap[adjustTime]
			if !ok {
				tComments = &timeRangeComments{
					start: timeRange.start,
					end: timeRange.end,
					textOriginals: make(map[string]bool),
				}
				pVideo.timeRangeCommentsMap[adjustTime] = tComments
			}
			if tComments.end == 0 && timeRange.end != 0 {
				tComments.end = timeRange.end
			}
			tComments.textOriginals[comment.TextOriginal] = true
		}
	}

	for videoId, pVideo := range pageSource.videos {
		for _, tComments := range pVideo.timeRangeCommentsMap {
			log.Printf("===========")
			log.Printf("channelId = %v, videoId = %v, start = %vs, end = %vs", channel.ChannelId, videoId,  tComments.start, tComments.end)
			for textOriginal,_ := range tComments.textOriginals {
				log.Printf("---")
				log.Printf("%v", textOriginal)
			}
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

func NewBuilder(buildDirPath string, maxTimeRange int64, adjustStartTime int64, channels youtubedataapi.Channels, databaseOperator *database.DatabaseOperator) (*Builder, error)  {
        if buildDirPath == "" {
                return nil, errors.New("no build directory path")
        }
	if maxTimeRange == 0 {
                return nil, errors.New("no max time range")
	}
	if adjustStartTime == 0 {
                return nil, errors.New("no adjust start time")
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
		maxTimeRange: maxTimeRange,
		adjustStartTime: adjustStartTime,
	}, nil
}

