package tui

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	return initialModel(s, sqliteDatabase), []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseAllMotion()}
}

func initialModel(s ssh.Session, sqlConnection *sql.DB) model {
	pty, _, _ := s.Pty()
	ti := textarea.New()
	ti.Placeholder = "Type to chat, enter to send"
	ti.ShowLineNumbers = false
	ti.Focus()
	ti.CharLimit = 140
	ti.SetWidth(pty.Window.Width)
	ti.SetHeight(1)
	userName := s.User()

	chatHistoryArea := viewport.New(pty.Window.Width-5, pty.Window.Height-8)
	chatHistoryArea.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBackground(lipgloss.Color("#AAAAAA")).
		BorderForeground(lipgloss.Color("#222222"))

	return model{
		userName:        userName,
		inputArea:       ti,
		chatHistoryArea: chatHistoryArea,
		err:             nil,
		chatHistory:     "",
		latestMessage:   0,
		sqlConnection:   sqlConnection,
	}
}

func handleWindowResize(m *model, msg tea.WindowSizeMsg) {
	m.chatHistoryArea.Width = msg.Width - 5
	m.chatHistoryArea.Height = msg.Height - 8

	m.inputArea.SetWidth(msg.Width - 5)
}

func (m model) Init() tea.Cmd {
	m.readIn()
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmdIA  tea.Cmd
		cmdCHA tea.Cmd
		cmds   []tea.Cmd
	)

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
			m.chatHistoryArea.GotoBottom()
			return m, nil

		case tea.KeyTab:
			if m.inputArea.Height() > 2 {
				return m, nil
			}
			m.inputArea.SetHeight(m.inputArea.Height() + 1)
			m.inputArea.CursorDown()

		case tea.KeyUp:
			fallthrough
		case tea.KeyDown:
			fallthrough
		case tea.KeyPgUp:
			fallthrough
		case tea.KeyPgDown:
			m.chatHistoryArea, cmdCHA = m.chatHistoryArea.Update(msg)
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			fallthrough
		case tea.MouseButtonWheelDown:
			m.chatHistoryArea, cmdCHA = m.chatHistoryArea.Update(msg)
		}

	case tea.WindowSizeMsg:
		handleWindowResize(&m, msg)

	case errMsg:
		m.err = msg
		return m, nil

	}
	m.inputArea, cmdIA = m.inputArea.Update(msg)
	cmds = append(cmds, cmdCHA, cmdIA)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	userName := lipgloss.NewStyle().SetString(m.userName).Foreground(lipgloss.Color("#FF00FF"))
	chatText := fmt.Sprintf("%v\n\n", m.chatHistoryArea.View())
	userInput := fmt.Sprintf("%s%s\n%s\n\n%s",
		userName, " >", m.inputArea.View(), "ctrl+c to quit")
	return chatText + userInput
}
