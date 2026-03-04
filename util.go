package main

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
)

// GetStatusCode checks and returns status code for URL
func GetStatusCode(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	req.Header.Add("User-Agent", "Zatsu_Monitor/"+Version+"("+Revision+")")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// IsSuccessfulStatus returns whether status code is successful. (2xx or 3xx)
func IsSuccessfulStatus(statusCode int) bool {
	n := statusCode / 100

	// Successful: 2xx, 3xx
	return n == 2 || n == 3
}
