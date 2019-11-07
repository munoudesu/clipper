package youtubedataapi

type YtInitialData struct {
	ContinuationContents struct {
		LiveChatContinuation struct {
			Actions []struct {
				ReplayChatItemAction struct {
					Actions []struct {
						AddChatItemAction struct {
							ClientID string `json:"clientId"`
							Item     struct {
								LiveChatPaidMessageRenderer struct {
									AuthorExternalChannelID string `json:"authorExternalChannelId"`
									AuthorName              struct {
										SimpleText string `json:"simpleText"`
									} `json:"authorName"`
									AuthorNameTextColor int64 `json:"authorNameTextColor"`
									AuthorPhoto         struct {
										Thumbnails []struct {
											Height int64  `json:"height"`
											URL    string `json:"url"`
											Width  int64  `json:"width"`
										} `json:"thumbnails"`
									} `json:"authorPhoto"`
									BodyBackgroundColor      int64 `json:"bodyBackgroundColor"`
									BodyTextColor            int64 `json:"bodyTextColor"`
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
									HeaderBackgroundColor int64  `json:"headerBackgroundColor"`
									HeaderTextColor       int64  `json:"headerTextColor"`
									ID                    string `json:"id"`
									Message               struct {
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
									PurchaseAmountText struct {
										SimpleText string `json:"simpleText"`
									} `json:"purchaseAmountText"`
									TimestampColor int64 `json:"timestampColor"`
									TimestampText  struct {
										SimpleText string `json:"simpleText"`
									} `json:"timestampText"`
									TimestampUsec  string `json:"timestampUsec"`
									TrackingParams string `json:"trackingParams"`
								} `json:"liveChatPaidMessageRenderer"`
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
						AddLiveChatTickerItemAction struct {
							DurationSec string `json:"durationSec"`
							Item        struct {
								LiveChatTickerPaidMessageItemRenderer struct {
									Amount struct {
										SimpleText string `json:"simpleText"`
									} `json:"amount"`
									AmountTextColor         int64  `json:"amountTextColor"`
									AuthorExternalChannelID string `json:"authorExternalChannelId"`
									AuthorPhoto             struct {
										Thumbnails []struct {
											Height int64  `json:"height"`
											URL    string `json:"url"`
											Width  int64  `json:"width"`
										} `json:"thumbnails"`
									} `json:"authorPhoto"`
									DurationSec        int64  `json:"durationSec"`
									EndBackgroundColor int64  `json:"endBackgroundColor"`
									FullDurationSec    int64  `json:"fullDurationSec"`
									ID                 string `json:"id"`
									ShowItemEndpoint   struct {
										ClickTrackingParams string `json:"clickTrackingParams"`
										CommandMetadata     struct {
											WebCommandMetadata struct {
												IgnoreNavigation bool `json:"ignoreNavigation"`
											} `json:"webCommandMetadata"`
										} `json:"commandMetadata"`
										ShowLiveChatItemEndpoint struct {
											Renderer struct {
												LiveChatPaidMessageRenderer struct {
													AuthorExternalChannelID string `json:"authorExternalChannelId"`
													AuthorName              struct {
														SimpleText string `json:"simpleText"`
													} `json:"authorName"`
													AuthorNameTextColor int64 `json:"authorNameTextColor"`
													AuthorPhoto         struct {
														Thumbnails []struct {
															Height int64  `json:"height"`
															URL    string `json:"url"`
															Width  int64  `json:"width"`
														} `json:"thumbnails"`
													} `json:"authorPhoto"`
													BodyBackgroundColor      int64 `json:"bodyBackgroundColor"`
													BodyTextColor            int64 `json:"bodyTextColor"`
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
													HeaderBackgroundColor int64  `json:"headerBackgroundColor"`
													HeaderTextColor       int64  `json:"headerTextColor"`
													ID                    string `json:"id"`
													Message               struct {
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
													PurchaseAmountText struct {
														SimpleText string `json:"simpleText"`
													} `json:"purchaseAmountText"`
													TimestampColor int64 `json:"timestampColor"`
													TimestampText  struct {
														SimpleText string `json:"simpleText"`
													} `json:"timestampText"`
													TimestampUsec  string `json:"timestampUsec"`
													TrackingParams string `json:"trackingParams"`
												} `json:"liveChatPaidMessageRenderer"`
											} `json:"renderer"`
										} `json:"showLiveChatItemEndpoint"`
									} `json:"showItemEndpoint"`
									StartBackgroundColor int64  `json:"startBackgroundColor"`
									TrackingParams       string `json:"trackingParams"`
								} `json:"liveChatTickerPaidMessageItemRenderer"`
							} `json:"item"`
						} `json:"addLiveChatTickerItemAction"`
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
			ViewerName     string `json:"viewerName"`
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
				Csn                string `json:"csn"`
				DelegatedSessionID string `json:"delegatedSessionId"`
				SessionIndex       int64  `json:"sessionIndex"`
				VisitorData        string `json:"visitorData"`
			} `json:"ytConfigData"`
		} `json:"webResponseContextExtensionData"`
	} `json:"responseContext"`
	TrackingParams string `json:"trackingParams"`
}
