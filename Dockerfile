FROM golang:1.21.6-alpine3.18 AS base

FROM base AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum main.go ./
RUN go mod download && go mod verify
RUN go build -ldflags "-s -w"

FROM base AS runner
RUN apk add ffmpeg yt-dlp
COPY --from=builder /usr/src/app/ccmusic /usr/bin/ccmusic

WORKDIR /tmp
EXPOSE 3000
CMD ["ccmusic"]