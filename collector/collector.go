package main


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
)

type LiveCharCollector struct {
	videoId string
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
	log.Printf("url = %v", url)
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
	nextId := ytInitialData.ContinuationContents.LiveChatContinuation.Continuations[0].LiveChatReplayContinuationData.Continuation
	log.Printf("nextId: %v", nextId)
	log.Printf("actions1: %v", len(ytInitialData.ContinuationContents.LiveChatContinuation.Actions))
	for _, action1 := range ytInitialData.ContinuationContents.LiveChatContinuation.Actions {
		log.Printf("actions2: %v", len(action1.ReplayChatItemAction.Actions))
		for _, action2 := range action1.ReplayChatItemAction.Actions {
			clientId := action2.AddChatItemAction.ClientID
			id := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.ID
			timestampUsec := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampUsec
			timestampText := action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.TimestampText.SimpleText
			for _, run := range action2.AddChatItemAction.Item.LiveChatTextMessageRenderer.Message.Runs {
				text := run.Text
				log.Printf("clientId = %v, id = %v, timestampUsec = %v, timestampText = %v, text = %v", clientId, id, timestampUsec, timestampText, text)
			}
		}
	}

	return "", nil
	//https://www.youtube.com/live_chat_replay?continuation=
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
	return nil
}

func NewLiveChatCollector(videoId string) (*LiveCharCollector) {
	return &LiveCharCollector {
		videoId: videoId,
	}
}

func main() {

	videoId := "mLXydlAHODA"
	liveChatCollector := NewLiveChatCollector(videoId)
	err := liveChatCollector.Collect()
	if err != nil {
		log.Printf("can not collect live char: %v", err)
	}

}

