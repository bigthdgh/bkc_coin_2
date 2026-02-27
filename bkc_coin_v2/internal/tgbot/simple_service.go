package tgbot

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SimpleService struct {
	bot *tgbotapi.BotAPI
}

func NewSimpleService(bot *tgbotapi.BotAPI) *SimpleService {
	return &SimpleService{
		bot: bot,
	}
}

func (s *SimpleService) StartSimple() {
	// Установка команд
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	// Обработка обновлений
	for update := range updates {
		if update.Message != nil {
			s.handleMessage(update.Message)
		}
	}
}

func (s *SimpleService) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName

	log.Printf("Message from %s (%d): %s", username, userID, message.Text)

	// Обработка команд
	switch message.Command() {
	case "start":
		s.sendStartMessage(userID)
	case "balance":
		s.sendBalanceMessage(userID)
	case "tap":
		s.sendTapMessage(userID)
	case "energy":
		s.sendEnergyMessage(userID)
	case "p2p":
		s.sendP2PMessage(userID)
	case "nft":
		s.sendNFTMessage(userID)
	case "help":
		s.sendHelpMessage(userID)
	default:
		if message.Text != "" {
			s.sendDefaultMessage(userID)
		}
	}
}

func (s *SimpleService) sendStartMessage(userID int64) {
	text := `🎮 Добро пожаловать в BKC Coin!

💰 Твой баланс: 1,000 BKC
⚡ Энергия: 300/300
👆 Тапай чтобы заработать!

📊 Текущий курс: 1000 BKC = $1.00
🎯 NFT Bronze: 30,000 BKC ($30)

📋 Доступные команды:
/balance - 💰 Баланс
/tap - 👆 Тапать
/energy - ⚡ Энергия
/p2p - 📈 P2P маркетплейс
/nft - 🖼️ NFT магазин
/help - ❓ Помощь`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getMainKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendBalanceMessage(userID int64) {
	text := `💰 Твой баланс:

🪙 BKC: 1,000.00
💵 USD: $1.00
🔷 TON: 0.70

📊 Курсы:
1000 BKC = $1.00
1 TON = $1.43`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getMainKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendTapMessage(userID int64) {
	text := `👆 Тап!

💰 Заработано: +1 BKC
⚡ Энергия: 299/300
🔥 Комбо: x1

👆 Продолжай тапать!`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getTapKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendEnergyMessage(userID int64) {
	text := `⚡ Энергия:

🔋 Текущая: 299/300
⏱️ Восстановление: 1 в секунду
🔋 Максимум: 300

⚡ Подожди полного восстановления!`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getMainKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendP2PMessage(userID int64) {
	text := `📈 P2P Маркетплейс

🔥 Популярные ордера:
📈 BUY: 1000 BKC @ $1.00
📉 SELL: 500 BKC @ $0.99

💹 Комиссия: 3%
🔒 Escrow: Защита сделок

📊 График цен: /chart
🛍️ Создать ордер: /create_order`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getP2PKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendNFTMessage(userID int64) {
	text := `🖼️ NFT Магазин

🥉 Bronze NFT - 30,000 BKC ($30)
   • +10% к тапам
   • +50 энергии
   
🥈 Silver NFT - 80,000 BKC ($80)
   • +25% к тапам
   • +150 энергии
   
🥇 Gold NFT - 300,000 BKC ($300)
   • +50% к тапам
   • +300 энергии

💳 Оплата: BKC или TON
📈 Инвестиция в будущее!`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getNFTKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendHelpMessage(userID int64) {
	text := `❓ Помощь - BKC Coin Bot

📋 Основные команды:
/start - 🎮 Начать игру
/balance - 💰 Показать баланс
/tap - 👆 Тапать монету
/energy - ⚡ Энергия
/p2p - 📈 P2P маркетплейс
/nft - 🖼️ NFT магазин
/help - ❓ Эта помощь

🎯 Цель игры:
Собирай BKC, торгуй на P2P, покупай NFT, развивай свою крипто-империю!

💰 Твой ID: ` + strconv.FormatInt(userID, 10) + `

🔗 Поддержка: @bkc_support

🚀 Удачи в игре!`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getMainKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) sendDefaultMessage(userID int64) {
	text := `👆 Тап!

💰 Заработано: +1 BKC
⚡ Энергия: 298/300
🔥 Комбо: x1

👆 Продолжай тапать!`

	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = s.getTapKeyboard()

	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (s *SimpleService) getMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Баланс", "balance"),
			tgbotapi.NewInlineKeyboardButtonData("⚡ Энергия", "energy"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 P2P", "p2p"),
			tgbotapi.NewInlineKeyboardButtonData("🖼️ NFT", "nft"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
		),
	)
	return keyboard
}

func (s *SimpleService) getTapKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👆 Тапнуть!", "tap"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ Энергия", "energy"),
			tgbotapi.NewInlineKeyboardButtonData("💰 Баланс", "balance"),
		),
	)
	return keyboard
}

func (s *SimpleService) getP2PKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 График цен", "chart"),
			tgbotapi.NewInlineKeyboardButtonData("🛍️ Создать ордер", "create_order"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Мои ордера", "my_orders"),
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_main"),
		),
	)
	return keyboard
}

func (s *SimpleService) getNFTKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🥉 Bronze (30K BKC)", "buy_bronze"),
			tgbotapi.NewInlineKeyboardButtonData("🥈 Silver (80K BKC)", "buy_silver"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🥇 Gold (300K BKC)", "buy_gold"),
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_main"),
		),
	)
	return keyboard
}
