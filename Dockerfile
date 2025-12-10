FROM golang:1.24-bullseye AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG PISCES_VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w -X main.version=${PISCES_VERSION}" \
      -o /app/pisces \
      ./cmd/cli

FROM chromedp/headless-shell:latest

ENV PATH="/headless-shell:${PATH}"

COPY --from=builder /app/pisces /usr/local/bin/pisces

COPY . /app

WORKDIR /app

ENTRYPOINT ["pisces"]
# CMD ["--help"]
