package tui

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/charmbracelet/ssh"
	"github.com/jpleatherland/chatroom/internal/config"
)

func TeaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Error("Could not read config file", "error", err)
		os.Exit(1)
	}
	file := cfg.DatabaseFile
	sqliteDatabase, _ := sql.Open("sqlite3", file)
	return initialModel(s, sqliteDatabase), []tea.ProgramOption{tea.WithAltScreen()}
}

func initialModel(s ssh.Session, sqlConnection *sql.DB) model {
	pty, _, _ := s.Pty()
	ti := textarea.New()
	ti.Placeholder = "Type to chat, enter to send"
	ti.ShowLineNumbers = false
	ti.Focus()
	ti.CharLimit = 140
	ti.SetWidth(pty.Window.Width - 20)
	ti.SetHeight(1)
	userName := s.User()

	return model{
		userName:      userName,
		inputArea:     ti,
		err:           nil,
		chatHistory:   []string{},
		latestMessage: 0,
		sqlConnection: sqlConnection,
	}
}

func (m model) Init() tea.Cmd {
	m.readIn()
	return textinput.Blink
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.readDelta()
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.sqlConnection.Close()
			return m, tea.Quit

		case tea.KeyEnter:
			err := m.writeOut(m.inputArea.Value())
			if err != nil {
				m.inputArea.SetValue(err.Error())
				return m, nil
			}
			m.inputArea.Reset()
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	}
	m.inputArea, cmd = m.inputArea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	chatText := fmt.Sprintf("chathistory:\n%v\n\n", strings.Join(m.chatHistory, "\n"))
	userInput := fmt.Sprintf("%s%s\n%s\n\n%s", m.userName, "> ", m.inputArea.View(), "ctrl+c to quit")
	return chatText + userInput
}
