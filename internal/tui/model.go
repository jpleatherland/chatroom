package tui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

type model struct {
	// term      string
	// bg        string
	// txtStyle  lipgloss.Style
	// quitStyle lipgloss.Style
	inputArea       textarea.Model
	chatHistoryArea viewport.Model
	chatHistory     string
	latestMessage   int
	err             error
	userName        string
	sqlConnection   *sql.DB
}
