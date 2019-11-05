package youtubedataapi


import (
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

type LiveCharCollector struct {
	videoId          string
	liveChatComments []*database.LiveChatComment
	databaseOperator   *database.DatabaseOperator
	verbose            bool
}

func (l *LiveCharCollector)getVideoPage() ([]byte, error) {
	url := "https://www.youtube.com/watch?v=" + l.videoId
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "can not create http request (url = %v)", url)
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

func (l *LiveCharCollector)getFirstLiveChatReplayUrl(body []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", errors.Wrapf(err, "can not parse body (videoId = %v)", l.videoId)
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

func (l *LiveCharCollector)getLiveChat(url string)(string, error) {
	if l.verbose {
		log.Printf("live chat replay url = %v", url)
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:], )
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	var scripts []*cdp.Node
	var html string
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Nodes(`body>script`, &scripts, chromedp.ByQueryAll),
		chromedp.ActionFunc(func(ctx context.Context) (error) {
			for _, script := range scripts {
				if script.ChildNodeCount == 0 {
					continue
				}
				if strings.Contains(script.Children[0].NodeValue, "ytInitialData") {
					var err error
					html, err = dom.GetOuterHTML().WithNodeID(script.Children[0].NodeID).Do(ctx)
					if err != nil {
						return errors.Wrapf(err, "can not get outer html (url = %v)", url)
					}
					return nil
				}
			}
			return errors.Errorf("not found ytInitialData (url = %v)", url)
		}),
	});
	if err != nil {
		return "", errors.Wrapf(err, "can not navigate (url = %v)", url)
	}
	strs := strings.SplitN(html, "=", 2)
	if len(strs) < 2 {
		return "", errors.Errorf("not found ytInitialData (url = %v)", url)

	}
	yuInitialDataStr := strings.TrimSuffix(strings.TrimSpace(strs[1]), ";")
	var ytInitialData YtInitialData
	err = json.Unmarshal([]byte(yuInitialDataStr), &ytInitialData)
	if err != nil {
		return "", errors.Wrapf(err, "can not parse ytInitialData (url = %v)", url)
	}
	var nextId string
	if len(ytInitialData.ContinuationContents.LiveChatContinuation.Continuations) >= 2 {
		nextId = ytInitialData.ContinuationContents.LiveChatContinuation.Continuations[0].LiveChatReplayContinuationData.Continuation
	}
	for _, action1 := range ytInitialData.ContinuationContents.LiveChatContinuation.Actions {
		for _, action2 := range action1.ReplayChatItemAction.Actions {
			clientId := action2.AddChatItemAction.ClientID
			id := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.ID
			timestampUsec := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampUsec
			timestampText := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampText.SimpleText
			authorName := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.AuthorName.SimpleText
			for _, run := range action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.Message.Runs {
				liveChatComment := &database.LiveChatComment{
					UniqueId: l.videoId + "." + clientId + "." + id + "." + timestampUsec,
					VideoId: l.videoId,
					ClientId: clientId,
					ChatId: id,
					TimestampUsec: timestampUsec,
					TimestampText: timestampText,
					AuthorName: authorName,
					Text: run.Text,
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
	body, err := l.getVideoPage()
	if err != nil {
		return errors.Wrapf(err, "can not get video page (videoId = %v)", l.videoId)
	}
	firstLiveChatReplayUrl, err := l.getFirstLiveChatReplayUrl(body)
	if err != nil {
		return errors.Wrapf(err, "can not get first live char replay url (videoId = %v)", l.videoId)
	}
	if l.verbose {
		log.Printf("first live chat replay url = %v", firstLiveChatReplayUrl)
	}
	nextUrl := firstLiveChatReplayUrl
	for {
		nextUrl, err := l.getLiveChat(nextUrl)
		if err != nil {
			return errors.Wrapf(err, "can not get live chats (videoId = %v)", l.videoId)
		}
		if nextUrl == "" {
			break
		}
	}
	for _, liveChatComment := range l.liveChatComments {
		log.Printf("%v, %v, %v, %v, %v, %v, %v",
			liveChatComment.UniqueId,
			liveChatComment.VideoId,
			liveChatComment.ClientId,
			liveChatComment.ChatId,
			liveChatComment.AuthorName,
			liveChatComment.TimestampUsec,
			liveChatComment.TimestampText,
			liveChatComment.Text)
	}
	return nil
}

func NewLiveChatCollector(videoId string, databaseOperator *database.DatabaseOperator, verbose bool) (*LiveCharCollector) {
	return &LiveCharCollector {
		videoId:          videoId,
		liveChatComments: make([]*database.LiveChatComment, 0, 1000),
		databaseOperator: databaseOperator,
		verbose: verbose,
	}
}
