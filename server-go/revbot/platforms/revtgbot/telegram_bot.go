package revtgbot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"context"
	"github.com/strongo/log"
	"github.com/prizarena/reversi/server-go/revsecrets"
)

var botsBy bots.SettingsBy

func Bots(c context.Context, env strongo.Environment, router bots.WebhooksRouter) bots.SettingsBy {
	if len(botsBy.ByCode) == 0 {
		routerByProfile := func(profile string) bots.WebhooksRouter {
			return router // We have single profile for now
		}

		switch env {
		case strongo.EnvProduction:
			botsBy = bots.NewBotSettingsBy(routerByProfile,
				telegram.NewTelegramBot(strongo.EnvProduction, "Reversi",
					revsecrets.TelegramProdBot, revsecrets.TelegramProdToken,
					"", "", revsecrets.GaTrackingID, strongo.LocaleEnUS),
			)
		default:
			log.Errorf(c, "Unknown environment: %v=%v", env, strongo.EnvironmentNames[env])
		}
	}
	return botsBy
}
