package youtubedataapi


import (
	"time"
	"strconv"
	"encoding/json"
	"strings"
	"log"
	"context"
	"bytes"
	"net/http"
	"io/ioutil"
	"github.com/pkg/errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/munoudesu/clipper/database"
)

const(
	userAgent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:70.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.120 Safari/537.36 Gecko/20100101 Firefox/70.0"
)

type LiveCharCollector struct {
	channelId        string
	videoId          string
	scraping         bool
	liveChatComments []*database.LiveChatComment
	maxRetry         int
	databaseOperator *database.DatabaseOperator
	verbose          bool
}

func (l *LiveCharCollector)timestampUsecToISO8601(timestampUsec string) (string) {
        t := time.Time{}
        ts, err := strconv.ParseInt(timestampUsec, 10, 64)
        if err == nil  {
                sec := ts / 1000000
                nsec := (ts % 1000000) * 1000
                t = time.Unix(sec, nsec)
        }
        return t.UTC().Format("2006-01-02T15:04:05.000Z")
}


func (l *LiveCharCollector)getPage(url string, useUserAgent bool) ([]byte, error) {
	if l.verbose {
		log.Printf("retrive url = %v", url)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "can not create http request (url = %v)", url)
	}
	if useUserAgent {
		req.Header.Set("User-Agent", userAgent)
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "can not request of http (url = %v)", url)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("response have unexpected status (url = %v, status = %v)", url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "can not read body (url = %v)", url)
	}
	return body, nil
}

func (l *LiveCharCollector)getFirstLiveChatReplayUrl() (string, error) {
	url := "https://www.youtube.com/watch?v=" + l.videoId
        body, err := l.getPage(url, false)
        if err != nil {
                return "", errors.Wrapf(err, "can not get video page (url = %v)", url)
        }
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", errors.Wrapf(err, "can not parse body (url = %v)", url)
	}
	var firstLiveChatReplayUrl string
	doc.Find("#live-chat-iframe").Each(func(i int, s *goquery.Selection) {
		url, ok := s.Attr("src")
		if !ok {
			return
		}
		firstLiveChatReplayUrl = url
	})
	return firstLiveChatReplayUrl, nil
}

func (l *LiveCharCollector)getYtInitialData(url string)(string, error) {
        body, err := l.getPage(url, true)
        if err != nil {
                return "", errors.Wrapf(err, "can not get video page (url = %v)", url)
        }
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", errors.Wrapf(err, "can not parse body (url = %v)", url)
	}
	var yuInitialDataStr string
	doc.Find("body>script").EachWithBreak(func(i int, s *goquery.Selection) (bool) {
		html := s.Text()
		if strings.Contains(html, "ytInitialData") {
			elems := strings.SplitN(html, "=", 2)
			if len(elems) < 2 {
				log.Printf("can not not parse ytInitialData (url = %v, html = %v)", url, html)
				return true
			}
			yuInitialDataStr = strings.TrimSuffix(strings.TrimSpace(elems[1]), ";")
			return false
		}
		return true
	})
	if yuInitialDataStr == "" {
		return "", errors.Wrapf(err, "not found ytInitialData (url = %v)", url)
	}
	return yuInitialDataStr, nil
}

func (l *LiveCharCollector)getYtInitialDataWithScraping(url string)(string, error) {
	if l.verbose {
		log.Printf("live chat replay url = %v", url)
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:], )
	ctx1, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel1()
	ctx2, cancel2 := chromedp.NewContext(ctx1, chromedp.WithLogf(log.Printf))
	defer cancel2()
	ctx3, cancel3 := context.WithTimeout(ctx2, 60 * time.Second)
	defer cancel3()
	var scripts []*cdp.Node
	var yuInitialDataStr string
	err := chromedp.Run(ctx3, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Nodes(`body>script`, &scripts, chromedp.ByQueryAll),
		chromedp.WaitVisible("#contents", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) (error) {
			for _, script := range scripts {
				if script.ChildNodeCount == 0 {
					continue
				}
				if !strings.Contains(script.Children[0].NodeValue, "ytInitialData") {
					continue
				}
				html, err := dom.GetOuterHTML().WithNodeID(script.Children[0].NodeID).Do(ctx)
				if err != nil {
					return errors.Wrapf(err, "can not get outer html (url = %v)", url)
				}
				elems := strings.SplitN(html, "=", 2)
				if len(elems) < 2 {
					return errors.Errorf("can not not parse ytInitialData (url = %v, html = %v)", url, html)
				}
				yuInitialDataStr = strings.TrimSuffix(strings.TrimSpace(elems[1]), ";")
				return nil
			}
			return errors.Errorf("not found ytInitialData (url = %v)", url)
		}),
	});
	if err != nil {
		return "", errors.Wrapf(err, "can not navigate (url = %v)", url)
	}
	return yuInitialDataStr, nil
}

