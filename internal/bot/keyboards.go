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

func tournamentsPage(bt *Bot, c tb.Context, tournaments map[domain.TournamentID]string, page int) error {
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
	menu.Inline(rows...)

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
	return c.Send("Ваши турниры:", menu)
}
