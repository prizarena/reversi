package revbot

import (
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/telegram"
	"context"
	"github.com/prizarena/reversi/server-go/revbot/platforms/revtgbot"
	"github.com/prizarena/reversi/server-go/revbot/revrouting"
	"github.com/prizarena/reversi/server-go/revsecrets"
)

func InitBot(httpRouter *httprouter.Router, botHost bots.BotHost, appContext bots.BotAppContext) error {
	gaSettings := bots.AnalyticsSettings{
		GaTrackingID: revsecrets.GaTrackingID,
		// Enabled: func(r *http.Request) bool {
		// 	return false
		// },
	}

	driver := bots.NewBotDriver(gaSettings, appContext, botHost,
		"Please report any issues to @trakhimenok",
	)
	// routing.WebhooksRouter

	if driver == nil {
		panic("driver == nil")
	}

	newTranslator := func(c context.Context) strongo.Translator {
		return strongo.NewMapTranslator(c, nil)
	}

	driver.RegisterWebhookHandlers(httpRouter, "/bot",
		telegram.NewTelegramWebhookHandler(
			func(c context.Context) bots.SettingsBy {
				return revtgbot.Bots(c, strongo.EnvProduction, revrouting.WebhooksRouter) // gaestandard.GetEnvironment(c)
			},             // Maps of bots by code, language, token, etc...
			newTranslator, // Creates newTranslator that gets a context.Context (for logging purpose)
		),
	)

	return nil
}