func (l *LiveCharCollector)getLiveChat(url string)(string, error) {
	var yuInitialDataStr string
	if l.scraping {
		str, err := l.getYtInitialDataWithScraping(url)
		if err != nil {
			return "", errors.Wrapf(err, "can not get ytInitialData with scraping (url = %v)", url)
		}
		yuInitialDataStr = str
	} else {
		str, err := l.getYtInitialData(url)
		if err != nil {
			return "", errors.Wrapf(err, "can not get ytInitialData (url = %v)", url)
		}
		yuInitialDataStr = str
	}
	if yuInitialDataStr == "" {
		return "", errors.Errorf("not found ytInitialData (url = %v)", url)
	}
	var ytInitialData YtInitialData
	err := json.Unmarshal([]byte(yuInitialDataStr), &ytInitialData)
	if err != nil {
		return "", errors.Wrapf(err, "can not unmarshal ytInitialData (url = %v, yuInitialDataStr = %v)", url, yuInitialDataStr)
	}
	var nextId string
	if len(ytInitialData.ContinuationContents.LiveChatContinuation.Continuations) >= 2 {
		nextId = ytInitialData.ContinuationContents.LiveChatContinuation.Continuations[0].LiveChatReplayContinuationData.Continuation
	}
	if l.verbose {
		log.Printf("nextId = %v", nextId)
	}
	for _, action1 := range ytInitialData.ContinuationContents.LiveChatContinuation.Actions {
		videoOffsetTimeMsec := action1.ReplayChatItemAction.VideoOffsetTimeMsec
		for _, action2 := range action1.ReplayChatItemAction.Actions {
			clientId := action2.AddChatItemAction.ClientID
			if action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.TimestampText.SimpleText != "" {
				timestampText := action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.TimestampText.SimpleText
				authorName := action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.AuthorName.SimpleText
				var authorPhotoUrl string
				if len(action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.AuthorPhoto.Thumbnails) > 0 {
					authorPhotoUrl = action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.AuthorPhoto.Thumbnails[0].URL
				}
				id := action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.ID
				var messageText string
				for _, run := range action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.Message.Runs {
					messageText += run.Text
				}
				purchaseAmountText := action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.PurchaseAmountText.SimpleText
				timestampUsec := action2.AddChatItemAction.Item.LiveChatPaidMessageRenderer.TimestampUsec
				liveChatComment := &database.LiveChatComment{
					UniqueId: l.videoId + ".paid." + id + "." + timestampUsec + "." + clientId,
					ChannelId: l.channelId,
					VideoId: l.videoId,
					ClientId: clientId,
					MessageId: id,
					TimestampAt: l.timestampUsecToISO8601(timestampUsec),
					TimestampText: timestampText,
					AuthorName: authorName,
					AuthorPhotoUrl: authorPhotoUrl,
					MessageText: messageText,
					PurchaseAmountText: purchaseAmountText,
					VideoOffsetTimeMsec: videoOffsetTimeMsec,
				}
				l.liveChatComments = append(l.liveChatComments, liveChatComment)
			} else if action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampText.SimpleText != "" {
				timestampText := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampText.SimpleText
				authorName := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.AuthorName.SimpleText
				var authorPhotoUrl string
				if len(action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.AuthorPhoto.Thumbnails) > 0 {
					authorPhotoUrl = action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.AuthorPhoto.Thumbnails[0].URL
				}
				id := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.ID
				var messageText string
				for _, run := range action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.Message.Runs {
					messageText += run.Text
				}
				timestampUsec := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampUsec
				liveChatComment := &database.LiveChatComment{
					UniqueId: l.videoId + ".text." + id + "." + timestampUsec + "." + clientId,
					ChannelId: l.channelId,
					VideoId: l.videoId,
					ClientId: clientId,
					MessageId: id,
					TimestampAt: l.timestampUsecToISO8601(timestampUsec),
					TimestampText: timestampText,
					AuthorName: authorName,
					AuthorPhotoUrl: authorPhotoUrl,
					MessageText: messageText,
					PurchaseAmountText: "",
					VideoOffsetTimeMsec: videoOffsetTimeMsec,
				}
				l.liveChatComments = append(l.liveChatComments, liveChatComment)
			}
		}
	}
	if nextId == "" {
		return "", nil
	}
	return "https://www.youtube.com/live_chat_replay?continuation=" + nextId, nil
}

func (l *LiveCharCollector)Collect() (error) {
	liveChatComments, err := l.databaseOperator.GetLiveChatCommentsByVideoId(l.videoId)
	if err != nil {
		return errors.Wrapf(err, "can not get live chat from database (videoId = %v)", l.videoId)
	}
	if len(liveChatComments) > 0 {
		if l.verbose {
			log.Printf("already exists live chat in database (videoId = %v)", l.videoId)
		}
		return nil
	}
	firstLiveChatReplayUrl, err := l.getFirstLiveChatReplayUrl()
	if err != nil {
		return errors.Wrapf(err, "can not get first live chat replay url (videoId = %v)", l.videoId)
	}
	if l.verbose {
		log.Printf("first live chat replay url = %v", firstLiveChatReplayUrl)
	}
	if firstLiveChatReplayUrl == "" {
		if l.verbose {
			log.Printf("skip collect live chat because can not get first live chat replay url")
		}
		return nil
	}
	nextUrl := firstLiveChatReplayUrl
	for {
		retry := 0
		for {
			url, err := l.getLiveChat(nextUrl)
			if err == nil {
				nextUrl = url
				break
			}
			retry += 1
			if retry < l.maxRetry {
				log.Printf("can not get live chat (videoId = %v), retry ...: %v", l.videoId, err)
				time.Sleep(time.Second)
				continue
			} else {
				return errors.Wrapf(err, "can not get live chat (videoId = %v)", l.videoId)
			}
		}
		if nextUrl == "" {
			break
		}
	}
	err = l.databaseOperator.UpdateLiveChatComments(l.liveChatComments)
	if err != nil {
		return errors.Wrapf(err, "can not update live chat (videoId = %v)", l.videoId)
	}
	return nil
}

func NewLiveChatCollector(video *database.Video, scraping bool, databaseOperator *database.DatabaseOperator, verbose bool) (*LiveCharCollector) {
	return &LiveCharCollector {
		channelId:        video.ChannelId,
		videoId:          video.VideoId,
		scraping:         scraping,
		liveChatComments: make([]*database.LiveChatComment, 0, 1000),
		maxRetry:         10,
		databaseOperator: databaseOperator,
		verbose: verbose,
	}
}
