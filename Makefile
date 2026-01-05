go.get:
	go get ./...

go.tidy:
	go mod tidy

build.cli:
	go build -o build/pisces -ldflags="-X main.version=$(shell git tag --sort=-v:refname | head -1)" cmd/cli/main.go

check:
	golangci-lint run

test: check
	go test ./...

watch:
	watchexec -r -e go,js -- "make test && make build.cli"

chromedp.pull:
	docker pull chromedp/headless-shell:latest

chromedp.run:
	docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell

.PHONY: release

test.docker:
	docker build -f Dockerfile.dev -t pisces-dev . && \
	docker run --rm -it -v "$$PWD":/app -w /app pisces-dev go test -v ./...

run.docker:
	docker build -t pisces . && \
	docker run --rm -it \
		--network host \
		-v "$$PWD":/app \
		-w /app \
		pisces $(ARGS)

release:
	@echo "Creating new release..."
	@# Get the latest tag, default to v0.0.0 if no tags exist
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "Latest tag: $$LATEST_TAG"; \
	MAJOR=$$(echo $$LATEST_TAG | sed 's/v\([0-9]*\)\.[0-9]*\.[0-9]*/\1/'); \
	MINOR=$$(echo $$LATEST_TAG | sed 's/v[0-9]*\.\([0-9]*\)\.[0-9]*/\1/'); \
	PATCH=$$(echo $$LATEST_TAG | sed 's/v[0-9]*\.[0-9]*\.\([0-9]*\)/\1/'); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	echo "Creating new tag: $$NEW_TAG"; \
	git tag -s -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	echo "Pushing tag to origin..."; \
	git push origin $$NEW_TAG; \
	echo "Release $$NEW_TAG created and pushed successfully!"
