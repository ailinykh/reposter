# reposter

[![Build Status](https://github.com/ailinykh/reposter/workflows/build/badge.svg)](https://github.com/ailinykh/reposter/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/ailinykh/reposter)](https://goreportcard.com/report/github.com/ailinykh/reposter)
![GitHub](https://img.shields.io/github/license/ailinykh/reposter.svg)

This bot helps your telegram chat consume content in a more native way.

For example, if someone sends a link to a webm video:

![bot-webm-video](https://user-images.githubusercontent.com/939390/95298451-c7757100-0884-11eb-9140-4c6474959720.gif)


Or sends a video file as a document:

![bot-video-convert](https://user-images.githubusercontent.com/939390/95298623-07d4ef00-0885-11eb-92e4-b3c2015f7ecc.gif)


It even supports links to Twitter videos:

![bot-twitter-video](https://user-images.githubusercontent.com/939390/95298730-3783f700-0885-11eb-9650-b0c04e40aa2f.gif)

... and images as well!

![bot-twitter-images](https://user-images.githubusercontent.com/939390/95298790-4cf92100-0885-11eb-8bb2-8adbc91f5b23.gif)

## How to Run

Set up the environment:

```shell
brew install go ffmpeg yt-dlp
```

Clone the repository:

```shell
git clone https://github.com/ailinykh/reposter.git
cd reposter
```

Install Go dependencies:

```shell
go mod download
```

Obtain your bot token from [@BotFather](https://t.me/BotFather) and your Telegram ID using the `/info` command from [@reposter](https://t.me/reposter).

Copy the example environment file:

```shell
cp .env.example .env
```

Specify the required secrets in the `.env` file, then run:

```shell
make
```
