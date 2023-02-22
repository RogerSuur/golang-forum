package models

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models:no matching record found")

type PostData struct {
	PostID      string
	ParentID    string
	UserID      string
	UserName    string
	PostTitle   string
	PostContent string
	PostImage   string
	PostLiked   sql.NullInt32
	Positive    sql.NullInt32
	Negative    sql.NullInt32
	PostTime    time.Time
	Parents     sql.NullInt32
}

type NotificationsData struct {
	UserID    string
	ReactorID string
	PostID    string
	Type      string
	PostTitle string
	ParentID  string
	Parents   sql.NullInt32
}

type UserData struct {
	UserID   string
	UserName string
	Email    string
	PwdHash  []byte
	JoinTime string
}

type SessionData struct {
	SessionID     string
	UserID        string
	UserName      string
	SessionStart  string
	SessionActive int
}
