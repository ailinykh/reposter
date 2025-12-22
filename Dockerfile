FROM --platform=linux/amd64 golang:1.22 as builder
WORKDIR /go/src/github.com/ailinykh/reposter
ADD . .
RUN go mod download
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -v -a -installsuffix cgo -ldflags '-extldflags "-static"' ./cmd/bot

FROM alpine
LABEL maintainer="Anton Ilinykh <anthonyilinykh@gmail.com>"
RUN apk update && apk add --no-cache ffmpeg tzdata python3  && \
    wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp
COPY --from=builder /go/src/github.com/ailinykh/reposter/bot /usr/local/bin/bot
WORKDIR /usr/local/share
VOLUME [ "/usr/local/share/reposter-data" ]
ENTRYPOINT bot