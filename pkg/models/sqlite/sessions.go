package sqlite

import (
	"fmt"
	"groupforum/pkg/models"
	"log"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

const SessionLength int = 60 * 60

func (m *DBModel) StoreSession(sID, UserID, SessionStart string) {
	stmt := `INSERT INTO Sessions (SessionID, UserID, SessionStart)
	VALUES(?,?,?)`
	statement, err := m.DB.Prepare(stmt)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec(sID, UserID, SessionStart)
}

func (m *DBModel) GetUser(w http.ResponseWriter, r *http.Request) *models.SessionData {

	// get cookie
	cookie, err := r.Cookie("forum")
	if err != nil {
		sID := uuid.NewV4().String()
		cookie = &http.Cookie{
			Name:   "forum",
			Value:  sID,
			Path:   "/",
			MaxAge: SessionLength,
		}
		http.SetCookie(w, cookie)
	}

	// if the user exists already, get user
	row := m.DB.QueryRow(`SELECT Sessions.UserID, Users.UserName from Sessions JOIN Users on Sessions.UserID = Users.UserID WHERE SessionID = ? AND SessionActive = 1`, cookie.Value)
	session := &models.SessionData{}
	if err := row.Scan(&session.UserID, &session.UserName); err != nil {
		//fmt.Println(err)
		session.UserID = ""
		return session
	}
	return session
}

func (m *DBModel) AlreadyLoggedIn(UserID string) (bool, []string) {
	rows, err := m.DB.Query(`SELECT SessionID from Sessions WHERE UserID = ? AND SessionActive = 1`, UserID)
	if err != nil {
		fmt.Println("Error with retrieving sessions from DB", err)
	}

	var sessionIDs []string

	defer rows.Close()

	for rows.Next() {
		var sessionID string
		err := rows.Scan(&sessionID)
		if err != nil {
			fmt.Println("Error with storing session IDs from DB", err)
		}
		sessionIDs = append(sessionIDs, sessionID)
	}
	err = rows.Err()
	if err != nil {
		return false, nil
	}

	return true, sessionIDs
}

func (m *DBModel) IsLoggedIn(req *http.Request) bool {
	cookie, err := req.Cookie("forum")
	if err != nil {
		return false
	}

	row := m.DB.QueryRow(`SELECT UserID from Sessions WHERE SessionID = ? AND SessionActive = 1`, cookie.Value)
	session := &models.SessionData{}
	if err := row.Scan(&session.UserID); err != nil {
		fmt.Println("The user has no active session", err)
		return false
	}

	return true
}

func (m *DBModel) Logout(w http.ResponseWriter, req *http.Request) {
	logged := m.IsLoggedIn(req)
	if !logged {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	cookie, _ := req.Cookie("forum")

	m.RemoveSession(cookie.Value)

	// remove the cookie
	cookie = &http.Cookie{
		Name:   "forum",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	log.Println("User logged out")

	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func (m *DBModel) RemoveSession(SessionID string) {
	// delete the session

	statement, err := m.DB.Prepare(`UPDATE Sessions	SET SessionActive = 0 WHERE SessionID = ?`)
	if err != nil {
		fmt.Println("Error with deactivating sessions:", err)
	}
	statement.Exec(SessionID)
}
