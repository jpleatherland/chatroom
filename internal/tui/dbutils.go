package tui

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

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
		m.chatHistory = m.chatHistory + fmt.Sprintf("[%s] %s: %s \n", timeStamp, user, message)
		m.chatHistoryArea.SetContent(m.chatHistory)
	}
	m.chatHistoryArea.GotoBottom()
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
		m.chatHistory = m.chatHistory + fmt.Sprintf("[%s] %s: %s \n", timeStamp, user, message)
		m.chatHistoryArea.SetContent(m.chatHistory)
	}
	return nil
}

type (
	errMsg error
)
