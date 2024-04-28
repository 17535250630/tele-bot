package main

import (
	controller "tele-bot"
	"tele-bot/pub"

	tele "gopkg.in/telebot.v3"
)

type Context struct {
	TokenMint string
}

func main() {
	c, err := controller.NewController[Context]("", 1, "")
	if err != nil {
		panic(err.Error())
	}
	c.AddHandler(1, "/start", func(ctx tele.Context) error {
		if !c.Filter(ctx) {
			return nil
		}
		c.Mailbox.Publish(pub.Msg[Context]{
			MsgID:      1,
			MsgContext: Context{},
		})
		return nil
	})
	c.Start()
}
