package main

import (
	"net/http"
	"time"

	"shareme/share"

	sentry "github.com/getsentry/sentry-go"
)

func init() {
	sentry.Init(
		sentry.ClientOptions{
			Dsn: share.Config.SentryDsn,
			HTTPTransport: &http.Transport{
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 30 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				IdleConnTimeout:       30 * time.Second,
				MaxIdleConnsPerHost:   64,
			},
		},
	)
}
