package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Comics struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Score int    `json:"score"`
}

type SearchResponse struct {
	Comics []Comics `json:"comics"`
	Total  int      `json:"total"`
}

func HandleSearchCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, apiAddress string, log *slog.Logger) {
	chatID := update.Message.Chat.ID

	args := strings.TrimSpace(update.Message.CommandArguments())
	if args == "" {
		msg := tgbotapi.NewMessage(chatID, "Использование: /search <фраза для поиска>\n\nПример: /search quantum physics")
		bot.Send(msg)
		return
	}

	// Send "searching" message
	searchingMsg := tgbotapi.NewMessage(chatID, "🔍 Ищу комиксы...")
	searchingMsgSent, _ := bot.Send(searchingMsg)

	apiURL := fmt.Sprintf("%s/api/search?phrase=%s&limit=10", apiAddress, url.QueryEscape(args))
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Error("failed to call API", "error", err)
		errorMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, "❌ Ошибка подключения к сервису поиска")
		bot.Send(errorMsg)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("failed to read response", "error", err)
			errorMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, "❌ Ошибка чтения результатов")
			bot.Send(errorMsg)
			return
		}

		var searchResp SearchResponse
		if err := json.Unmarshal(body, &searchResp); err != nil {
			log.Error("failed to parse response", "error", err)
			errorMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, "❌ Ошибка обработки результатов")
			bot.Send(errorMsg)
			return
		}

		if len(searchResp.Comics) == 0 {
			noResultsMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, "😔 Ничего не найдено. Попробуйте изменить поисковый запрос.")
			bot.Send(noResultsMsg)
			return
		}

		// Format results
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("✅ Найдено комиксов: %d\n\n", searchResp.Total))

		for i, comic := range searchResp.Comics {
			if i >= 5 { // Limit to 5 results in message
				resultText.WriteString(fmt.Sprintf("\n... и еще %d результатов", len(searchResp.Comics)-5))
				break
			}
			resultText.WriteString(fmt.Sprintf("🎴 <b>#%d</b> (Score: %d)\n", comic.ID, comic.Score))
			resultText.WriteString(fmt.Sprintf("🔗 <a href=\"%s\">Открыть комикс</a>\n\n", comic.URL))
		}

		// Send results
		resultMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, resultText.String())
		resultMsg.ParseMode = "HTML"
		resultMsg.DisableWebPagePreview = false
		bot.Send(resultMsg)

	} else if resp.StatusCode == http.StatusNotFound {
		noResultsMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, "😔 Ничего не найдено. Попробуйте изменить поисковый запрос.")
		bot.Send(noResultsMsg)
	} else {
		body, _ := io.ReadAll(resp.Body)
		log.Error("API returned error", "status", resp.StatusCode, "body", string(body))
		errorMsg := tgbotapi.NewEditMessageText(chatID, searchingMsgSent.MessageID, "❌ Ошибка сервиса поиска")
		bot.Send(errorMsg)
	}
}

func HandleStartCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userName := update.Message.From.FirstName

	welcomeText := fmt.Sprintf(
		"👋 Привет, %s!\n\n"+
			"Я бот для поиска комиксов XKCD.\n\n"+
			"<b>Доступные команды:</b>\n"+
			"/search <фраза> - поиск комиксов\n"+
			"/help - показать справку\n\n"+
			"<b>Пример:</b>\n"+
			"/search quantum physics",
		userName,
	)

	msg := tgbotapi.NewMessage(chatID, welcomeText)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func HandleHelpCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	helpText := "📖 <b>Справка по использованию бота</b>\n\n" +
		"<b>Команды:</b>\n" +
		"/search <фраза> - поиск комиксов по ключевым словам\n" +
		"/help - показать эту справку\n\n" +
		"<b>Примеры:</b>\n" +
		"/search quantum\n" +
		"/search physics\n" +
		"/search computer science\n\n" +
		"Бот найдет комиксы, содержащие указанные ключевые слова, и покажет до 10 результатов с оценкой релевантности."

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func HandleUnknownCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	msg := tgbotapi.NewMessage(chatID, "❓ Неизвестная команда. Используйте /help для списка доступных команд.")
	bot.Send(msg)
}
