package tui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/textarea"
)

type model struct {
	// term      string
	// bg        string
	// txtStyle  lipgloss.Style
	// quitStyle lipgloss.Style
	inputArea     textarea.Model
	chatHistory   []string
	latestMessage int
	err           error
	userName      string
	sqlConnection *sql.DB
}
