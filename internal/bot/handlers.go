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
			return editTournamentsPage(c, tournaments, 0)
		case ListTournaments:
			tournaments, err := bt.svc.GetTournaments()
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return allTournamentsPage(c, tournaments, 0)
		}

		if strings.HasPrefix(data, "myts_page_") {
			pageStr := strings.TrimPrefix(data, "myts_page_")
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 0 {
				page = 0
			}
			tournaments, err := bt.svc.GetUserTournaments(userID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return editTournamentsPage(c, tournaments, page)
		}

		if strings.HasPrefix(data, "allts_page_") {
			pageStr := strings.TrimPrefix(data, "allts_page_")
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 0 {
				page = 0
			}

			list, err := bt.svc.GetTournaments()
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return allTournamentsPage(c, list, page)
		}

		if strings.HasPrefix(data, "app_approve_") {
			parts := strings.Split(strings.TrimPrefix(data, "app_approve_"), "_")
			if len(parts) != 2 {
				return c.Send("Ошибка: Некорректные данные кнопки")
			}

			tID64, err1 := strconv.Atoi(parts[0])
			tgID64, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				return c.Send("Ошибка: Некорректный формат tgID или tournamentID")
			}

			tID := domain.TournamentID(tID64)
			tgID := domain.TelegramUserID(tgID64)

			app, err := bt.svc.GetApplication(tID, tgID)
			if err != nil {
				return c.Send("Ошибка при получении заявки: " + err.Error())
			}
			t, _ := bt.svc.GetTournament(tID)

			tag := ""
			if app.TelegramTag != nil {
				tag = "(@" + *app.TelegramTag + ")"
			}

			if err := bt.svc.ApproveApplication(tID, tgID); err != nil {
				return c.Send("Ошибка при одобрении: " + err.Error())
			}

			return c.Edit(
				fmt.Sprintf("✅ Заявка пользователя %s %s на турнир %s одобрена!", app.Name, tag, t.Title),
				mainMenu(),
			)
		}

		if strings.HasPrefix(data, "app_reject_") {
			parts := strings.Split(strings.TrimPrefix(data, "app_reject_"), "_")
			if len(parts) != 2 {
				return c.Send("Ошибка: Некорректные данные кнопки")
			}

			tID64, err1 := strconv.ParseInt(parts[0], 10, 64)
			tgID64, err2 := strconv.ParseInt(parts[1], 10, 64)
			if err1 != nil || err2 != nil {
				return c.Send("Некорректный формат ID")
			}

			tID := domain.TournamentID(tID64)
			tgID := domain.TelegramUserID(tgID64)

			app, err := bt.svc.GetApplication(tID, tgID)
			if err != nil {
				return c.Send("Ошибка при получении заявки: " + err.Error())
			}
			t, _ := bt.svc.GetTournament(tID)

			tag := ""
			if app.TelegramTag != nil {
				tag = "(@" + *app.TelegramTag + ")"
			}

			if err := bt.svc.RejectApplication(tID, tgID); err != nil {
				return c.Send("Ошибка при отклонении: " + err.Error())
			}

			return c.Edit(fmt.Sprintf("❌ Заявка пользователя %s @%s на турнир с %s отклонена!", app.Name, tag, t.Title), mainMenu())
		}

		if data == ApplySkipText {
			ac := bt.getApply(userID)
			if ac == nil {
				bt.setState(userID, StateMainMenu)
				return c.Edit("Сессия сброшена. Выберите действие", mainMenu())
			}
			app := &domain.Application{
				TournamentID:   ac.TournamentID,
				TelegramUserID: userID,
				Name:           ac.Name,
				TelegramTag:    &c.Sender().Username,
				Text:           nil,
			}
			if err := bt.svc.ApplyToTournament(app); err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			bt.clearApply(userID)
			bt.setState(userID, StateMainMenu)
			_ = c.Edit("Заявка отправлена! Администратор скоро её рассмотрит.")
			return c.Send("Выберите действие", mainMenu())
		}

		if strings.HasPrefix(data, "tournament_") {
			idStr := strings.TrimPrefix(data, "tournament_")
			tID, err := strconv.Atoi(idStr)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			t, err := bt.svc.GetTournament(domain.TournamentID(tID))
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			tgID := domain.TelegramUserID(c.Sender().ID)
			if t.OwnerID == tgID {
				return tournamentMenu(c, t)
			}
			if t.UserParticipates(tgID) {
				return c.Send("МЕНЮ УЧАСТНИКА")
			}

			bt.setState(tgID, StateApplyEnterName)
			app, err := bt.svc.GetApplication(t.ID, tgID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			if app != nil {
				bt.setState(tgID, StateMainMenu)
				return c.Edit(fmt.Sprintf("Вы уже подали заявку на участие в турнире «%s». Ожидайте решения администратора.", t.Title))
			}
			bt.setApply(tgID, &applyCtx{TournamentID: t.ID})
			return c.Edit(fmt.Sprintf(
				"Хотите участвовать в турнире «%s»?\n\nНапишите, пожалуйста, своё имя:",
				t.Title,
			))
		}

		if strings.HasPrefix(data, "applications_") {
			tIDStr := strings.TrimPrefix(data, "applications_tournament")
			tID64, err := strconv.Atoi(tIDStr)
			tID := domain.TournamentID(tID64)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			apps, err := bt.svc.GetApplications(tID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			t, _ := bt.svc.GetTournament(tID)
			return c.Edit(applicationMenu(c, t, apps, 0))
		}

		if strings.HasPrefix(data, "start_tournament") {
			tIDStr := strings.TrimPrefix(data, "start_tournament")
			tID64, err := strconv.Atoi(tIDStr)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			tID := domain.TournamentID(tID64)
			err = bt.svc.StartTournament(tID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return c.Send(fmt.Sprintf("Турнир ID %d начат!", tID))
		}

		if strings.HasPrefix(data, "information_tournament") {
			tIDStr := strings.TrimPrefix(data, "information_tournament")
			tID64, err := strconv.Atoi(tIDStr)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			tID := domain.TournamentID(tID64)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return c.Send(fmt.Sprintf("Турнир ID %d - что с ним делать?", tID))
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

			return sendTournamentsPage(c, tournaments, 0)
		case StateApplyEnterName:
			name := strings.TrimSpace(c.Text())
			if name == "" {
				return c.Send("Имя не может быть пустым. Введите ещё раз.")
			}
			ac := bt.getApply(userID)
			if ac == nil {
				bt.setState(userID, StateMainMenu)
				return c.Send("Сессия сброшена. Выберите действие", mainMenu())
			}
			ac.Name = name
			bt.setState(userID, StateApplyEnterText)

			menu := &tb.ReplyMarkup{}
			btnSkip := menu.Data("Пропустить", ApplySkipText)
			btnMain := menu.Data("Отменить", MainMenu)
			menu.Inline(menu.Row(btnSkip, btnMain))

			return c.Send("Напишите текст заявки, либо нажмите «Пропустить».", menu)

		case StateApplyEnterText:
			text := strings.TrimSpace(c.Text())
			if text == "" {
				return c.Send("Можно написать пару слов или нажать «Пропустить».")
			}
			ac := bt.getApply(userID)
			if ac == nil {
				bt.setState(userID, StateMainMenu)
				return c.Send("Сессия сброшена. Выберите действие", mainMenu())
			}

			app := &domain.Application{
				TournamentID:   ac.TournamentID,
				TelegramUserID: userID,
				Name:           ac.Name,
				TelegramTag:    &c.Sender().Username,
				Text:           &text,
			}
			if err := bt.svc.ApplyToTournament(app); err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			bt.clearApply(userID)
			bt.setState(userID, StateMainMenu)
			_ = c.Send("Заявка отправлена! Администратор скоро её рассмотрит.")
			return c.Send("Выберите действие", mainMenu())
		}

		return c.Send("Не распознал команду")
	})

}
