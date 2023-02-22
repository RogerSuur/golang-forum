package sqlite

import (
	"fmt"
	"forum-advanced-features/pkg/models"
	"log"

	uuid "github.com/satori/go.uuid"
)

// This will process reactions to the database
func (m *DBModel) AddNotification(PostID, reactorID string, reaction int64) {
	fmt.Println("Adding notifications to DB")
	fmt.Println(PostID, reactorID, reaction)
	NotificationID := uuid.NewV4().String()

	var reactionType string
	if reaction != 1.0 && reaction != -1 {
		reactionType = "comment"
	} else {
		reactionType = "like"
	}

	var postAuthorID string
	findPostAuthor := m.DB.QueryRow(`SELECT UserID FROM Posts WHERE Posts.PostID= ?`, PostID)
	err := findPostAuthor.Scan(&postAuthorID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("PostAuthorID:", postAuthorID)

	stmt := `INSERT INTO Notifications VALUES (?,?,?,?,?)`
	statement, err := m.DB.Prepare(stmt)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec(NotificationID, postAuthorID, reactorID, PostID, reactionType)

	fmt.Println("Added a new notification to DB")
}

//Get all notifications to that user
func (m *DBModel) GetUserNotifications(session *models.SessionData) (notifications []*models.NotificationsData, err error) {
	fmt.Println("User likes")
	stmt := `
	SELECT Notifications.NotificationID, Notifications.UserID, Users.UserName, Notifications.PostID,Notifications.Type, Posts.ParentID, Posts.PostTitle
FROM Notifications 
LEFT JOIN Users ON Notifications.ReactorID = Users.UserID 
LEFT JOIN Posts ON Notifications.PostID = Posts.PostID
WHERE Notifications.UserID = ?`

	rows, err := m.DB.Query(stmt, session.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := &models.NotificationsData{}

		err = rows.Scan(&s.NotificationID, &s.UserID, &s.ReactorID, &s.PostID, &s.Type, &s.ParentID, &s.PostTitle)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

// This will delete notifications from the database
func (m *DBModel) DeleteNotification(PostID string) {
	fmt.Println("Deleting notifications to DB")
	//
	fmt.Println("Deleted a notification from DB")
}