type YtInitialData struct {
	ContinuationContents struct {
		LiveChatContinuation struct {
			Actions []struct {
				ReplayChatItemAction struct {
					Actions []struct {
						AddChatItemAction struct {
							ClientID string `json:"clientId"`
							Item     struct {
								LiveChatTextMessageRenderer struct {
									AuthorBadges []struct {
										LiveChatAuthorBadgeRenderer struct {
											Accessibility struct {
												AccessibilityData struct {
													Label string `json:"label"`
												} `json:"accessibilityData"`
											} `json:"accessibility"`
											CustomThumbnail struct {
												Thumbnails []struct {
													URL string `json:"url"`
												} `json:"thumbnails"`
											} `json:"customThumbnail"`
											Tooltip string `json:"tooltip"`
										} `json:"liveChatAuthorBadgeRenderer"`
									} `json:"authorBadges"`
									AuthorExternalChannelID string `json:"authorExternalChannelId"`
									AuthorName              struct {
										SimpleText string `json:"simpleText"`
									} `json:"authorName"`
									AuthorPhoto struct {
										Thumbnails []struct {
											Height int64  `json:"height"`
											URL    string `json:"url"`
											Width  int64  `json:"width"`
										} `json:"thumbnails"`
									} `json:"authorPhoto"`
									ContextMenuAccessibility struct {
										AccessibilityData struct {
											Label string `json:"label"`
										} `json:"accessibilityData"`
									} `json:"contextMenuAccessibility"`
									ContextMenuEndpoint struct {
										ClickTrackingParams string `json:"clickTrackingParams"`
										CommandMetadata     struct {
											WebCommandMetadata struct {
												IgnoreNavigation bool `json:"ignoreNavigation"`
											} `json:"webCommandMetadata"`
										} `json:"commandMetadata"`
										LiveChatItemContextMenuEndpoint struct {
											Params string `json:"params"`
										} `json:"liveChatItemContextMenuEndpoint"`
									} `json:"contextMenuEndpoint"`
									ID      string `json:"id"`
									Message struct {
										Runs []struct {
											Emoji struct {
												EmojiID string `json:"emojiId"`
												Image   struct {
													Accessibility struct {
														AccessibilityData struct {
															Label string `json:"label"`
														} `json:"accessibilityData"`
													} `json:"accessibility"`
													Thumbnails []struct {
														Height int64  `json:"height"`
														URL    string `json:"url"`
														Width  int64  `json:"width"`
													} `json:"thumbnails"`
												} `json:"image"`
												IsCustomEmoji bool     `json:"isCustomEmoji"`
												SearchTerms   []string `json:"searchTerms"`
												Shortcuts     []string `json:"shortcuts"`
											} `json:"emoji"`
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"message"`
									TimestampText struct {
										SimpleText string `json:"simpleText"`
									} `json:"timestampText"`
									TimestampUsec string `json:"timestampUsec"`
								} `json:"liveChatTextMessageRenderer"`
								LiveChatViewerEngagementMessageRenderer struct {
									Icon struct {
										IconType string `json:"iconType"`
									} `json:"icon"`
									ID      string `json:"id"`
									Message struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"message"`
									TimestampUsec string `json:"timestampUsec"`
								} `json:"liveChatViewerEngagementMessageRenderer"`
							} `json:"item"`
						} `json:"addChatItemAction"`
					} `json:"actions"`
					VideoOffsetTimeMsec string `json:"videoOffsetTimeMsec"`
				} `json:"replayChatItemAction"`
			} `json:"actions"`
			ClientMessages struct {
				FatalError struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"fatalError"`
				GenericError struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"genericError"`
				ReconnectMessage struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"reconnectMessage"`
				ReconnectedMessage struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"reconnectedMessage"`
				UnableToReconnectMessage struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"unableToReconnectMessage"`
			} `json:"clientMessages"`
			Continuations []struct {
				LiveChatReplayContinuationData struct {
					ClickTrackingParams      string `json:"clickTrackingParams"`
					Continuation             string `json:"continuation"`
					TimeUntilLastMessageMsec int64  `json:"timeUntilLastMessageMsec"`
				} `json:"liveChatReplayContinuationData"`
				PlayerSeekContinuationData struct {
					ClickTrackingParams string `json:"clickTrackingParams"`
					Continuation        string `json:"continuation"`
				} `json:"playerSeekContinuationData"`
			} `json:"continuations"`
			Header struct {
				LiveChatHeaderRenderer struct {
					CollapseButton struct {
						ButtonRenderer struct {
							Accessibility struct {
								Label string `json:"label"`
							} `json:"accessibility"`
							IsDisabled     bool   `json:"isDisabled"`
							Size           string `json:"size"`
							Style          string `json:"style"`
							TrackingParams string `json:"trackingParams"`
						} `json:"buttonRenderer"`
					} `json:"collapseButton"`
					OverflowMenu struct {
						MenuRenderer struct {
							Accessibility struct {
								AccessibilityData struct {
									Label string `json:"label"`
								} `json:"accessibilityData"`
							} `json:"accessibility"`
							Items []struct {
								MenuNavigationItemRenderer struct {
									Icon struct {
										IconType string `json:"iconType"`
									} `json:"icon"`
									NavigationEndpoint struct {
										ClickTrackingParams string `json:"clickTrackingParams"`
										CommandMetadata     struct {
											WebCommandMetadata struct {
												IgnoreNavigation bool `json:"ignoreNavigation"`
											} `json:"webCommandMetadata"`
										} `json:"commandMetadata"`
										UserFeedbackEndpoint struct {
											BucketIdentifier string `json:"bucketIdentifier"`
											Hack             bool   `json:"hack"`
										} `json:"userFeedbackEndpoint"`
									} `json:"navigationEndpoint"`
									Text struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"text"`
									TrackingParams string `json:"trackingParams"`
								} `json:"menuNavigationItemRenderer"`
								MenuServiceItemRenderer struct {
									Icon struct {
										IconType string `json:"iconType"`
									} `json:"icon"`
									ServiceEndpoint struct {
										ToggleLiveChatTimestampsEndpoint struct {
											Hack bool `json:"hack"`
										} `json:"toggleLiveChatTimestampsEndpoint"`
									} `json:"serviceEndpoint"`
									Text struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"text"`
									TrackingParams string `json:"trackingParams"`
								} `json:"menuServiceItemRenderer"`
							} `json:"items"`
							TrackingParams string `json:"trackingParams"`
						} `json:"menuRenderer"`
					} `json:"overflowMenu"`
					ViewSelector struct {
						SortFilterSubMenuRenderer struct {
							Accessibility struct {
								AccessibilityData struct {
									Label string `json:"label"`
								} `json:"accessibilityData"`
							} `json:"accessibility"`
							SubMenuItems []struct {
								Accessibility struct {
									AccessibilityData struct {
										Label string `json:"label"`
									} `json:"accessibilityData"`
								} `json:"accessibility"`
								Continuation struct {
									ReloadContinuationData struct {
										ClickTrackingParams string `json:"clickTrackingParams"`
										Continuation        string `json:"continuation"`
									} `json:"reloadContinuationData"`
								} `json:"continuation"`
								Selected bool   `json:"selected"`
								Subtitle string `json:"subtitle"`
								Title    string `json:"title"`
							} `json:"subMenuItems"`
							TrackingParams string `json:"trackingParams"`
						} `json:"sortFilterSubMenuRenderer"`
					} `json:"viewSelector"`
				} `json:"liveChatHeaderRenderer"`
			} `json:"header"`
			IsReplay bool `json:"isReplay"`
			ItemList struct {
				LiveChatItemListRenderer struct {
					EnablePauseChatKeyboardShortcuts bool  `json:"enablePauseChatKeyboardShortcuts"`
					MaxItemsToDisplay                int64 `json:"maxItemsToDisplay"`
					MoreCommentsBelowButton          struct {
						ButtonRenderer struct {
							AccessibilityData struct {
								AccessibilityData struct {
									Label string `json:"label"`
								} `json:"accessibilityData"`
							} `json:"accessibilityData"`
							Icon struct {
								IconType string `json:"iconType"`
							} `json:"icon"`
							Style          string `json:"style"`
							TrackingParams string `json:"trackingParams"`
						} `json:"buttonRenderer"`
					} `json:"moreCommentsBelowButton"`
				} `json:"liveChatItemListRenderer"`
			} `json:"itemList"`
			PopoutMessage struct {
				MessageRenderer struct {
					Button struct {
						ButtonRenderer struct {
							ServiceEndpoint struct {
								PopoutLiveChatEndpoint struct {
									URL string `json:"url"`
								} `json:"popoutLiveChatEndpoint"`
							} `json:"serviceEndpoint"`
							Style string `json:"style"`
							Text  struct {
								Runs []struct {
									Text string `json:"text"`
								} `json:"runs"`
							} `json:"text"`
							TrackingParams string `json:"trackingParams"`
						} `json:"buttonRenderer"`
					} `json:"button"`
					Text struct {
						Runs []struct {
							Text string `json:"text"`
						} `json:"runs"`
					} `json:"text"`
					TrackingParams string `json:"trackingParams"`
				} `json:"messageRenderer"`
			} `json:"popoutMessage"`
			Ticker struct {
				LiveChatTickerRenderer struct {
					Sentinel bool `json:"sentinel"`
				} `json:"liveChatTickerRenderer"`
			} `json:"ticker"`
			TrackingParams string `json:"trackingParams"`
		} `json:"liveChatContinuation"`
	} `json:"continuationContents"`
	ResponseContext struct {
		ServiceTrackingParams []struct {
			Params []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"params"`
			Service string `json:"service"`
		} `json:"serviceTrackingParams"`
		WebResponseContextExtensionData struct {
			YtConfigData struct {
				Csn         string `json:"csn"`
				VisitorData string `json:"visitorData"`
			} `json:"ytConfigData"`
		} `json:"webResponseContextExtensionData"`
	} `json:"responseContext"`
	TrackingParams string `json:"trackingParams"`
}
