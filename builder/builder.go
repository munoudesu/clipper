package builder

import (
        "fmt"
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
	// XXXXX comment updatedAt 最新のもの TODO
	timeRangeList []*timeRangeProperty
}

type channelProperty struct {
	videos map[string]*videoProperty
}

type Builder struct {
	sourceDirPath         string
	resourceDirPath       string
	templateDirPath       string
	buildDirPath          string
	buildJsDirPath        string
	buildJsonDirPath      string
	channels              youtubedataapi.Channels
	databaseOperator      *database.DatabaseOperator
	timeRangeRegexp       *regexp.Regexp
	startEndSepRegexp     *regexp.Regexp
	maxDuration           int64
	adjustStartTimeSpan   int64
	templates             *template.Template
}

func (b *Builder)timeStringToSeconds(timeString string) (int64) {
	// parse 01:02, 01:02:03
	elems := strings.Split(timeString, ":")
	switch len(elems) {
	case 2:
		minString := elems[0]
		min, err := strconv.ParseInt(strings.TrimSpace(minString), 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v, timeString = %v)", minString, timeString)
			return 0
		}
		secString := elems[1]
		sec, err := strconv.ParseInt(strings.TrimSpace(secString), 10, 64)
		if err != nil {
			log.Printf("can not parse sec string (secString = %v, timeString = %v)", secString, timeString)
			return 0
		}
		return sec + (min * 60)
	case 3:
		hourString := elems[0]
		hour, err := strconv.ParseInt(strings.TrimSpace(hourString), 10, 64)
		if err != nil {
			log.Printf("can not parse hour string (hourString = %v, timeString = %v)", hourString, timeString)
			return 0
		}
		minString := elems[1]
		min, err := strconv.ParseInt(strings.TrimSpace(minString), 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v, timeString = %v)", minString, timeString)
			return 0
		}
		secString := elems[2]
		sec, err := strconv.ParseInt(strings.TrimSpace(secString), 10, 64)
		if err != nil {
			log.Printf("can not parse sec string (secString = %v, timeString = %v)", secString, timeString)
			return 0
		}
		return sec + (min * 60) + (hour * 3600)
	default:
		log.Printf("can not parse tims string (timeString = %v, timeString = %v)", timeString)
		return 0
	}
}

