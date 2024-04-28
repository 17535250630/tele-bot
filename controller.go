package controller

import (
	"tele-bot/pub"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Controller[T any] struct {
	Mailbox  *pub.ControllerPublisher[T]
	bot      *tele.Bot
	banUsers mapset.Set[int64]
}

func NewController[T any](token string, admin int64, helloMsg string) (*Controller[T], error) {
	mailBox := pub.NewPublisher[T](12, 1024)
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 20 * time.Second},
	}
	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}
	adminOnly := b.Group()
	adminOnly.Use(middleware.Whitelist(admin))
	controller := &Controller[T]{
		Mailbox:  mailBox,
		bot:      b,
		banUsers: mapset.NewSet[int64](),
	}
	ok, err := controller.SentMessage(helloMsg, admin)
	if err != nil {
		return nil, err
	}
	_ = ok
	return controller, nil
}

func (self *Controller[T]) AddHandler(msgID int, endpoint string, handlerFunc func(ctx tele.Context) error) int {
	self.bot.Handle(endpoint, handlerFunc)
	return msgID
}

func (self *Controller[T]) Filter(c tele.Context) bool {
	if self.banUsers.Contains(c.Chat().ID) {
		return false
	}
	return true
}

func (self *Controller[T]) SentMessage(msg string, chatID int64) (bool, error) {
	send, err := self.bot.Send(tele.ChatID(chatID), msg, &tele.SendOptions{
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeMarkdown,
	})
	if err != nil {
		return false, err
	}
	_ = send
	return true, nil
}

func (self *Controller[T]) Start() {
	self.bot.Start()
}
