package telegram

import "gopkg.in/telebot.v3"

func Send(toID int64, what interface{}) (int, error) {
	msg, err := telegramBot.Send(&telebot.User{ID: toID}, what, telebot.ModeDefault)
	if err != nil {
		return 0, err
	}
	return msg.ID, nil
}
