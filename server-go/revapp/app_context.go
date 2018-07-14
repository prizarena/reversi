package revapp

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"reflect"
	"time"
	"github.com/prizarena/reversi/server-go/revmodels"
	"github.com/prizarena/reversi/server-go/revtrans"
)

type revAppContext struct {
}

func (appCtx revAppContext) AppUserEntityKind() string {
	return revmodels.UserKind
}

func (appCtx revAppContext) AppUserEntityType() reflect.Type {
	return reflect.TypeOf(&revmodels.UserEntity{})
}

func (appCtx revAppContext) NewBotAppUserEntity() bots.BotAppUser {
	return &revmodels.UserEntity{
		DtCreated: time.Now(),
	}
}

func (appCtx revAppContext) NewAppUserEntity() strongo.AppUser {
	return appCtx.NewBotAppUserEntity()
}

func (appCtx revAppContext) GetTranslator(c context.Context) strongo.Translator {
	return strongo.NewMapTranslator(c, revtrans.TRANS)
}

type LocalesProvider struct {
}

func (LocalesProvider) GetLocaleByCode5(code5 string) (strongo.Locale, error) {
	return strongo.LocaleEnUS, nil
}

func (appCtx revAppContext) SupportedLocales() strongo.LocalesProvider {
	return RevLocalesProvider{}
}

type RevLocalesProvider struct {
}

func (RevLocalesProvider) GetLocaleByCode5(code5 string) (locale strongo.Locale, err error) {
	switch code5 {
	case strongo.LocaleCodeEnUS:
		return strongo.LocaleEnUS, nil
	case strongo.LocalCodeRuRu:
		return strongo.LocaleRuRu, nil
	case strongo.LocaleCodeEsES:
		return strongo.LocaleEsEs, nil
	case strongo.LocaleCodeFrFR:
		return strongo.LocaleFrFr, nil
	case strongo.LocaleCodeEnUK:
		return strongo.LocaleEsEs, nil
	case strongo.LocaleCodeFaIR:
		return strongo.LocaleFaIr, nil
	case strongo.LocaleCodeDeDE:
		return strongo.LocaleDeDe, nil
	case strongo.LocaleCodeItIT:
		return strongo.LocaleItIt, nil
	case strongo.LocaleCodeUzUZ:
		return strongo.LocaleUzUz, nil
	default:
		return locale, errors.New("Unsupported locale: " + code5)
	}
}

var _ strongo.LocalesProvider = (*RevLocalesProvider)(nil)

func (appCtx revAppContext) GetBotChatEntityFactory(platform string) func() bots.BotChat {
	switch platform {
	case telegram.PlatformID:
		return func() bots.BotChat {
			return &telegram.ChatEntity{
				TgChatEntityBase: *telegram.NewTelegramChatEntity(),
			}
		}
	default:
		panic("Unknown platform: " + platform)
	}
}

var _ bots.BotAppContext = (*revAppContext)(nil)

