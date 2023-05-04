package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	BotToken   = "6254482252:AAGBiwZdfPtOa_3gIilTmC-25VwLY4HQxyI"
	WebhookURL = "https://app-golang-telebot.herokuapp.com"
)

var rss = map[string]string{
	"Habr": "https://habr.com/rss/best",
}

type RSS struct {
	Items []Item `xml:"channel>item"`
}

type Item struct {
	URL   string `xml:"guid"`
	Title string `xml:"title"`
}

func getNews(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	rss := new(RSS)
	err = xml.Unmarshal(body, rss)
	if err != nil {
		return nil, err
	}
	return rss, nil
}

func main() {
	port := os.Getenv("PORT")

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	// bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":"+port, nil)
	fmt.Println("start listen :" + port)

	//receive all updates from the channel `updates`
	for update := range updates {
		if url, ok := rss[update.Message.Text]; ok {
			rss, err := getNews(url)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"sorry, error happend",
				))
			}
			for _, item := range rss.Items {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					item.URL+"\n"+item.Title,
				))
			}
		} else {
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`there is only Habr feed available`,
			))
		}
	}
}