func (b *Builder)durationStringToSeconds(durationString string) (int64) {
	// parse 1h2m3s, 1h2m 1h3s, 1m2s, 1s
	var seconds int64
	elems := strings.SplitN(durationString, "h", 2)
	if len(elems) == 2 {
		hourString := elems[0]
		hour, err := strconv.ParseInt(strings.TrimSpace(hourString), 10, 64)
		if err != nil {
			log.Printf("can not parse hour string (hourString = %v, durationString = %v)", hourString, durationString)
			return 0
		}
		seconds += hour * 3600
		durationString = elems[1]
	}
	elems = strings.SplitN(durationString, "m", 2)
	if len(elems) == 2 {
		minString := elems[0]
		min, err := strconv.ParseInt(strings.TrimSpace(minString), 10, 64)
		if err != nil {
			log.Printf("can not parse min string (minString = %v, durationString = %v)", minString, durationString)
			return 0
		}
		seconds += min * 60
		durationString = elems[1]
	}
	elems = strings.SplitN(durationString, "s", 2)
	if len(elems) == 2 {
		secString := elems[0]
		sec, err := strconv.ParseInt(strings.TrimSpace(secString), 10, 64)
		if err != nil {
			log.Printf("can not parse sec string (minString = %v, durationString = %v)", secString, durationString)
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

func (b *Builder)makeChannelProperty(channel *database.Channel) (*channelProperty, error) {
	comments, err := b.databaseOperator.GetAllCommentsByChannelIdAndLikeText(channel.ChannelId, "%:%")
	if err != nil {
		return nil, errors.Wrapf(err, "can not get comments from database (channelId = %v)", channel.ChannelId)
	}
	channelProp := &channelProperty{
		videos: make(map[string]*videoProperty),
	}
	for _, comment := range comments {
		videoProp, ok := channelProp.videos[comment.VideoId]
		if !ok {
			videoProp = &videoProperty{
				timeRangeList: make([]*timeRangeProperty, 0),
			}
			channelProp.videos[comment.VideoId] = videoProp
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
	return channelProp, nil
}

type Clip struct {
	VideoId string
	Start   int64
	End     int64
	Recommenders []string
}

func (b *Builder)buildPageProp(channel *database.Channel) ([]*Clip, error) {
	clipProp, err := b.makeChannelProperty(channel)
	if err != nil {
		return nil, errors.Wrapf(err, "can not make clip property (channelId = %v)", channel.ChannelId)
	}
	pageProp := make([]*Clip, 0)
	for videoId, videoProp := range clipProp.videos {
		for _, timeRangeProp := range videoProp.timeRangeList {
			authors := make([]string, 0, len(timeRangeProp.comments))
			for commentProp, _ := range timeRangeProp.comments {
				authors = append(authors, commentProp.author)
			}
			sort.Slice(authors, func(i, j int) bool {
				return authors[i] < authors[j]
			})
			clip := &Clip {
				VideoId: videoId,
				Start: timeRangeProp.start,
				End: timeRangeProp.end,
				Recommenders: authors,
			}
			pageProp = append(pageProp, clip)
		}
	}
	return pageProp, nil
}

func (b *Builder)Build() (error) {
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
	// create index.html
	indexHtmlPath := filepath.Join(b.buildDirPath, "index.html")
	indexHtmlFile, err := os.OpenFile(indexHtmlPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "can not open index html file (path = %v)", indexHtmlPath)
	}
	defer indexHtmlFile.Close()
	err = b.templates.ExecuteTemplate(indexHtmlFile, "index.tmpl", dbChannels)
	if err != nil {
		return errors.Wrapf(err, "can not write to index html file (path = %v)", indexHtmlPath)
	}
	// create channel page
	for _, dbChannel := range dbChannels {
		pageHtmlPath := filepath.Join(b.buildDirPath, dbChannel.ChannelId + ".html")
		pageHtmlFile, err := os.OpenFile(pageHtmlPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrapf(err, "can not open page html file (path = %v)", pageHtmlPath)
		}
		defer pageHtmlFile.Close()
		err = b.templates.ExecuteTemplate(pageHtmlFile, "page.tmpl", dbChannel)
		if err != nil {
			return errors.Wrapf(err, "can not write to page html file (path = %v)", pageHtmlPath)
		}
	}
	// create page prop
	for _, dbChannel := range dbChannels {
		lastChannelPage, ok, err := b.databaseOperator.GetChannelPageByChannelId(dbChannel.ChannelId)
                if err != nil {
                        return  errors.Wrapf(err, "can not get channel page from database (channelId = %v)", dbChannel.ChannelId)
                }
		pageProp, err := b.buildPageProp(dbChannel)
		if err != nil {
			return errors.Wrapf(err, "can not get build page (channelId = %v)", dbChannel.ChannelId)
		}
		jsonBytes, err := json.Marshal(pageProp)
		if err != nil {
			return errors.Wrapf(err, "can not marshal json (channelId = %v)", dbChannel.ChannelId)
		}
		newPageHash := fmt.Sprintf("%x", sha1.Sum(jsonBytes))
		if ok && lastChannelPage.PageHash == newPageHash {
			log.Printf("skip because same page hash (oldPageHash = %v, newPageHash = %v", lastChannelPage.PageHash, newPageHash)
			continue
		}
		jsonPath := filepath.Join(b.buildJsonDirPath, dbChannel.ChannelId + ".json")
		err = ioutil.WriteFile(jsonPath, jsonBytes, 0644)
		if err != nil {
			return errors.Wrapf(err, "can not write json to file (channelId = %v, path = %v)", dbChannel.ChannelId, jsonPath)
		}
		err = b.databaseOperator.UpdatePageHashAndDirtyOfChannelPage(dbChannel.ChannelId, newPageHash, 1)
                if err != nil {
                        return  errors.Wrapf(err, "can not update page hash and dirty of channelPage (channelId = %v, newPageHash = %v)", dbChannel.ChannelId, newPageHash)
                }
	}
	return nil
}

func NewBuilder(sourceDirPath string, buildDirPath string, maxDuration int64, adjustStartTimeSpan int64, channels youtubedataapi.Channels, databaseOperator *database.DatabaseOperator) (*Builder, error)  {
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
	templatePattern := filepath.Join(templateDirPath, "*.tmpl")
	templates, err := template.ParseGlob(templatePattern)
	if err != nil {
		return nil, errors.Wrapf(err, "can not parse templates (template pattern = %v)", templatePattern)
	}
        buildJsDirPath := filepath.Join(buildDirPath, "js")
         _, err = os.Stat(buildJsDirPath)
	if err != nil {
                err := os.MkdirAll(buildJsDirPath, 0755)
                if err != nil {
                        return nil, errors.Wrapf(err, "can not create directory (%v)", buildJsDirPath)
                }
        }
        buildJsonDirPath := filepath.Join(buildDirPath, "json")
         _, err = os.Stat(buildJsonDirPath)
	if err != nil {
                err := os.MkdirAll(buildJsonDirPath, 0755)
                if err != nil {
                        return nil, errors.Wrapf(err, "can not create directory (%v)", buildJsonDirPath)
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
		buildDirPath: buildDirPath,
		buildJsDirPath: buildJsDirPath,
		buildJsonDirPath: buildJsonDirPath,
		channels: channels,
		databaseOperator: databaseOperator,
		timeRangeRegexp: timeRangeRegexp,
		startEndSepRegexp: startEndSepRegexp,
		maxDuration: maxDuration,
		adjustStartTimeSpan: adjustStartTimeSpan,
		templates: templates,
	}, nil
}

