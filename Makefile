.PHONY: build
VERSION=0.0.2

default: build publish

build: COMMIT=$(shell git rev-list -1 HEAD | grep -o "^.\{10\}")
build: DATE=$(shell date +'%Y-%m-%d %H:%M')
build: 
	env GOOS=darwin  GOARCH=amd64 go build -ldflags '-X "main.Version=$(VERSION) ($(COMMIT) - $(DATE))"' -o build/$(VERSION)/ssm-run-command-$(VERSION)-darwin
	env GOOS=linux   GOARCH=amd64 go build -ldflags '-X "main.Version=$(VERSION) ($(COMMIT) - $(DATE))"' -o build/$(VERSION)/ssm-run-command-$(VERSION)-linux
	env GOOS=windows GOARCH=amd64 go build -ldflags '-X "main.Version=$(VERSION) ($(COMMIT) - $(DATE))"' -o build/$(VERSION)/ssm-run-command-$(VERSION)-windows.exe

publish:
	rsync -a build/ /keybase/public/justmiles/artifacts/ssm-run-command/