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

			_ = c.Send(fmt.Sprintf("✅ Заявка пользователя %s%s на турнир «%s» одобрена!", app.Name, tag, t.Title))

			apps, err := bt.svc.GetApplications(tID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return applicationMenu(c, t, apps)
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

			_ = c.Send(fmt.Sprintf("❌ Заявка пользователя %s%s на турнир «%s» отклонена.", app.Name, tag, t.Title))

			apps, err := bt.svc.GetApplications(tID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}
			return applicationMenu(c, t, apps)
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
				return participantMenu(c, t, tgID)
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

		if strings.HasPrefix(data, "applications_tournament") {
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

			applicationMenu(c, t, apps)
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
			t, _ := bt.svc.GetTournament(tID)

			return c.Edit(participantMenu(c, t, userID))
		}

		if strings.HasPrefix(data, "pinfo_tournament") {
			tIDStr := strings.TrimPrefix(data, "pinfo_tournament")
			tID64, err := strconv.Atoi(tIDStr)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			t, err := bt.svc.GetTournament(domain.TournamentID(tID64))
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			info := TournamentInfoMessage(t)

			menu := &tb.ReplyMarkup{}
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", t.ID))
			menu.Inline(menu.Row(btnBack))
			return c.Edit(info, menu)
		}

		if strings.HasPrefix(data, "pmatches_tournament") {
			rest := strings.TrimPrefix(data, "pmatches_tournament")
			parts := strings.Split(rest, "_")
			if len(parts) != 2 {
				return c.Send("Некорректные данные кнопки")
			}
			tID64, err1 := strconv.Atoi(parts[0])
			tgID64, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				return c.Send("Некорректный формат ID")
			}
			tID := domain.TournamentID(tID64)
			tgID := domain.TelegramUserID(tgID64)

			t, err := bt.svc.GetTournament(tID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			p := t.FindParticipantBytgID(tgID)
			text := t.GetMatchesHistory(p.ID)

			menu := &tb.ReplyMarkup{}
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", t.ID))

			matches := t.GetParticipantMatches(p.ID)
			if m := matches[len(matches)-1]; m != nil {
				winRes, loseRes := "p1", "p2"
				if m.P1 != p.ID {
					winRes, loseRes = "p2", "p1"
				}
				btnWin := menu.Data("✅ Я выиграл", fmt.Sprintf("pmatch_report_%d_%d_%s", t.ID, m.ID, winRes))
				btnDraw := menu.Data("🤝 Ничья", fmt.Sprintf("pmatch_report_%d_%d_draw", t.ID, m.ID))
				btnLose := menu.Data("❌ Я проиграл", fmt.Sprintf("pmatch_report_%d_%d_%s", t.ID, m.ID, loseRes))
				menu.Inline(menu.Row(btnWin, btnDraw, btnLose), menu.Row(btnBack))
			} else {
				menu.Inline(menu.Row(btnBack))
			}
			if err := c.Edit(text, menu); err != nil {
				return c.Send(text, menu)
			}
			return c.Edit(text, menu)
		}

		if strings.HasPrefix(data, "pmatch_report_") {
			rest := strings.TrimPrefix(data, "pmatch_report_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Send("Некорректные данные кнопки")
			}
			tID64, err1 := strconv.ParseInt(parts[0], 10, 64)
			mID64, err2 := strconv.ParseInt(parts[1], 10, 64)
			res := parts[2] // "p1" | "p2" | "draw"
			if err1 != nil || err2 != nil || (res != "p1" && res != "p2" && res != "draw") {
				return c.Send("Некорректные данные результата")
			}

			tID := domain.TournamentID(tID64)
			mID := domain.MatchID(mID64)
			uid := domain.TelegramUserID(c.Sender().ID)

			t, err := bt.svc.ReportMatchResult(tID, mID, uid, domain.ResultType(res))
			if err != nil {
				return c.Send("Ошибка при отправке результата: " + err.Error())
			}

			_ = c.Send("✅ Ваш отчёт о матче принят.")

			p := t.FindParticipantBytgID(uid)
			text := t.GetMatchesHistory(p.ID)

			menu := &tb.ReplyMarkup{}
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", t.ID))
			menu.Inline(menu.Row(btnBack))

			if err := c.Edit(text, menu); err != nil {
				return c.Send(text, menu)
			}
			return nil
		}
		if strings.HasPrefix(data, "adm_matches_tournament") {
			tIDStr := strings.TrimPrefix(data, "adm_matches_tournament")
			tID64, err := strconv.ParseInt(tIDStr, 10, 64)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			tID := domain.TournamentID(tID64)

			t, err := bt.svc.GetTournament(tID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			text := buildCurrentMatchesText(t)
			menu := &tb.ReplyMarkup{}
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", t.ID))
			btnSet := menu.Data("✏️ Изменить результат", fmt.Sprintf("adm_setresult_%d", t.ID))
			menu.Inline(menu.Row(btnSet), menu.Row(btnBack))
			return c.Edit(text, menu)
		}

		if strings.HasPrefix(data, "adm_setresult_") {
			tIDStr := strings.TrimPrefix(data, "adm_setresult_")
			tID64, err := strconv.ParseInt(tIDStr, 10, 64)
			if err != nil {
				return c.Send("Некорректный ID турнира")
			}
			tID := domain.TournamentID(tID64)

			adminID := domain.TelegramUserID(c.Sender().ID)
			bt.setState(adminID, StateAdminAwaitMatchID)
			bt.setAdminCtx(adminID, &adminSetResultCtx{TournamentID: tID})

			menu := &tb.ReplyMarkup{}
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("adm_matches_tournament%d", tID))
			menu.Inline(menu.Row(btnBack))
			return c.Edit("Введите ID матча (число):", menu)
		}

		if strings.HasPrefix(data, "adm_apply_result_") {
			rest := strings.TrimPrefix(data, "adm_apply_result_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Send("Некорректные данные кнопки")
			}

			tID64, err1 := strconv.ParseInt(parts[0], 10, 64)
			mID64, err2 := strconv.ParseInt(parts[1], 10, 64)
			res := parts[2]
			if err1 != nil || err2 != nil || (res != "p1" && res != "p2" && res != "draw") {
				return c.Send("Некорректные данные результата")
			}
			tID := domain.TournamentID(tID64)
			mID := domain.MatchID(mID64)
			adminID := domain.TelegramUserID(c.Sender().ID)

			t, err := bt.svc.SetMatchResultByAdmin(tID, mID, adminID, domain.ResultType(res))
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			_ = c.Send("✅ Результат матча обновлён администратором.")

			text := buildCurrentMatchesText(t)
			menu := &tb.ReplyMarkup{}
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", t.ID))
			btnSet := menu.Data("✏️ Изменить результат", fmt.Sprintf("adm_setresult_%d", t.ID))
			menu.Inline(menu.Row(btnSet), menu.Row(btnBack))
			if err := c.Edit(text, menu); err != nil {
				return c.Send(text, menu)
			}
			return nil
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
		case StateAdminAwaitMatchID:
			adminID := domain.TelegramUserID(c.Sender().ID)
			ctx := bt.getAdminCtx(adminID)
			if ctx == nil {
				bt.setState(adminID, StateMainMenu)
				return c.Send("Сессия админа сброшена.", mainMenu())
			}

			mID64, err := strconv.ParseInt(strings.TrimSpace(c.Text()), 10, 64)
			if err != nil {
				return c.Send("Нужно число — ID матча. Попробуйте ещё раз.")
			}
			mID := domain.MatchID(mID64)

			t, err := bt.svc.GetTournament(ctx.TournamentID)
			if err != nil {
				return c.Send("Ошибка: " + err.Error())
			}

			var match *domain.Match
			for _, ms := range t.Matches {
				for _, m := range ms {
					if m.ID == mID {
						match = m
						break
					}
				}
				if match != nil {
					break
				}
			}
			if match == nil {
				return c.Send("Матч с таким ID не найден. Введите другой ID.")
			}

			// рисуем кнопки результата
			menu := &tb.ReplyMarkup{}
			btnP1 := menu.Data("🏆 Выиграл первый", fmt.Sprintf("adm_apply_result_%d_%d_p1", t.ID, mID))
			btnDraw := menu.Data("🤝 Ничья", fmt.Sprintf("adm_apply_result_%d_%d_draw", t.ID, mID))
			btnP2 := menu.Data("🏆 Выиграл второй", fmt.Sprintf("adm_apply_result_%d_%d_p2", t.ID, mID))
			btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("adm_matches_tournament%d", t.ID))
			menu.Inline(menu.Row(btnP1, btnDraw, btnP2), menu.Row(btnBack))

			bt.setState(adminID, StateMainMenu) // выходим из режима ввода
			bt.clearAdminCtx(adminID)

			return c.Send(
				fmt.Sprintf("Матч ID %d:\nP1=%d, P2=%d.\nВыберите результат:", mID, match.P1, match.P2),
				menu,
			)
		}

		return c.Send("Не распознал команду")
	})
}

func buildCurrentMatchesText(t *domain.Tournament) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Текущие матчи турнира «%s» (Раунд %d):", t.Title, t.CurrentRound))
	
	name := func(id domain.ParticipantID) string {
		for _, p := range t.Participants {
			if p.ID == id {
				return p.Name
			}
		}
		return fmt.Sprintf("ID %d", id)
	}

	has := false
	if ms, ok := t.Matches[t.CurrentRound]; ok {
		for _, m := range ms {
			if m.State == domain.MatchCompleted {
				continue
			}
			has = true
			lines = append(lines, fmt.Sprintf("• #%d: %s vs %s", m.ID, name(m.P1), name(m.P2)))
		}
	}
	if !has {
		lines = append(lines, "— нет незавершённых матчей.")
	}
	return strings.Join(lines, "\n")
}
