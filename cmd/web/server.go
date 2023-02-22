package web

import (
	"database/sql"
	"flag"
	"fmt"
	"forum-advanced-features/pkg/models/sqlite"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	database *sqlite.DBModel
}

func Server() {
	file, err := os.OpenFile("forum_DB.db", os.O_RDWR|os.O_CREATE, 0755) // If not exists, create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("Database file successfully opened")

	dsn := flag.String("dsn", "forum_DB.db", "sql date")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()
	createTable(db)

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		database: &sqlite.DBModel{DB: db},
	}

	srv := &http.Server{
		Addr:     ":8080",
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", srv.Addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	fmt.Println("dsn:", dsn)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func createTable(db *sql.DB) {

	usersTableSQL := `CREATE TABLE IF NOT EXISTS "Users" (
		"UserID"	VARCHAR(36) UNIQUE,
		"UserName"	VARCHAR(200),
		"Email"	VARCHAR(100) NOT NULL,
		"PwdHash"	VARCHAR(100),
		"JoinTime"	TIMESTAMP,
		PRIMARY KEY("UserID")
	);` // sql Statement for create table

	sessionTableSQL := `CREATE TABLE IF NOT EXISTS "Sessions" (
		"SessionID"	VARCHAR(36) UNIQUE,
		"UserID"	VARCHAR(36) NOT NULL,
		"SessionStart"	TIMESTAMP,
		"SessionActive"	INTEGER DEFAULT 1,
		PRIMARY KEY("SessionID"),
		FOREIGN KEY("UserID") REFERENCES "Users"("UserID")
	);` // sql Statement for create table

	postsTableSQL := `CREATE TABLE IF NOT EXISTS "Posts" (
		"PostID"	VARCHAR(36) UNIQUE,
		"ParentID"	VARCHAR(36),
		"UserID"	VARCHAR(36) NOT NULL,
		"PostTitle"	TEXT,
		"PostContent"	TEXT,
		"PostImage"	TEXT,
		"PostTime"	TIMESTAMP,
		PRIMARY KEY("PostID"),
		FOREIGN KEY("UserID") REFERENCES "Users"("UserID")
	);` // sql Statement for create table

	postCatTableSQL := `CREATE TABLE IF NOT EXISTS "PostCatRelations" (
		"PostID"	VARCHAR(36),
		"Category"	VARCHAR(36),
		FOREIGN KEY("PostID") REFERENCES "Posts"("PostID")
	);` // sql Statement for create table

	likesTableSQL := `CREATE TABLE IF NOT EXISTS "Likes" (
		"UserID"	VARCHAR(36),
		"PostID"	VARCHAR(36),
		"LikeValue"	INTEGER,
		FOREIGN KEY("UserID") REFERENCES "Users"("UserID"),
		FOREIGN KEY("PostID") REFERENCES "Posts"("PostID")		
	);` // sql Statement for create table

	notificationsTableSQL := `CREATE TABLE IF NOT EXISTS "Notifications" (
		"NotificationID" VARCHAR(36),
		"UserID"	VARCHAR(36),
		"ReactorID"	VARCHAR(36),
		"PostID"	VARCHAR(36),
		"Type"	VARCHAR(36),
		FOREIGN KEY("UserID") REFERENCES "Users"("UserID"),
		FOREIGN KEY("ReactorID") REFERENCES "Users"("UserID"),
		FOREIGN KEY("PostID") REFERENCES "Posts"("PostID")		
	);` // sql Statement for create table

	createTablesSQL := map[string]string{
		"Users":           usersTableSQL,
		"Sessions":        sessionTableSQL,
		"Posts":           postsTableSQL,
		"Post Categories": postCatTableSQL,
		"Post Likes":      likesTableSQL,
		"Notifications":   notificationsTableSQL,
	}

	for key, value := range createTablesSQL {
		statement, err := db.Prepare(value) // Prepare SQL Statement
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("%v Table initialized", key)
		statement.Exec() // Execute SQL Statement
	}

}
