package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// Version represents app version (injected from ldflags)
	Version string

	// Revision represents app revision (injected from ldflags)
	Revision string
)

func main() {
	lambda.Start(lambdaHandler)
}

func lambdaHandler(ctx context.Context) {
	config, err := LoadConfigFromFile("./config.yml")

	if err != nil {
		panic(err)
	}

	for name, values := range config {
		perform(name, values)
	}
}

func printVersion() {
	fmt.Printf("zatsu_monitor %s, build %s\n", Version, Revision)
}

func perform(name string, values map[string]string) {
	notifierType := values["type"]

	var notifier Notifier

	switch notifierType {
	case "chatwork":
		notifier = NewChatworkNotifier(values["api_token"], values["room_id"])
	case "slack":
		notifier = NewSlackNotifier(values["api_token"], values["webhook_url"], values["user_name"], values["channel"])
	default:
		panic(fmt.Sprintf("Unknown type: %s in %s", notifierType, "./config.yml"))
	}

	// If it does not exist even one expected key, skip
	for _, expectedKey := range notifier.ExpectedKeys() {
		if _, ok := values[expectedKey]; !ok {
			return
		}
	}

	checkURL := values["check_url"]

	start := time.Now()
	currentStatusCode, httpError := GetStatusCode(checkURL)
	end := time.Now()
	responseTime := (end.Sub(start)).Seconds()

	fmt.Printf("time:%v\tcheck_url:%s\tstatus:%d\tresponse_time:%f\terror:%v\n", time.Now(), checkURL, currentStatusCode, responseTime, httpError)

	store := NewStatusStore("unused")
	beforeStatusCode, err := store.GetDbStatus(name)

	if err != nil {
		panic(err)
	}

	err = store.SaveDbStatus(name, currentStatusCode)

	if err != nil {
		panic(err)
	}

	onlyCheckOnTheOrderOf100 := false
	if values["check_only_top_of_status_code"] == "true" {
		onlyCheckOnTheOrderOf100 = true
	}

	if isNotify(beforeStatusCode, currentStatusCode, onlyCheckOnTheOrderOf100) {
		// When status code changes from the previous, notify
		param := PostStatusParam{
			CheckURL:          checkURL,
			BeforeStatusCode:  beforeStatusCode,
			CurrentStatusCode: currentStatusCode,
			HTTPError:         httpError,
			ResponseTime:      responseTime,
		}
		notifier.PostStatus(&param)
	}
}

func isNotify(beforeStatusCode int, currentStatusCode int, checkOnlyTopOfStatusCode bool) bool {
	if beforeStatusCode == NotFoundKey {
		return false
	}

	if checkOnlyTopOfStatusCode {
		if beforeStatusCode/100 == currentStatusCode/100 {
			return false
		}

	} else {
		if beforeStatusCode == currentStatusCode {
			return false
		}
	}

	return true
}
