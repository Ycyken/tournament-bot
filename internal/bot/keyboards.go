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
const ApplySkipText = "apply_skip_text"

func mainMenu() *tb.ReplyMarkup {
	menu := &tb.ReplyMarkup{}
	btnList := menu.Data("Посмотреть список турниров", ListTournaments)
	btnMyTs := menu.Data("Мои турниры", MyTournaments)
	menu.Inline(menu.Row(btnList), menu.Row(btnMyTs))
	return menu
}

func tournamentsPageView(c tb.Context, tournaments map[domain.TournamentID]string, page int) (string, *tb.ReplyMarkup) {
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
		navBtns = append(navBtns, menu.Data("<< Назад", fmt.Sprintf("myts_page_%d", page-1)))
	}
	if end < len(ids) {
		navBtns = append(navBtns, menu.Data("Вперёд >>", fmt.Sprintf("myts_page_%d", page+1)))
	}
	if len(navBtns) > 0 {
		rows = append(rows, menu.Row(navBtns...))
	}

	btnCreate := menu.Data("Создать турнир", CreateTournament)
	btnMain := menu.Data("Главное меню", MainMenu)
	rows = append(rows, menu.Row(btnCreate, btnMain))

	menu.Inline(rows...)
	return "Ваши турниры:", menu
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
	btnInfo := menu.Data("Информация о турнире", fmt.Sprintf("pinfo_tournament%d", t.ID))
	btnCur := menu.Data("Текущие матчи", fmt.Sprintf("adm_matches_tournament%d", t.ID))

	if t.CurrentRound == 0 {
		rows = append(rows, menu.Row(btnApps), menu.Row(btnStart))
	} else {
		rows = append(rows, menu.Row(btnCur))
	}
	rows = append(rows, menu.Row(btnInfo), menu.Row(btnMain))
	menu.Inline(rows...)
	return c.Edit(fmt.Sprintf("Турнир %s | ID %d", t.Title, t.ID), menu)
}

func sendTournamentsPage(c tb.Context, tournaments map[domain.TournamentID]string, page int) error {
	text, markup := tournamentsPageView(c, tournaments, page)
	return c.Send(text, markup)
}
func editTournamentsPage(c tb.Context, tournaments map[domain.TournamentID]string, page int) error {
	text, markup := tournamentsPageView(c, tournaments, page)
	return c.Edit(text, markup)
}

func allTournamentsPage(c tb.Context, ts []*domain.Tournament, page int) error {
	menu := &tb.ReplyMarkup{}
	var rows []tb.Row

	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })

	start := page * pageSize
	end := start + pageSize
	if start >= len(ts) {
		start = 0
		end = pageSize
	}
	if end > len(ts) {
		end = len(ts)
	}

	for _, t := range ts[start:end] {
		btn := menu.Data(fmt.Sprintf("%s | ID %d", t.Title, t.ID), fmt.Sprintf("tournament_%d", t.ID))
		rows = append(rows, menu.Row(btn))
	}

	var nav []tb.Btn
	if start > 0 {
		nav = append(nav, menu.Data("<< Назад", fmt.Sprintf("allts_page_%d", page-1)))
	}
	if end < len(ts) {
		nav = append(nav, menu.Data("Вперёд >>", fmt.Sprintf("allts_page_%d", page+1)))
	}
	if len(nav) > 0 {
		rows = append(rows, menu.Row(nav...))
	}

	rows = append(rows, menu.Row(menu.Data("Главное меню", MainMenu)))

	menu.Inline(rows...)
	return c.Edit("Список турниров:", menu)
}

func applicationMenu(c tb.Context, t *domain.Tournament, apps []*domain.Application) error {
	menu := &tb.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", fmt.Sprintf("tournament_%d", t.ID))

	if len(apps) == 0 {
		menu.Inline(menu.Row(btnBack))
		return c.Edit("Заявок нет", menu)
	}
	app := apps[0]

	text := ""
	tag := ""
	if app.Text != nil {
		text = *app.Text
	}
	if app.TelegramTag != nil {
		tag = "(@" + *app.TelegramTag + ")"
	}
	title := fmt.Sprintf(
		"Заявка от участника %s %s на турнир %s:\n\n%s",
		app.Name, tag, t.Title, text,
	)

	btnApprove := menu.Data("✅ Принять", fmt.Sprintf("app_approve_%d_%d", app.TournamentID, app.TelegramUserID))
	btnReject := menu.Data("❌ Отклонить", fmt.Sprintf("app_reject_%d_%d", app.TournamentID, app.TelegramUserID))

	menu.Inline(
		menu.Row(btnApprove, btnReject),
		menu.Row(btnBack),
	)
	return c.Edit(title, menu)
}

func participantMenu(c tb.Context, t *domain.Tournament, tgID domain.TelegramUserID) error {
	menu := &tb.ReplyMarkup{}

	btnInfo := menu.Data("ℹ️ Информация о турнире", fmt.Sprintf("pinfo_tournament%d", t.ID))
	btnMyMatches := menu.Data("📅 Мои матчи", fmt.Sprintf("pmatches_tournament%d_%d", t.ID, tgID))
	btnMain := menu.Data("Главное меню", MainMenu)

	menu.Inline(
		menu.Row(btnInfo),
		menu.Row(btnMyMatches),
		menu.Row(btnMain),
	)
	return c.Edit(fmt.Sprintf("Турнир %s | ID %d", t.Title, t.ID), menu)
}
