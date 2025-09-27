package domain

type Application struct {
	TournamentID   TournamentID
	TelegramUserID TelegramUserID
	Name           string
	TelegramTag    *string
	Text           *string
}
