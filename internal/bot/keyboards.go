package bot

import (
	"fmt"
	"log"
	"sort"

	"github.com/Ycyken/tournament-bot/internal/domain"
	tb "gopkg.in/telebot.v3"
)

const pageSize = 5

const MainMenu = "main_menu"
const CreateTournament = "create_tournament"
const ListTournaments = "list_tournaments"
const MyTournaments = "my_tournaments"

func mainMenu() *tb.ReplyMarkup {
	menu := &tb.ReplyMarkup{}
	btnList := menu.Data("Посмотреть список турниров", ListTournaments)
	btnMyTs := menu.Data("Мои турниры", MyTournaments)
	menu.Inline(menu.Row(btnList), menu.Row(btnMyTs))
	return menu
}

func tournamentsPage(c tb.Context, tournaments map[domain.TournamentID]string, page int) error {
	log.Printf("number of tournaments is %d", len(tournaments))
	menu := &tb.ReplyMarkup{}
	var rows []tb.Row

	var ids []domain.TournamentID
	for id := range tournaments {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	start := page * pageSize
	end := start + pageSize
	if start >= len(ids) {
		start = 0
		end = pageSize
	}
	if end > len(ids) {
		end = len(ids)
	}

	for _, id := range ids[start:end] {
		btn := menu.Data(fmt.Sprintf("%s | ID %d", tournaments[id], id), fmt.Sprintf("tournament_%d", id))
		rows = append(rows, menu.Row(btn))
	}

	var navBtns []tb.Btn
	if start > 0 {
		navBtns = append(navBtns, menu.Data("<< Назад", fmt.Sprintf("page_%d", page-1)))
	}
	if end < len(ids) {
		navBtns = append(navBtns, menu.Data("Вперёд >>", fmt.Sprintf("page_%d", page+1)))
	}
	if len(navBtns) > 0 {
		rows = append(rows, menu.Row(navBtns...))
	}

	btnCreate := menu.Data("Создать турнир", CreateTournament)
	btnMain := menu.Data("Главное меню", MainMenu)
	rows = append(rows, menu.Row(btnCreate, btnMain))

	menu.Inline(rows...)
	return c.Edit("Ваши турниры:", menu)
}

func tournamentMenu(c tb.Context, t *domain.Tournament) error {
	menu := &tb.ReplyMarkup{}
	tgID := domain.TelegramUserID(c.Sender().ID)

	btnMain := menu.Data("Главное меню", MainMenu)
	if tgID != t.OwnerID {
		menu.Inline(menu.Row(btnMain))
		return c.Send("какая-то инфа о турнире", menu)
	}
	var rows []tb.Row
	btnApps := menu.Data("Заявки на турнир", fmt.Sprintf("applications_tournament%d", t.ID))
	btnStart := menu.Data("Начать турнир", fmt.Sprintf("start_tournament%d", t.ID))
	btnInfo := menu.Data("Информация о турнире", fmt.Sprintf("information_tournament%d", t.ID))

	if t.CurrentRound == 0 {
		rows = append(rows, menu.Row(btnApps), menu.Row(btnStart))
	}
	rows = append(rows, menu.Row(btnInfo), menu.Row(btnMain))
	menu.Inline(rows...)
	return c.Edit(fmt.Sprintf("Турнир %s | ID %d", t.Title, t.ID), menu)
}

func applicationMenu(c tb.Context, tID domain.TournamentID, apps []*domain.Application, idx int) error {
	menu := &tb.ReplyMarkup{}

	btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", tID))
	if len(apps) == 0 {
		menu.Inline(menu.Row(btnBack))
		return c.Edit("Заявок нет", menu)
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(apps) {
		idx = len(apps) - 1
	}
	app := apps[idx]

	title := fmt.Sprintf(
		"Заявка от участника с tgID: %d",
		app.TelegramUserID,
	)

	btnApprove := menu.Data("✅ Принять", fmt.Sprintf("app_approve_%d_%d", app.TournamentID, app.TelegramUserID))
	btnReject := menu.Data("❌ Отклонить", fmt.Sprintf("app_reject_%d_%d", app.TournamentID, app.TelegramUserID))

	menu.Inline(
		menu.Row(btnApprove, btnReject),
		menu.Row(btnBack),
	)
	return c.Edit(title, menu)
}
