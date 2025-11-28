go.get:
	go get ./...

go.tidy:
	go mod tidy

build.cli:
	go build -o build/pisces -ldflags="-X main.version=$(shell git tag --sort=-v:refname)" cmd/cli/main.go

check:
	golangci-lint run

test: check
	go test ./...

watch:
	watchexec -r -e go -- "make test && make build.cli"

chromedp.pull:
	docker pull chromedp/headless-shell:latest

chromedp.run:
	docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
