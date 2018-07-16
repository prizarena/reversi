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
	//Flips: {
	//	"en-US": "<b>Turns</b>: %v",
	//	"ru-RU": "<b>Ходов</b>: %v",
	//},
	GameCompleted: {
		//--------------------------------------------------------------------------------------------------------------
		"en-US": `Game is completed.

If you liked this bot please <a href="https://t.me/storebot?start=reversigamebot">give us 5 stars</a>! We appreciate your feedback 🙏

<i>P.S. Support game development by subscribing to @prizarena channel.</i>`,
		//--------------------------------------------------------------------------------------------------------------
		"ru-RU": `Игра окончена.

Если вам понравился наш бот пожалуйста <a href="https://t.me/storebot?start=reversigamebot">поставьте нам 5 звёзд</a>! Мы будем очень признательны 🙏

<i>P.S. Поддержите разработку игры подписашись на канал @prizarena.</i>` ,
		//--------------------------------------------------------------------------------------------------------------
		"de-DE": `Das Spiel ist abgeschlossen.

Wenn Ihnen dieser Bot gefallen hat, <a href="https://t.me/storebot?start=reversigamebot">geben Sie uns 5 Sterne</a>! Wir freuen uns über Ihr Feedback 🙏

<i> P.S. Unterstütze die Entwicklung von Spielen, indem du @prizarena abonnierst.</i>`,
		//--------------------------------------------------------------------------------------------------------------
		"es-ES": `El juego está completo.

Si te gustó este robot,  ¡<a href="https://t.me/storebot?start=reversigamebot">danos 5 estrellas </a>! Agradecemos sus comentarios 🙏

<i>P.S. Apoya el desarrollo de juegos suscribiéndote al canal @prizarena.</i>`,
		//--------------------------------------------------------------------------------------------------------------
		//"fa-IR": ``,
		//--------------------------------------------------------------------------------------------------------------
		"fr-FR": `Le jeu est terminé.

Si vous avez aimé ce bot s'il vous plaît <a href="https://t.me/storebot?start=reversigamebot">donnez-nous 5 étoiles</a>! Nous apprécions vos commentaires 🙏

<i>P.S. Soutenez le développement de jeux en vous abonnant au canal @prizarena.</i>`,
		//--------------------------------------------------------------------------------------------------------------
		"it-IT": `Il gioco è completato.

Se ti è piaciuto questo bot, ti preghiamo di <a href="https://t.me/storebot?start=reversigamebot">darci 5 stelle</a>! Apprezziamo il tuo feedback 🙏

<i> P.S. Supporta lo sviluppo del gioco iscrivendoti al canale @prizarena.</i>`,
		//--------------------------------------------------------------------------------------------------------------
	},
	NewGameInlineTitle: {
		"en-US": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - new game",
		"ru-RU": emoji.BlackCircle + emoji.WhiteCircle + " Реверси - новая игра",
		"de-DE": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - neues Spiel",
		"es-ES": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - juego nuevo",
		"fa-IR": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - بازی جدید",
		"fr-FR": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - nouveau jeu",
		"it-IT": emoji.BlackCircle + emoji.WhiteCircle + " Reversi - nuovo gioco",
	},
	NewGameInlineDescription: {
		"en-US": "Starts a new Reversi game",
		"ru-RU": "Создать новую игру Реверси",
		"de-DE": "Startet ein neues Reversi-Spiel",
		"es-ES": "Inicia un nuevo juego de Reversi",
		"fa-IR": "یک بازی Reversi جدید شروع می کند",
		"fr-FR": "Commence un nouveau jeu Reversi",
		"it-IT": "Inizia una nuova partita di Reversi",
	},
	GameCardTitle: {
		"en-US": "Reversi game",
		"ru-RU": "Игра: Реверси",
		"de-DE": "Reversi Spiel",
		"es-ES": "Reversi juego",
		"fa-IR": "بازی Reversi",
		"fr-FR": "Reversi jeu",
		"it-IT": "Gioco Reversi",
	},
	Me: {
		"en-US": "me",
		"ru-RU": "я",
		"de-DE": "mich",
		"es-ES": "yo",
		"fa-IR": "من",
		"fr-FR": "moi",
		"it-IT": "me",
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
	//Board: {
	//	"en-US": "Board",
	//	"ru-RU": "Доска",
	//},
	//Tournaments: {
	//	"en-US": "🏆 Tournaments",
	//	"ru-RU": "🏆 Турниры",
	//},
	FirstMoveDoneAwaitingSecond: {
		"en-US": "Player <b>%v</b> made choice, awaiting another player...",
		"ru-RU": "Игрок <b>%v</b> сделал свой ход, ожидается ход второго игрока...",
	},
	NewGameText: {
		"en-US": `<b>Reversi game</b>

Blacks make 1st move.`,
		"ru-RU": `Игра: <b>Реверси</b>

Чёрные ходят первыми.`,
		"de-DE": `<b>Reversi Spiel</b>

Schwarze machen den ersten Zug.`,
		"es-ES": `<b>Juego Reversi</b>

Los negros hacen primer movimiento.`,
		"fa-IR": `<b>بازی Reversi</b>

سیاه پوستان، حرکت اول را انجام می دهند.`,
		"fr-FR": `<b>Jeu Reversi</b>

Les Noirs font le premier mouvement.`,
		"it-IT": ``,
	},
}
