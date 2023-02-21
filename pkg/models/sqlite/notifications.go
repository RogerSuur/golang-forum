package sqlite

import (
	"fmt"
	"log"
)

// This will process reactions to the database
func (m *DBModel) AddNotification(PostID, reactorID string, reaction int64) {
	fmt.Println("Adding notifications to DB")
	fmt.Println(PostID, reactorID, reaction)

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

	stmt := `INSERT INTO Notifications VALUES (?,?,?,?)`
	statement, err := m.DB.Prepare(stmt)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec(postAuthorID, reactorID, PostID, reactionType)

	fmt.Println("Added a new notification to DB")
}

// This will delete notifications from the database
func (m *DBModel) DeleteNotification(PostID, reactorID string, reaction int64) {
	fmt.Println("Deleting notifications to DB")
	//
	fmt.Println("Deleted a notification from DB")
}
