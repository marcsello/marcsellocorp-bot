package main

import (
	"github.com/marcsello/marcsellocorp-bot/api"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/memdb"
	"github.com/marcsello/marcsellocorp-bot/telegram"
	"gitlab.com/MikeTTh/env"
	"log"
	"sync"
)

func main() {
	log.Println("Staring Marcsello Corp. Telegram Bot...")
	debug := env.Bool("DEBUG", false)

	log.Println("Connecting to DB...")
	err := db.Connect()
	if err != nil {
		panic(err)
	}

	log.Println("Connecting to Redis...")
	err = memdb.InitRedisConnection()
	if err != nil {
		panic(err)
	}

	log.Println("Init BOT...")
	botRun, err := telegram.InitTelegramBot(debug)
	if err != nil {
		panic(err)
	}

	log.Println("Init API...")
	apiRun, err := api.InitApi(debug)
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
		botRun()
		wg.Done()
	}()

	wg.Wait()

}
