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
	StateApplyEnterName = "application_enter_name"
	StateApplyEnterText = "application_enter_text"
)

type applyCtx struct {
	TournamentID domain.TournamentID
	Name         string
}

func (b *Bot) setApply(uid domain.TelegramUserID, a *applyCtx) {
	if b.apply == nil {
		b.apply = make(map[domain.TelegramUserID]*applyCtx)
	}
	b.apply[uid] = a
}
func (b *Bot) getApply(uid domain.TelegramUserID) *applyCtx {
	if b.apply == nil {
		return nil
	}
	return b.apply[uid]
}
func (b *Bot) clearApply(uid domain.TelegramUserID) {
	if b.apply != nil {
		delete(b.apply, uid)
	}
}

type Bot struct {
	bot    *tb.Bot
	svc    *service.Service
	states map[domain.TelegramUserID]string
	apply  map[domain.TelegramUserID]*applyCtx
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
