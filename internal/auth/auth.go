package auth

import (
	"database/sql"
	"errors"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func GetUser(pubKey ssh.PublicKey, sqlConnection *sql.DB) (bool, error) {
	selectSQL := "SELECT * FROM Users WHERE publickey = ?"
	pubKeyByte := []byte(pubKey.Marshal())
	rows, err := sqlConnection.Query(selectSQL, pubKeyByte)
	if err != nil {
		log.Error(err.Error())
		return false, errors.New("DB Error")
	}
	defer rows.Close()
	for rows.Next() {
		var userid int
		var username string
		var publickey []byte
		var authorised int
		err = rows.Scan(&userid, &username, &publickey, &authorised)
		if err != nil {
			log.Error(err.Error())
			return false, err
		}
		keyMatch := checkKeyEquality(pubKeyByte, publickey)
		if keyMatch && authorised == 1 {
			return true, nil
		}
		if keyMatch && authorised == 0 {
			return false, errors.New("Unauthorised")
		}
	}
	return false, nil
}

func AddUser(userName string, pubKey ssh.PublicKey, sqlConnection *sql.DB) error {
	insertSQL := "INSERT INTO Users(userid, username, publickey, authorised) VALUES(?, ?, ?, ?)"
	statement, err := sqlConnection.Prepare(insertSQL)
	pubKeyByte := []byte(pubKey.Marshal())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = statement.Exec(sql.NullInt32{}, userName, pubKeyByte, 1)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func checkKeyEquality(pubKey []byte, storedKey []byte) bool {
	if len(pubKey) != len(storedKey) {
		return false
	}
	for i := range pubKey {
		if pubKey[i] != storedKey[i] {
			return false
		}
	}
	return true
}
