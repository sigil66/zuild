all: zuild

clean:
	rm -rf dist

zuild: clean
	go build -o dist/usr/bin/zuild github.com/sigil66/zuild/cli/zuild

fmt:
	goimports -w $$(go list -f {{.Dir}} ./... | grep -v /vendor/)

.PHONY: deps fmt