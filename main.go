package main

import (
	"github.com/marcsello/marcsellocorp-bot/api"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/telegram"
	"log"
	"sync"
)

func main() {
	log.Println("Staring Marcsello Corp. Telegram Bot...")

	log.Println("Connecting to DB...")
	db.Connect()

	log.Println("Init BOT...")
	bot, err := telegram.InitTelegramBot()
	if err != nil {
		panic(err)
	}

	log.Println("Init API...")
	apiRun, err := api.InitApi(bot)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		log.Println("Staring API...")
		apiRun()
		wg.Done()
	}()

	go func() {
		log.Println("Staring BOT...")
		bot.Start()
		wg.Done()
	}()

	wg.Wait()

}
