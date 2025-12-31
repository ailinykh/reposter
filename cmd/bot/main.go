package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/ailinykh/reposter/v2/internal/api/http_logger"
	"github.com/ailinykh/reposter/v2/internal/api/image_uploader"
	"github.com/ailinykh/reposter/v2/internal/api/logger"
	"github.com/ailinykh/reposter/v2/internal/api/xui"
	"github.com/ailinykh/reposter/v2/internal/legacy/api"
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/infrastructure"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	workingDir := "."
	if wd, ok := os.LookupEnv("WORKING_DIR"); ok {
		workingDir = wd
	}
	logger := logger.NewGoogleLogger(ctx, workingDir)
	logger.Info("config loaded", "working_dir", workingDir)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Error("signal received", "signal", sig)
		cancel()
	}()

	dbFile := path.Join(workingDir, "reposter.db")

	settingsProvider := infrastructure.CreateSettingsStorage(dbFile)
	boolSettingProvider := helpers.CreateBoolSettingProvider(settingsProvider)
	databaseChatStorage := infrastructure.CreateChatStorage(dbFile, logger)
	inMemoryChatStorage := infrastructure.CreateInMemoryChatStorage()
	chatStorageDecorator := usecases.CreateChatStorageDecorator(inMemoryChatStorage, databaseChatStorage)

	chatID, err := strconv.ParseInt(os.Getenv("REPORT_CHAT_ID"), 10, 64)
	if err != nil {
		logger.Info("report to chat disabled", "error", err)
	} else {
		logger.Info("using report", "chat_id", chatID)
	}

	transport := http.DefaultTransport
	if _, ok := os.LookupEnv("EXTENDED_LOGGING_ENABLED"); ok {
		logger.Info("extended logging enabled")
		transport = http_logger.NewLoggingRoundTripper(
			http.DefaultTransport,
			infrastructure.NewLogStorage(path.Join(workingDir, "reposter-logs.db"), logger),
			logger,
		)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Minute,
	}

	telebot := api.CreateTelebot(
		logger,
		api.WithBotToken(os.Getenv("TELEGRAM_BOT_TOKEN")),
		api.WithReportChatId(chatID),
		api.WithClient(client),
	)
	telebot.SetupInfo()

	databaseUserStorage := infrastructure.CreateUserStorage(dbFile, logger)
	inMemoryUserStorage := infrastructure.CreateInMemoryUserStorage()
	userStorageDecorator := usecases.CreateUserStorageDecorator(inMemoryUserStorage, databaseUserStorage)
	bootstrapFlow := usecases.CreateBootstrapFlow(logger, chatStorageDecorator, userStorageDecorator)
	telebot.AddHandler(bootstrapFlow)

	localizer := infrastructure.CreateGameLocalizer()
	gameStorage := infrastructure.CreateGameStorage(dbFile)
	rand := infrastructure.CreateMathRand()
	commandService := usecases.CreateCommandService()
	gameFlow := usecases.CreateGameFlow(logger, localizer, gameStorage, rand, settingsProvider, commandService)
	telebot.AddHandler(gameFlow)

	converter := infrastructure.CreateFfmpegConverter(logger)
	videoFlow := usecases.CreateVideoFlow(logger, converter, converter)
	telebot.AddHandler(videoFlow)

	fileDownloader := infrastructure.CreateFileDownloader(logger)
	remoteMediaSender := helpers.CreateSendMediaStrategy()
	sendVideoStrategy := helpers.CreateSendVideoStrategy()
	sendVideoStrategySplitDecorator := helpers.CreateSendVideoStrategySplitDecorator(logger, sendVideoStrategy, converter)
	localMediaSender := helpers.CreateUploadMediaDecorator(logger, remoteMediaSender, fileDownloader, converter, sendVideoStrategySplitDecorator)

	var task core.ITask = core.TaskMock{}
	if amqpUrl, ok := os.LookupEnv("AMQP_URL"); ok {
		rabbit, close := infrastructure.CreateRabbitFactory(logger, amqpUrl)
		defer close()
		task = rabbit.NewTask("twitter_queue")
	}

	twitterMediaFactory := api.CreateTwitterMediaFactory(logger, task)
	twitterFlow := usecases.CreateTwitterFlow(logger, twitterMediaFactory, localMediaSender)
	twitterTimeout := usecases.CreateTwitterTimeout(logger, twitterFlow)
	twitterParser := usecases.CreateTwitterParser(logger, twitterTimeout)
	twitterRemoveSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, twitterParser, core.STwitterFlowRemoveSource, boolSettingProvider)
	telebot.AddHandler(twitterRemoveSourceDecorator)

	httpClient := api.CreateHttpClient()
	convertMediaSender := helpers.CreateConvertMediaStrategy(logger, localMediaSender, fileDownloader, converter, converter)
	linkFlow := usecases.CreateLinkFlow(logger, httpClient, converter, convertMediaSender)
	removeLinkSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, linkFlow, core.SLinkFlowRemoveSource, boolSettingProvider)
	telebot.AddHandler(removeLinkSourceDecorator)

	tiktokHttpClient := api.CreateHttpClient() // domain specific headers and cookies
	ytdlpApi := api.CreateYtDlpApi([]string{}, logger)
	tiktokMediaFactory := api.CreateTikTokMediaFactory(ytdlpApi)
	tiktokFlow := usecases.CreateTikTokFlow(tiktokHttpClient, tiktokMediaFactory, localMediaSender)
	telebot.AddHandler(tiktokFlow)

	imageUploader := image_uploader.NewPostimages(logger)
	//TODO: image_downloader := api.CreateTelebotImageDownloader()
	imageFlow := usecases.CreateImageFlow(logger, imageUploader, telebot)
	telebot.AddHandler(imageFlow)

	{
		chatId := os.Getenv("PUBLISHER_CHAT_ID")
		username := os.Getenv("PUBLISHER_USERNAME")
		if len(chatId) > 0 && len(username) > 0 {
			logger.Info("publisher logic enabled", "chat_id", chatId, "username", username)
			chatID, err := strconv.ParseInt(chatId, 10, 64)
			if err != nil {
				logger.Error("failed to parse publisher chat id", "error", err)
			} else {
				publisherFlow := usecases.CreatePublisherFlow(chatID, username, logger)
				telebot.AddHandler(publisherFlow)
				telebot.AddHandler("/loh666", publisherFlow.HandleRequest)
			}
		}
	}

	youtubeMediaFactory := api.CreateYoutubeMediaFactory(logger, ytdlpApi, fileDownloader)
	youtubeFlow := usecases.CreateYoutubeFlow(logger, youtubeMediaFactory, youtubeMediaFactory, sendVideoStrategySplitDecorator)
	removeYoutubeSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, youtubeFlow, core.SYoutubeFlowRemoveSource, boolSettingProvider)
	telebot.AddHandler(removeYoutubeSourceDecorator)

	telebot.AddHandler("/proxy", func(m *core.Message, bot core.IBot) error {
		_ = commandService.EnableCommands(m.Chat.ID, []core.Command{{Text: "proxy", Description: "proxy server for telegram"}}, bot)
		_, err := bot.SendText("tg://proxy?server=proxy.ailinykh.com&port=443&secret=dd71ce3b5bf1b7015dc62a76dc244c5aec")
		return err
	})

	iDoNotCare := usecases.CreateIDoNotCare()
	telebot.AddHandler(iDoNotCare)

	if cookiesFilePath := os.Getenv("INSTAGRAM_COOKIES_FILE_PATH"); len(cookiesFilePath) > 0 {
		logger.Info("instagram logic enabled. Cookies file: %s", cookiesFilePath)
		cookies := path.Join(workingDir, cookiesFilePath)
		instaAPI := api.CreateYtDlpApi([]string{"--cookies", cookies}, logger)
		instaFlow := usecases.CreateInstagramFlow(logger, instaAPI, localMediaSender)
		removeInstaSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, instaFlow, core.SInstagramFlowRemoveSource, boolSettingProvider)
		telebot.AddHandler(removeInstaSourceDecorator)
	} else {
		logger.Info("instagram logic disabled")
	}

	{
		baseUrl := os.Getenv("XUI_BASE_URL")
		login := os.Getenv("XUI_LOGIN")
		password := os.Getenv("XUI_PASSWORD")
		if len(baseUrl) > 0 && len(login) > 0 && len(password) > 0 {
			logger.Info("xui vpn logic enabled", "username", login)
			api := xui.NewClient(logger, baseUrl, login, password)
			localizer := infrastructure.CreateVpnLocalizer()
			flow := usecases.CreateVpnFlow(logger, localizer, api, settingsProvider)
			telebot.AddHandler(flow)
		} else {
			logger.Info("xui vpn logic disabled")
		}
	}

	commonLocalizer := infrastructure.CreateCommonLocalizer()
	startFlow := usecases.CreateStartFlow(logger, commonLocalizer, settingsProvider, commandService)
	telebot.AddHandler("/start", startFlow.Start)
	telebot.AddHandler("/help", startFlow.Help)
	// Start endless loop
	go telebot.Run()

	logger.Info("waiting for context...")
	<-ctx.Done()
	telebot.Stop()
	logger.Info("attempt to shutdown gracefully...")
}
