package sqlite

import (
	"database/sql"
	"fmt"
	"log"

	"forum-advanced-features/pkg/models"
)

// This will insert a new post into the database.
func (m *DBModel) Insert(PostID, ParentID, UserID, PostTitle, PostContent, PostImage, PostTime string, TagsSelected []string) {
	fmt.Println("Insert post")
	if len(PostTitle) > 8 {
		if PostTitle[0:8] == "Re: Re: " {
			PostTitle = PostTitle[4:]
		}
	}

	stmt := `INSERT INTO Posts (PostID, ParentID, UserID, PostTitle, PostContent, PostImage, PostTime)
	VALUES(?,?,?,?,?,?,?)`
	statement, err := m.DB.Prepare(stmt)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec(PostID, ParentID, UserID, PostTitle, PostContent, PostImage, PostTime)
	fmt.Printf("Inserted post '%v' into database table Posts\n", PostTitle)

	for _, cat := range TagsSelected {
		insertCategorySQL := `INSERT INTO PostCatRelations(PostID, Category) VALUES (?, ?)`
		m.DB.Exec(insertCategorySQL, PostID, cat)
		fmt.Printf("Inserted Post as %v into database table PostCatRelations\n", cat)
	}
}

// This will return created posts.
func (m *DBModel) Latest(session *models.SessionData, TagsSelected []string) ([]*models.PostData, error) {
	fmt.Println("Latest posts")
	var recentPosts []*models.PostData
	// create query statement
	categoriesSQL := ``
	if len(TagsSelected) != 0 && TagsSelected[0] != "All" {
		categoriesSQL = `LEFT JOIN PostCatRelations USING(PostID)
		WHERE`
		for _, cat := range TagsSelected {
			categoriesSQL += ` PostCatRelations.Category = "` + cat + `" OR`
		}
		categoriesSQL = categoriesSQL[:len(categoriesSQL)-3] + ` AND`
	} else {
		categoriesSQL = `WHERE`
	}

	stmt := `SELECT DISTINCT
	Posts.*,
		Users.UserName,
			(SELECT Likes.LikeValue
			FROM Likes
			WHERE Likes.UserID = ? AND Posts.PostID = Likes.PostID)
		LikeValue,
			(Select SUM(Likes.LikeValue)
			From Likes
			WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue > 0)
		Positive,
			(Select SUM(Likes.LikeValue)
			From Likes
			WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue < 0)
		Negative
	FROM (Select *, COUNT(Posts.ParentID)-1 Parents
			From Posts
			GROUP BY Posts.ParentID) Posts
	LEFT JOIN Users USING(UserID)
		` + categoriesSQL + ` Posts.ParentID = Posts.PostID
		ORDER BY Posts.PostTime DESC`

	// process the query
	rows, err := m.DB.Query(stmt, session.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err == sql.ErrNoRows {
		fmt.Println("models: no matching record found")
	}

	for rows.Next() {
		// Create a pointer to a new zeroed PostData struct.
		s := &models.PostData{}

		err := rows.Scan(&s.PostID, &s.ParentID, &s.UserID, &s.PostTitle, &s.PostContent, &s.PostImage, &s.PostTime, &s.Parents, &s.UserName, &s.PostLiked, &s.Positive, &s.Negative)
		if err != nil {
			fmt.Println("ERR: ", err)
		}
		// Append it to the slice of Posts.
		recentPosts = append(recentPosts, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK then return the PostsData slice.
	return recentPosts, nil
}

func (m *DBModel) UserPosts(session *models.SessionData) (userPosts []*models.PostData, err error) {
	fmt.Println("UserPosts")
	stmt := `SELECT DISTINCT
	Posts.*, 
	Users.UserName,
		(SELECT Likes.LikeValue
		FROM Likes
		WHERE Likes.UserID = ? AND Posts.PostID = Likes.PostID)
	LikeValue,
		(Select SUM(Likes.LikeValue)
		From Likes
		WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue > 0)
	Positive,
		(Select SUM(Likes.LikeValue)
		From Likes
		WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue < 0)
	Negative,
		(COUNT(Posts.PostID) OVER (PARTITION BY Posts.ParentID))-1 AS Parents
	FROM Posts
	LEFT JOIN Users USING(UserID)
	WHERE Posts.UserID = ?
	GROUP BY Posts.PostID
	ORDER BY Posts.PostTime DESC`

	rows, err := m.DB.Query(stmt, session.UserID, session.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := &models.PostData{}

		err = rows.Scan(&s.PostID, &s.ParentID, &s.UserID, &s.PostTitle, &s.PostContent, &s.PostImage, &s.PostTime, &s.UserName, &s.PostLiked, &s.Positive, &s.Negative, &s.Parents)
		if err != nil {
			return nil, err
		}

		userPosts = append(userPosts, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userPosts, nil
}

func (m *DBModel) UserLikes(session *models.SessionData) (userPosts []*models.PostData, err error) {
	fmt.Println("User likes")
	stmt := `SELECT
	Posts.*, 
	Users.UserName,
		(SELECT Likes.LikeValue
		FROM Likes
		WHERE Likes.UserID = ? AND Posts.PostID = Likes.PostID)
	LikeValue,
		(Select SUM(Likes.LikeValue)
		From Likes
		WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue > 0)
	Positive,
		(Select SUM(Likes.LikeValue)
		From Likes
		WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue < 0)
	Negative,
		(COUNT(Posts.PostID) OVER (PARTITION BY Posts.ParentID))-1 AS Parents
	FROM Posts
	LEFT JOIN Users USING(UserID)
	WHERE Posts.PostID IN (
		SELECT Likes.PostID
		FROM Likes
		WHERE Likes.UserID = ?)
	GROUP BY Posts.PostID
	ORDER BY Posts.PostTime DESC`

	rows, err := m.DB.Query(stmt, session.UserID, session.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := &models.PostData{}

		err = rows.Scan(&s.PostID, &s.ParentID, &s.UserID, &s.PostTitle, &s.PostContent, &s.PostImage, &s.PostTime, &s.UserName, &s.PostLiked, &s.Positive, &s.Negative, &s.Parents)
		if err != nil {
			return nil, err
		}

		userPosts = append(userPosts, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userPosts, nil
}

func (m *DBModel) GetThread(session *models.SessionData, thread string) (threadPosts []*models.PostData, err error) {
	fmt.Println("Get thread")
	stmt := `SELECT DISTINCT
		Posts.*, 
		Users.UserName,
			(SELECT Likes.LikeValue
			FROM Likes
			WHERE Likes.UserID = ? AND Posts.PostID = Likes.PostID)
		LikeValue,
			(Select SUM(Likes.LikeValue)
			From Likes
			WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue > 0)
		Positive,
			(Select SUM(Likes.LikeValue)
			From Likes
			WHERE Posts.PostID = Likes.PostID AND Likes.LikeValue < 0)
		Negative,
			(Select COUNT(Posts.PostID)
			From Posts
			WHERE Posts.ParentID = ?)-1
		Parents
	FROM Posts
	LEFT JOIN Sessions USING(UserID)
	LEFT JOIN Users USING(UserID) 
	WHERE Posts.PostID = ? OR Posts.ParentID = ?
	ORDER BY Posts.PostTime`

	rows, err := m.DB.Query(stmt, session.UserID, thread, thread, thread)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := &models.PostData{}

		err = rows.Scan(&s.PostID, &s.ParentID, &s.UserID, &s.PostTitle, &s.PostContent, &s.PostImage, &s.PostTime, &s.UserName, &s.PostLiked, &s.Positive, &s.Negative, &s.Parents)
		if err != nil {
			return nil, err
		}

		threadPosts = append(threadPosts, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threadPosts, nil
}

func (m *DBModel) IsComment(PostParent string) bool {

	var postAuthorID string
	stmt := m.DB.QueryRow(`
	 SELECT UserID FROM Posts WHERE Posts.ParentID = ? AND Posts.PostID = ?`, PostParent, PostParent)
	err := stmt.Scan(&postAuthorID)
	if err != nil {
		fmt.Println("err")
		fmt.Println(err)
	}

	fmt.Println("PostAuthorID:", postAuthorID)

	fmt.Println("IsComment = true")

	return true
}
