.PHONY: build

AMPLITUDE_API_KEY ?=

LDFLAGS := -X github.com/toss/apps-in-toss-ax/pkg/instrumentation.AMPLITUDE_API_KEY=$(AMPLITUDE_API_KEY)

build:
	go build -a -ldflags "$(LDFLAGS)" -o ax .
