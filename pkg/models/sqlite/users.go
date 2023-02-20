package sqlite

import (
	"fmt"
	"log"

	"groupforum/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

func (m *DBModel) AddUser(UserID, UserName, Email string, PwdHash []byte, JoinTime string) error {

	row := m.DB.QueryRow(`SELECT * from Users WHERE UserName = ? OR Email = ?`, UserName, Email)
	user := &models.UserData{}
	row.Scan(&user.UserID, &user.UserName, &user.Email, &user.PwdHash, &user.JoinTime)

	if user.UserName == UserName {
		return fmt.Errorf("User with same name already exists")
	} else if user.Email == Email {
		return fmt.Errorf("User with same email already exists")
	}

	stmt := `INSERT INTO Users (UserID, UserName, Email, PwdHash, JoinTime)
	VALUES(?,?,?,?,?)`
	statement, err := m.DB.Prepare(stmt)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec(UserID, UserName, Email, PwdHash, JoinTime)

	fmt.Printf("Inserted user %s with email %s into database table Users\n", UserName, Email)
	return nil
}

func (m *DBModel) Login(UserName, Pwd string) (*models.UserData, error) {
	fmt.Println("Trying to log in", UserName)
	row := m.DB.QueryRow(`SELECT * from Users WHERE UserName = ?`, UserName)
	user := &models.UserData{}

	if err := row.Scan(&user.UserID, &user.UserName, &user.Email, &user.PwdHash, &user.JoinTime); err != nil {
		return nil, fmt.Errorf("User not found")
	}

	if err := bcrypt.CompareHashAndPassword(user.PwdHash, []byte(Pwd)); err != nil {
		return nil, fmt.Errorf("Incorrect password")
	}

	return user, nil
}
