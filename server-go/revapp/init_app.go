package revapp

import (
	"github.com/strongo/bots-framework/core"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/prizarena/reversi/server-go/revdal/revgaedal"
	"github.com/prizarena/reversi/server-go/revbot"
)

func InitApp(botHost bots.BotHost) {
	revgaedal.RegisterDal()

	httpRouter := httprouter.New()
	http.Handle("/", httpRouter)

	revbot.InitBot(httpRouter, botHost, revAppContext{})
}
