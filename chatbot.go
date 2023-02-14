package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	openai.SetAuth(os.Getenv("OPENAI_API_KEY"))

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! How can I help you today?")
				bot.Send(msg)
			case "image":
				// extract the number of images to generate from the command argument
				var number int
				if len(update.Message.CommandArguments()) > 0 {
					var err error
					number, err = strconv.Atoi(update.Message.CommandArguments())
					if err != nil {
						number = 1
					}
				} else {
					number = 1
				}

				go generateImage(bot, update.Message.Chat.ID, update.Message.Text, number)
			}
		} else {
			go generateResponse(bot, update.Message.Chat.ID, update.Message.Text)
		}
	}
}

func generateImage(bot *tgbotapi.BotAPI, chatID int64, commandText string, number int) {
	prompt := strings.TrimSpace(strings.TrimPrefix(commandText, "/image"))
	if prompt == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "Please provide a prompt for the image generation command."))
		return
	}

	var urls []string
	for i := 0; i < number; i++ {
		res, err := openai.CreateImage(prompt, nil)
		if err != nil {
			log.Printf("Error generating image: %s", err)
			bot.Send(tgbotapi.NewMessage(chatID, "Could not generate image, try again later."))
			return
		}
		urls = append(urls, res.URL)
	}

	for _, url := range urls {
		msg := tgbotapi.NewPhotoShare(chatID, url)
		bot.Send(msg)
	}
}

func generateResponse(bot *tgbotapi.BotAPI, chatID int64, messageText string) {
	response, err := openai.Complete(messageText, &openai.CompletionRequest{
		Model:            "text-davinci-003",
		Prompt:           "You are an AI named Sonic and you are in a conversation with a human. You can answer questions, provide information, and help with a wide variety of tasks.\n\n" + messageText,
		Temperature:      0.7,
		MaxTokens:        256,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	})
	if err != nil {
		log.Printf("Error generating response: %s", err)
		bot.Send(tgbot
