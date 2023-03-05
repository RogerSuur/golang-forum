package web

import "forum-advanced-features/pkg/models"

type templateData struct {
	PostsData         []*models.PostData
	SessionData       *models.SessionData
	UserPageData      string
	ThreadData        []string
	IsThread          bool
	NotificationData  []*models.NotificationsData
	NotificationCount string
}
