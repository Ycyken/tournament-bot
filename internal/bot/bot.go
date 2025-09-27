// `internal/bot/bot.go`
package bot

import (
	"log"
	"sync"
	"time"

	"github.com/Ycyken/tournament-bot/internal/domain"
	"github.com/Ycyken/tournament-bot/internal/service"
	tb "gopkg.in/telebot.v3"
)

const (
	StateMainMenu               = "main_menu"
	StateWaitingTournamentTitle = "waiting_tournament_title"
	StateTournamentManagement
	StateReviewApplications = "tournament_adding_participants"
)

type Bot struct {
	bot    *tb.Bot
	svc    *service.Service
	states map[domain.TelegramUserID]string
	mu     sync.RWMutex
}

func NewBot(svc *service.Service, token string) (*Bot, error) {
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	bt := &Bot{
		bot:    bot,
		svc:    svc,
		states: make(map[domain.TelegramUserID]string),
	}

	registerHandlers(bt)

	return bt, nil
}

func (b *Bot) Run() {
	b.bot.Start()
	return
}

func (b *Bot) setState(userID domain.TelegramUserID, state string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.states[userID] = state
}

func (b *Bot) getState(userID domain.TelegramUserID) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.states[userID]
}

func (b *Bot) clearState(userID domain.TelegramUserID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.states, userID)
}
