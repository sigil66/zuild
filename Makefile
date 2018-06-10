all: zuild

clean:
	rm -rf dist

deps:
	dep ensure

zuild: clean deps
	go build -o dist/usr/bin/zuild github.com/solvent-io/zuild/cli/zuild

fmt:
	goimports -w $$(go list -f {{.Dir}} ./... | grep -v /vendor/)

.PHONY: deps fmt