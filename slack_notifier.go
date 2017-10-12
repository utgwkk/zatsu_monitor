package main

import (
	"fmt"
	"github.com/nlopes/slack"
)

var slackExpectedKeys = []string{"type", "check_url", "api_token", "channel"}

// SlackNotifier represents notifier for Slack
type SlackNotifier struct {
	apiToken string
	userName string
	channel  string
}

// NewSlackNotifier create new SlackNotifier instance
func NewSlackNotifier(apiToken string, userName string, channel string) *SlackNotifier {
	s := new(SlackNotifier)
	s.apiToken = apiToken
	s.channel = channel

	if len(userName) == 0 {
		s.userName = "zatsu_monitor"
	} else {
		s.userName = userName
	}

	return s
}

// ExpectedKeys returns expected keys for SlackNotifier
func (s *SlackNotifier) ExpectedKeys() []string {
	return slackExpectedKeys
}

// PostStatus perform posting current status for URL
func (s *SlackNotifier) PostStatus(param *PostStatusParam) error {
	var statusText, iconEmoji, userName string

	successful := IsSuccessfulStatus(param.CurrentStatusCode)

	if successful {
		statusText = "ok"
		iconEmoji = ":green_heart:"
		userName = s.userName + " Successful"
	} else {
		statusText = "down"
		iconEmoji = ":broken_heart:"
		userName = s.userName + " Failure"
	}

	format := `%s is %s
statusCode: %d -> %d
responseTime: %f sec`
	message := fmt.Sprintf(format, param.CheckUrl, statusText, param.BeforeStatusCode, param.CurrentStatusCode, param.ResponseTime)

	if param.HttpError != nil {
		message += fmt.Sprintf("\nhttpError: %v", param.HttpError)
	}

	params := slack.NewPostMessageParameters()
	params.Username = userName
	params.IconEmoji = iconEmoji

	api := slack.New(s.apiToken)

	_, _, err := api.PostMessage(s.channel, message, params)

	return err
}
