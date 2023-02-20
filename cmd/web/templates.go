package web

import "groupforum/pkg/models"

type templateData struct {
	PostsData    []*models.PostData
	SessionData  *models.SessionData
	UserPageData string
	ThreadData   []string
	IsThread     bool
}
