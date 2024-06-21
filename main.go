package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	_ "github.com/mattn/go-sqlite3"
)

const (
	host = "localhost"
	port = "42069"
)

// var db *sqlite3.SQLiteDriver

var file string = os.Args[1]

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()
	<-done
	log.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() { cancel() }()
	if err = s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not shutdown server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	sqliteDatabase, _ := sql.Open("sqlite3", file)
	return initialModel(s, sqliteDatabase), []tea.ProgramOption{tea.WithAltScreen()}
}

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

func (m model) writeOut(message string) error {
	insertSQL := "INSERT INTO ChatHistory(rowId, User, TimeStamp, message) VALUES(?, ?, ?, ?)"
	statement, err := m.sqlConnection.Prepare(insertSQL)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = statement.Exec(sql.NullInt32{}, m.userName, time.Now().Format(time.RFC3339), message)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (m *model) readIn() error {
	selectSQL := "SELECT * FROM ChatHistory"
	rows, err := m.sqlConnection.Query(selectSQL)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var rowId int
		var user string
		var timeStamp string
		var message string
		err = rows.Scan(&rowId, &user, &timeStamp, &message)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		m.latestMessage = int(rowId)
		m.chatHistory = append(m.chatHistory, fmt.Sprintf("%s: %s", user, message))
	}
	return nil
}

func (m *model) readDelta() error {
	selectSQL := "SELECT * FROM ChatHistory WHERE rowId > ?"
	rows, err := m.sqlConnection.Query(selectSQL, m.latestMessage)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var rowId int
		var user string
		var timeStamp string
		var message string
		err = rows.Scan(&rowId, &user, &timeStamp, &message)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		m.latestMessage = int(rowId)
		m.chatHistory = append(m.chatHistory, fmt.Sprintf("[%s] %s: %s", timeStamp, user, message))
	}
	return nil
}

type (
	errMsg error
)

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
