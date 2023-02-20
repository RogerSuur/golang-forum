package sqlite

import (
	"fmt"
	"log"
)

// This will process reactions in the database
func (m *DBModel) ProcessReaction(PostID, UserID string, reaction int64) {
	var previousState int32
	var stmt string
	checkLike := m.DB.QueryRow(`SELECT LikeValue FROM Likes WHERE PostID = ? AND UserID = ?`, PostID, UserID)
	err := checkLike.Scan(&previousState)
	if err != nil {
		fmt.Println(err)
	}

	switch previousState {
	case 1:
		if reaction == -1 {
			stmt = `UPDATE Likes SET LikeValue = ? WHERE PostID = ? AND UserID = ?`
		} else {
			stmt = `DELETE from Likes WHERE LikeValue = ? AND PostID = ? AND UserID = ?`
		}
	case -1:
		if reaction == 1 {
			stmt = `UPDATE Likes SET LikeValue = ? WHERE PostID = ? AND UserID = ?`
		} else {
			stmt = `DELETE from Likes WHERE LikeValue = ? AND PostID = ? AND UserID = ?`
		}
	default:
		stmt = `INSERT INTO Likes (LikeValue, PostID, UserID)
				VALUES(?,?,?)`
	}

	statement, err := m.DB.Prepare(stmt)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec(reaction, PostID, UserID)
	fmt.Println("Reaction processed in database table Likes")
}
