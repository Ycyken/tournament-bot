// `internal/bot/handlers.go`
package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Ycyken/tournament-bot/internal/domain"
	tb "gopkg.in/telebot.v3"
)

func registerHandlers(bt *Bot) {
	bot := bt.bot
	bot.Handle("/start", func(c tb.Context) error {
		return c.Send("Выберите действие", mainMenu())
	})

	bot.Handle(tb.OnCallback, func(c tb.Context) error {
		_ = c.Respond(&tb.CallbackResponse{})

		userID := domain.TelegramUserID(c.Sender().ID)
		data := strings.TrimSpace(c.Callback().Data)
		switch data {
		case MainMenu:
			bt.setState(userID, StateMainMenu)
			return c.Edit("Выберите действие", mainMenu())
		case CreateTournament:
			bt.setState(userID, StateWaitingTournamentTitle)
			return c.Edit("Введите название турнира:")
		case MyTournaments:
			tournaments, err := bt.svc.GetUserTournaments(userID)
			log.Println(len(tournaments))
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return tournamentsPage(bt, c, tournaments, 0)
		case ListTournaments:
			tournaments, err := bt.svc.GetTournaments()
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			var strIDs []string
			for _, t := range tournaments {
				strIDs = append(strIDs, fmt.Sprintf("%s | ID %d", t.Title, t.ID))
			}

			result := strings.Join(strIDs, "\n")
			return c.Send(result)
		}

		if strings.HasPrefix(data, "page_") {
			pageStr := strings.TrimPrefix(data, "page_")
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 0 {
				page = 0
			}
			tournaments, err := bt.svc.GetUserTournaments(userID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return tournamentsPage(bt, c, tournaments, page)
		}

		if strings.HasPrefix(data, "tournament_") {
			idStr := strings.TrimPrefix(data, "tournament_")
			tID, err := strconv.Atoi(idStr)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			return c.Edit(fmt.Sprintf("Турнир ID %d — что сделать?", tID))
		}

		return nil
	})

	bot.Handle(tb.OnText, func(c tb.Context) error {
		userID := domain.TelegramUserID(c.Sender().ID)
		state := bt.getState(userID)

		switch state {
		case StateWaitingTournamentTitle:
			title := c.Text()
			if strings.TrimSpace(title) == "" {
				return c.Send("Название не может быть пустым. Попробуйте ещё раз.")
			}
			t, err := bt.svc.CreateTournament(userID, title)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			bt.setState(userID, StateTournamentManagement)
			_ = c.Send(fmt.Sprintf("Турнир '%s' создан! ID: %d", t.Title, t.ID))

			tournaments, err := bt.svc.GetUserTournaments(userID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			return tournamentsPage(bt, c, tournaments, 0)

		case StateReviewApplications:
			// Здесь логика обработки заявок
		}

		return c.Send("Не распознал команду")
	})

}
