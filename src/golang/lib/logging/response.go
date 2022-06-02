package logging

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/context_parsing"
	log "github.com/sirupsen/logrus"
)

type Component string

const (
	ServerComponent Component = "Server"
)

func LogRoute(
	ctx context.Context,
	routeName string,
	r *http.Request,
	excludedHeaderFields map[string]bool,
	statusCode int,
	component Component,
	serviceName string,
	err error,
) {
	headers := make(map[string][]string, len(r.Header)-len(excludedHeaderFields))
	for k, v := range r.Header {
		if _, ok := excludedHeaderFields[k]; !ok {
			headers[k] = v
		}
	}

	status := "SUCCEEDED"
	var errMsg string
	if err != nil {
		status = "ERROR"
		errMsg = err.Error()
	}

	log.WithFields(log.Fields{
		"ServiceName":   serviceName,
		"URL":           r.URL,
		"Headers":       headers,
		"Status":        status,
		"Code":          statusCode,
		"Component":     component,
		"Route":         routeName,
		"UserId":        ctx.Value(context_parsing.UserIdKey),
		"UserRequestId": ctx.Value(context_parsing.UserRequestIdKey),
		"Error":         errMsg,
	}).Info()
}

func LogAsyncEvent(
	ctx context.Context,
	component Component,
	serviceName string,
	err error,
) {
	status := "SUCCEEDED"
	var errMsg string
	if err != nil {
		status = "ERROR"
		errMsg = err.Error()
	}

	log.WithFields(log.Fields{
		"ServiceName":   serviceName,
		"Status":        status,
		"Component":     component,
		"UserId":        ctx.Value(context_parsing.UserIdKey),
		"UserRequestId": ctx.Value(context_parsing.UserRequestIdKey),
		"Error":         errMsg,
	}).Info()
}
