package slackbot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alexbrazier/slack"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/kjwardy/go-url/api/model"
	cache "github.com/patrickmn/go-cache"
)

var (
	urlModel             = &model.URL{}
	getKeyFromTextRegexp = regexp.MustCompile("\\bgo ([\\w-\\/]+)\\b")
)

// SlackBot .
type SlackBot struct {
	Cache  *cache.Cache
	Client *slack.Client
	RTM    *slack.RTM
	Logger *logging.Logger
}

func (s *SlackBot) log(action, level, message, result, operation, errorType string) {
	if s.Logger == nil {
		return
	}
	fields := map[string]interface{}{}
	if operation != "" {
		fields["slack.operation"] = operation
	}
	if errorType != "" {
		fields["error.type"] = errorType
	}
	s.Logger.Event(level, message, action, result, fields)
}

func (s *SlackBot) logTask(task, message string, err error) {
	if err != nil && s.Logger != nil {
		s.Logger.Event("ERROR", message, "background.task", "failure", map[string]interface{}{
			"background.task": task,
			"error.type":      "database_error",
		})
	}
}

func (s *SlackBot) getKeyFromText(text string) string {
	matches := getKeyFromTextRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}

func (s *SlackBot) deleteMessage(channel, timestamp string) {
	cachedTimestamp, found := s.Cache.Get(timestamp)
	if !found {
		s.log("slack.bot", "INFO", "Slack message is no longer cached", "success", "delete_message", "")
		return
	}
	s.RTM.DeleteMessage(channel, cachedTimestamp.(string))
}

// Init initialises slackbot
func (s *SlackBot) Init() {
	appConfig := config.GetConfig()
	s.log("slack.bot", "INFO", "Slack bot started", "success", "start", "")
	s.Cache = cache.New(60*time.Minute, 30*time.Minute)
	s.Client = slack.New(appConfig.Slack.Token)
	if appConfig.Debug && !appConfig.JSONLogs {
		s.Client.SetDebug(true)
	}

	s.RTM = s.Client.NewRTM()
	go s.RTM.ManageConnection()
	if s.Logger != nil {
		s.Logger.Event("INFO", "Slack bot event processor started", "background.task", "success", map[string]interface{}{
			"background.task": "slackbot_event_processor",
		})
	}

	for msg := range s.RTM.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// Stop the bots from taking over
			if ev.SubType == "bot_message" {
				break
			}

			key := s.getKeyFromText(ev.Text)
			edited := false
			if ev.SubType == "message_changed" {
				key = s.getKeyFromText(ev.SubMessage.Text)
				oldKey := s.getKeyFromText(ev.PreviousMessage.Text)
				// key hasn't changed
				if oldKey == key {
					break
				}
				if oldKey != "" && key == "" {
					s.deleteMessage(ev.Channel, ev.PreviousMessage.Timestamp)
					break
				}
				// Edit message if previous one exists
				if oldKey != "" {
					urls, err := urlModel.GetUrlsFromKeys([]string{oldKey})
					if err != nil {
						s.logTask("slackbot_lookup_old_key", "Slack bot URL lookup failed", err)
						break
					}
					if len(urls) > 0 {
						edited = true
					}
				}
			}

			if ev.SubType == "message_deleted" {
				s.deleteMessage(ev.Channel, ev.PreviousMessage.Timestamp)
				break
			}

			if key == "" {
				break
			}

			urls, err := urlModel.GetUrlsFromKeys([]string{key})
			if err != nil {
				s.logTask("slackbot_lookup_url", "Slack bot URL lookup failed", err)
				break
			}
			if len(urls) == 0 {
				if ev.PreviousMessage != nil {
					s.deleteMessage(ev.Channel, ev.PreviousMessage.Timestamp)
				}
				break
			}
			url := urls[0]
			appURI := appConfig.AppURI

			extraText := url.URL
			if url.URL == "" {
				aliasUrls, err := urlModel.GetUrlsFromKeys(url.Alias)
				if err != nil {
					s.logTask("slackbot_lookup_aliases", "Slack bot alias lookup failed", err)
					break
				}
				urlStrings := make([]string, len(aliasUrls))
				for i, url := range aliasUrls {
					urlStrings[i] = url.URL
				}
				extraText = strings.Join(urlStrings, "\n")
			}

			attachments := []slack.Attachment{
				{
					Footer: extraText,
				},
			}

			message := slack.PostMessageParameters{
				Attachments:     attachments,
				ThreadTimestamp: ev.ThreadTimestamp,
			}
			if message.ThreadTimestamp == "" && ev.PreviousMessage != nil {
				message.ThreadTimestamp = ev.PreviousMessage.ThreadTimestamp
			}
			messageText := fmt.Sprintf("%s/%s", appURI, key)

			if edited {
				cachedTimestamp, found := s.Cache.Get(ev.PreviousMessage.Timestamp)
				if !found {
					s.log("slack.bot", "INFO", "Slack message is no longer cached", "success", "update_message", "")
					break
				}
				_, _, _, err := s.Client.SendMessage(ev.Channel, slack.MsgOptionUpdate(cachedTimestamp.(string)), slack.MsgOptionText(messageText, true), slack.MsgOptionPostMessageParameters(message))
				if err != nil {
					s.log("slack.bot", "ERROR", "Slack message update failed", "failure", "update_message", "slack_api_error")
					break
				}
				s.log("slack.bot", "INFO", "Slack message updated", "success", "update_message", "")
			} else {
				_, timestamp, err := s.RTM.PostMessage(ev.Channel, messageText, message)
				if err != nil {
					s.log("slack.bot", "ERROR", "Slack message post failed", "failure", "post_message", "slack_api_error")
					break
				}
				s.Cache.Set(ev.Timestamp, timestamp, cache.DefaultExpiration)
				if ev.PreviousMessage != nil {
					s.Cache.Set(ev.PreviousMessage.Timestamp, timestamp, cache.DefaultExpiration)
				}
				s.log("slack.bot", "INFO", "Slack message posted", "success", "post_message", "")
			}

		case *slack.RTMError:
			s.log("slack.bot", "ERROR", "Slack bot connection error", "failure", "connection", "slack_rtm_error")

		case *slack.InvalidAuthEvent:
			s.log("slack.bot", "ERROR", "Slack bot authentication failed", "failure", "connection", "invalid_slack_credentials")
			break

		default:
			// Do nothing
		}
	}
	if s.Logger != nil {
		s.Logger.Event("INFO", "Slack bot event processor stopped", "background.task", "success", map[string]interface{}{
			"background.task": "slackbot_event_processor",
		})
	}
}
