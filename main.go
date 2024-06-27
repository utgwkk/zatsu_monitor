package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)
	slog.InfoContext(ctx, "started")

	config, err := LoadConfigFromFile("./config.yml")

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for name, values := range config {
		wg.Add(1)
		go func(name string, values map[string]string){
			defer func(){
				if err := recover(); err != nil {
					slog.Error("recovered from perform panic", slog.Any("error", err))
				}
			}()
			defer wg.Done()

			perform(ctx, name, values)
		}(name, values)
	}
	wg.Wait()
	slog.InfoContext(ctx, "done")
}

func printVersion() {
	fmt.Printf("zatsu_monitor %s, build %s\n", Version, Revision)
}

func perform(ctx context.Context, name string, values map[string]string) {
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
	currentStatusCode, httpError := GetStatusCode(ctx, checkURL)
	end := time.Now()
	responseTime := (end.Sub(start)).Seconds()

	slog.InfoContext(
		ctx,
		"request finished",
		slog.String("check_url", checkURL),
		slog.Int("status", currentStatusCode),
		slog.Float64("response_time", responseTime),
		slog.Any("error", httpError),
	)

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
		if err := notifier.PostStatus(&param); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to notify",
				slog.Any("error", err),
			)
		}
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
