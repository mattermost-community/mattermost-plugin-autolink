.PHONY: build test run clean stop check-style gofmt dist

MKFILE_PATH=$(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_DIR_NAME=$(notdir $(patsubst %/,%,$(dir $(MKFILE_PATH))))
PLUGIN_NAME=$(CURRENT_DIR_NAME)
GOOS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH=amd64

all: dist

check-style: gofmt
	@echo Checking for style guide compliance

gofmt:
	@echo Running GOFMT

	@for package in $$(go list ./server/...); do \
		echo "Checking "$$package; \
		files=$$(go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' $$package); \
		if [ "$$files" ]; then \
			gofmt_output=$$(gofmt -d -s $$files 2>&1); \
			if [ "$$gofmt_output" ]; then \
				echo "$$gofmt_output"; \
				echo "gofmt failure"; \
				exit 1; \
			fi; \
		fi; \
	done
	@echo "gofmt success"; \

test:
	go test -v -coverprofile=coverage.txt ./...

vendor: server/Gopkg.lock
	@echo Run this to updated the go lang dependencies after a major release
	cd server && dep ensure -update

dist: check-style
	@echo Building plugin

	# Clean old dist
	rm -rf dist
	rm -rf webapp/dist
	rm -f server/plugin.exe

	# Build files from server
	cd server && go get github.com/mitchellh/gox
	$(shell go env GOPATH)/bin/gox -osarch='darwin/amd64 linux/amd64 windows/amd64' -output 'dist/intermediate/plugin_{{.OS}}_{{.Arch}}' ./server

	# Copy plugin files
	mkdir -p dist/$(PLUGIN_NAME)/
	cp plugin.json dist/$(PLUGIN_NAME)/

	# Copy server executables & compress plugin
	mkdir -p dist/$(PLUGIN_NAME)/server
	mv dist/intermediate/plugin_darwin_amd64 dist/$(PLUGIN_NAME)/server/plugin.exe
	cd dist && tar -zcvf $(PLUGIN_NAME)-darwin-amd64.tar.gz $(PLUGIN_NAME)/*
	mv dist/intermediate/plugin_linux_amd64 dist/$(PLUGIN_NAME)/server/plugin.exe
	cd dist && tar -zcvf $(PLUGIN_NAME)-linux-amd64.tar.gz $(PLUGIN_NAME)/*
	mv dist/intermediate/plugin_windows_amd64.exe dist/$(PLUGIN_NAME)/server/plugin.exe
	cd dist && tar -zcvf $(PLUGIN_NAME)-windows-amd64.tar.gz $(PLUGIN_NAME)/*

	# Clean up temp files
	rm -rf dist/$(PLUGIN_NAME)
	rm -rf dist/intermediate

	@echo MacOS X plugin built at: dist/$(PLUGIN_NAME)-darwin-amd64.tar.gz
	@echo Linux plugin built at: dist/$(PLUGIN_NAME)-linux-amd64.tar.gz
	@echo Windows plugin built at: dist/$(PLUGIN_NAME)-windows-amd64.tar.gz

deploy: dist
	cp dist/$(PLUGIN_NAME)-$(GOOS)-$(GOARCH).tar.gz ../mattermost-server/plugins/
	rm -rf ../mattermost-server/plugins/$(PLUGIN_NAME)
	tar -C ../mattermost-server/plugins/ -zxvf ../mattermost-server/plugins/$(PLUGIN_NAME)-$(GOOS)-$(GOARCH).tar.gz

run: .npminstall
	@echo Not yet implemented

stop:
	@echo Not yet implemented

clean:
	@echo Cleaning plugin

	rm -rf dist
	rm -rf webapp/dist
	rm -rf webapp/node_modules
	rm -rf webapp/.npminstall
	rm -f server/plugin.exe