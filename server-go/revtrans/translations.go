package revtrans

import (
	"github.com/strongo/bots-framework/core"
	"github.com/prizarena/prizarena-public/patrans"
	"github.com/strongo/emoji/go/emoji"
)

func init() {
	patrans.RegisterTranslations(TRANS)
}

var TRANS = map[string]map[string]string{
	bots.MessageTextOopsSomethingWentWrong: {
		"ru-RU": "Упс, что-то пошло не так... \xF0\x9F\x98\xB3",
		"en-US": "Oops, something went wrong... \xF0\x9F\x98\xB3",
		"fa-IR": "اوه، یک جای کار مشکل دارد ...  \xF0\x9F\x98\xB3",
		"it-IT": "Ops, qualcosa e' andato storto... \xF0\x9F\x98\xB3",
	},
	MT_START_SELECT_LANG: {
		"en-US": "<b>Please select your language</b>\nПожалуйста выберите язык общения",
		"ru-RU": "<b>Пожалуйста выберите язык общения</b>\nPlease select your language",
	},
	FlagOfTheDay: {
		"en-US": `<i>To learn more about flag subscribe to</i> <a href="https://t.me/FlagOfTheDay">@FlagOfTheDay</a> <i>channel</i>.`,
		"ru-RU": `<i>Чтобы узнать больше о флагах подпишитесь на канал</i> <a href="https://t.me/FlagOfTheDay">@FlagOfTheDay</a>.`,
	},
	Flips: {
		"en-US": "<b>Turns</b>: %v",
		"ru-RU": "<b>Ходов</b>: %v",
	},
	NewGameInlineTitle: {
		"en-US": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - new game",
		"ru-RU": emoji.BlackCircle + emoji.WhiteCircle + " Реверси - новая игра",
	},
	NewGameInlineDescription: {
		"en-US": "Starts a new Reversi game",
		"ru-RU": "Создать новую игру Реверси",
	},
	GameCardTitle: {
		"en-US": "Reversi game",
		"ru-RU": "Игра: Реверси",
	},
	OnStartWelcome: {
		"en-US": `🀄 <b>Reversi</b> game

It has very simple <a href="https://en.wikipedia.org/wiki/Reversi#Rules">rules</a>.  

🤺 You can practice alone or play against friends. 

🏆 Create tournaments for your friends or subscribers of your channel.

💵 From time to time there are <b>free to play</b> sponsored tournaments with cash prizes! You can get notified about such tournaments by subscribing to @prizarena channel.`,
		"ru-RU": `🀄 <b>Игра</b>: Реверси

<a href="https://ru.wikipedia.org/wiki/%D0%A0%D0%B5%D0%B2%D0%B5%D1%80%D1%81%D0%B8#%D0%9F%D1%80%D0%B0%D0%B2%D0%B8%D0%BB%D0%B0">Правила</a> очень просты. 

🤺 Играть можно одному или с друзьями.

🏆 Проводите турниры среди друзей или подписчиков своего канала. 

💵 Иногда проводятся спонсорские турниры с <b>бесплатным участием</b> и денежными призами! Узнать о таких турнирах можно подписавшись на канал @prizarena.`,
		"fr-FR": `🀄 <b> Reversi </b> jeu

Il a des <a href="https://en.wikipedia.org/wiki/Reversi#Rules">règles très simples</a>.

🤺 Vous pouvez pratiquer seul ou jouer contre des amis.

🏆 Créez des tournois pour vos amis ou abonnés de votre chaîne.

💵 De temps en temps, il y a des <b>tournois gratuits </b> sponsorisés avec des prix en argent! Vous pouvez être averti de ces tournois en vous abonnant à la chaîne @prizarena.`,
		"es-ES": `🀄 <b>Reversi</b> juego

Tiene <a href="https://en.wikipedia.org/wiki/Reversi#Rules">reglas</a> muy simples.

🤺 Puedes practicar solo o jugar contra amigos.

🏆 Crea torneos para tus amigos o suscriptores de tu canal.

💵 De vez en cuando hay torneos patrocinados <b>gratis</b> con premios en efectivo. Puede recibir notificaciones sobre dichos torneos suscribiéndose al canal @prizarena.`,
		"de-DE": `🀄 <b>Reversi </b> Spiel

Es hat sehr einfache <a href="https://de.wikipedia.org/wiki/Reversi#Rules">Regeln</a>.

🤺 Du kannst alleine trainieren oder gegen Freunde spielen.

🏆 Erstelle Turniere für deine Freunde oder Abonnenten deines Kanals.

💵 Von Zeit zu Zeit gibt es <b>kostenlose</b> gesponserte Turniere mit Geldpreisen! Sie können über solche Turniere benachrichtigt werden, indem Sie @prizarena Kanal abonnieren.`,
		"fa-IR": `🀄 <b> بازی Reversi </b>

این <a href="https://en.wikipedia.org/wiki/Reversi#Rules"> قوانین </a> بسیار ساده است.

🤺 شما می توانید به تنهایی تمرین کنید و یا علیه دوستان بازی کنید.

🏆 ایجاد مسابقات برای دوستان یا مشترکین کانال شما.

💵 از زمان به زمان <b>رایگان برای بازی</b> مسابقات با حمایت مالی با جوایز نقدی وجود دارد! با عضویت در کانالprizarena می توانید در مورد این تورنماها اطلاعات دریافت کنید.`,
	},
	Board: {
		"en-US": "Board",
		"ru-RU": "Доска",
	},
	Tournaments: {
		"en-US": "🏆 Tournaments",
		"ru-RU": "🏆 Турниры",
	},
	FirstMoveDoneAwaitingSecond: {
		"en-US": "Player <b>%v</b> made choice, awaiting another player...",
		"ru-RU": "Игрок <b>%v</b> сделал свой ход, ожидается ход второго игрока...",
	},
	FindFast: {
		"en-US": "Find matching revs as fast as you can.",
		"ru-RU": "Найдите совпадающие пары настолько быстро как можете.",
	},
	RulesShort: {
		"en-US": `<pre></pre>`,
	},
	NewGameText: {
		"en-US": `<b>Reversi game</b>

Blacks make 1st move.`,
		"ru-RU": `Игра: <b>Реверси</b>

Чёрные ходят первыми.`,
	},
	MT_HOW_TO_START_NEW_GAME: {
		"en-US": `<b>To begin new game:</b>
  1. Open Telegram chat with your friend
  2. Type into the text input @BiddingTicTacToeBot and a space
  3. Wait for a popup menu and choose "New game"

<i>First 2 steps can be replaced by clicking the button below!</i>`,
		//
		"ru-RU": `<b>Чтобы начать игру:</b>
  1. Откройте чат с вашим другом
  2. Наберите в поле ввода @BiddingTicTacToeBot и пробел
  3. Дождитесь появлению меню и выберите "Новая игра"

<i>Два первых шага могут быть заменены одним кликом на кнопку ниже!</i>`,
	},
	MT_NEW_GAME_WELCOME: {
		"en-US": `Black makes first move.`,
		"ru-RU": `Чёрные ходят первыми.`,
	},
	MT_HOW_TO_INLINE: {
		"en-US": `To begin the game and to make first move:
  * choose a cell
  * click Start at bottom of the screen`,
		//
		"ru-RU": `Чтобы начать игру и сделать первый ход:
  * выберите клетку
  * нажмите Start внизу экрана`,
	},
	MT_PLAYER: {
		"en-US": "Player <b>%v</b>:",
		"ru-RU": "Игрок <b>%v</b>:",
	},
	MT_AWAITING_PLAYER: {
		"en-US": "awaiting...",
		"ru-RU": "ожидается...",
	},
	MT_PLAYER_BALANCE: {
		"en-US": "Balance: %d",
		"ru-RU": "Баланс: %d",
	},
	MT_ASK_TO_RATE: {
		"en-US": `🙋 <b>Did you like the game?</b> We'll appreciate if you <a href="https://t.me/storebot?start=BiddingTicTacToeBot">rate us</a> with 5⭐s!'`,
		"ru-RU": `🙋 <b>Понравилась игра?</b> Будем признательны если <a href="https://t.me/storebot?start=BiddingTicTacToeBot">оцените нас</a> на 5⭐!`,
	},
	// MT_FREE_GAME_SPONSORED_BY: {
	// 	"en-US": "🙏 Free game sponsored by %v",
	// 	"ru-RU": "🙏 Бесплатная игра при поддержке %v - бот для учёта долгов",
	// },
}
