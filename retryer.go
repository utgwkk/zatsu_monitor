package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
)

// ref: https://future-architect.github.io/articles/20211026a/

type Retryer struct {
	client.DefaultRetryer
}

func (r *Retryer) ShouldRetry(req *request.Request) bool {
	if isErrReadConnectionReset(req.Error) {
		return true
	}
	return r.DefaultRetryer.ShouldRetry(req)
}

func isErrReadConnectionReset(err error) bool {
	switch e := err.(type) {
	case awserr.Error:
		origErr := e.OrigErr()
		if origErr != nil {
			return isErrReadConnectionReset(origErr)
		}
	case interface{ Temporary() bool }:
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			return true
		}
	}
	return false
}
