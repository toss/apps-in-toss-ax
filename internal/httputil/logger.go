package httputil

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "httputil")

// RetryLogger는 go-retryablehttp의 LeveledLogger, RequestLogHook,
// ResponseLogHook을 하나의 구조체로 통합합니다.
type RetryLogger struct {
	entry *logrus.Entry
}

// NewRetryLogger는 RetryLogger를 생성합니다.
func NewRetryLogger() *RetryLogger {
	return &RetryLogger{entry: log}
}

// Apply는 retryablehttp.Client에 로거와 훅을 일괄 설정합니다.
func (l *RetryLogger) Apply(rc *retryablehttp.Client) {
	rc.Logger = l
	rc.RequestLogHook = l.OnRequest
	rc.ResponseLogHook = l.OnResponse
}

// --- LeveledLogger 구현 ---

func (l *RetryLogger) fields(keysAndValues ...any) *logrus.Entry {
	e := l.entry
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keysAndValues[i])
		}
		e = e.WithField(key, keysAndValues[i+1])
	}
	return e
}

func (l *RetryLogger) Error(msg string, keysAndValues ...any) {
	l.fields(keysAndValues...).Error(msg)
}

func (l *RetryLogger) Info(msg string, keysAndValues ...any) {
	l.fields(keysAndValues...).Info(msg)
}

func (l *RetryLogger) Debug(msg string, keysAndValues ...any) {
	l.fields(keysAndValues...).Debug(msg)
}

func (l *RetryLogger) Warn(msg string, keysAndValues ...any) {
	l.fields(keysAndValues...).Warn(msg)
}

// --- Hook 구현 ---

// OnRequest는 retryablehttp.RequestLogHook 시그니처에 맞는 훅입니다.
func (l *RetryLogger) OnRequest(_ retryablehttp.Logger, req *http.Request, attempt int) {
	l.entry.WithFields(logrus.Fields{
		"method":  req.Method,
		"url":     req.URL.String(),
		"attempt": attempt,
	}).Debug("request attempt")
}

// OnResponse는 retryablehttp.ResponseLogHook 시그니처에 맞는 훅입니다.
func (l *RetryLogger) OnResponse(_ retryablehttp.Logger, resp *http.Response) {
	l.entry.WithFields(logrus.Fields{
		"method": resp.Request.Method,
		"url":    resp.Request.URL.String(),
		"status": resp.StatusCode,
	}).Debug("response received")
}
