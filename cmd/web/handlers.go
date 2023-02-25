package web

import (
	"fmt"
	"forum-advanced-features/pkg/models"
	"forum-advanced-features/pkg/models/sqlite"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
	bcrypt "golang.org/x/crypto/bcrypt"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("home handler")
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	var TagsSelected []string
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error while handling new post from website /", r.PostForm)
		}
		TagsSelected = r.Form["formtags"]
	}

	session := app.database.GetUser(w, r)

	posts, err := app.database.Latest(session, TagsSelected)
	if err != nil {
		app.serveError(w, err)
		return
	}

	var threadPostID []string

	for _, element := range posts {
		threadPostID = append(threadPostID, "/thread?ID="+element.PostID)
	}

	data := &templateData{
		PostsData:   posts,
		SessionData: session,
		ThreadData:  threadPostID,
		IsThread:    false,
	}

	files := []string{
		"./ui/html/home.html",
		"./ui/html/posts.html",
		"./ui/html/base.layout.html",
	}
	// The template.ParseFiles() function reads the template file into a
	// template set. If there's an error, log the detailed error message and
	// the http.Error() function to send a generic 500 Internal Server Error
	// response to the user.
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	// The Execute() method on the template set writes the template
	// content as the response body.
	err = ts.Execute(w, data)
	if err != nil {
		app.serveError(w, err)
	}
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/register.html", // path relative to the root of the project cateory
		"./ui/html/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		err = ts.Execute(w, nil)
		if err != nil {
			app.serveError(w, err)
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error while handling new user registration", r.PostForm)
		}
		UserID := uuid.NewV4().String()
		UserName := r.FormValue("registerName")
		Email := r.FormValue("registerEmail")
		PasswdString := r.FormValue("registerPassword")
		creationTime := time.Now().Format("2006-01-02 15:04:05")
		PwdHash, err := bcrypt.GenerateFromPassword([]byte(PasswdString), 10)
		if err != nil {
			log.Println(err)
		}

		if err := app.database.AddUser(UserID, UserName, Email, PwdHash, creationTime); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// create session
		sID := uuid.NewV4().String()
		cookie := &http.Cookie{
			Name:   "forum",
			Value:  sID,
			Path:   "/",
			MaxAge: sqlite.SessionLength,
		}

		http.SetCookie(w, cookie)

		app.database.StoreSession(sID, UserID, creationTime)

		// Redirect the user to the relevant page

		http.Redirect(w, r, "./", http.StatusFound)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/login.html",
		"./ui/html/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		err = ts.Execute(w, nil)

		if err != nil {
			app.serveError(w, err)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error while handling login", r.PostForm)
		}

		loginName := r.FormValue("loginName")
		loginPasswd := r.FormValue("loginPassword")
		keepLogged := r.FormValue("keepLogged")
		keepSession := 0

		switch keepLogged {
		case "on":
			keepSession = sqlite.SessionLength * 24
		default:
			keepSession = sqlite.SessionLength
		}

		if err != nil {
			log.Println(err)
		}

		// check if user exists in database and log in
		user, err := app.database.Login(loginName, loginPasswd)
		if err != nil {
			log.Println(err)
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}

		// check if user has active session and if yes, cancel that
		isLogged, userSessions := app.database.AlreadyLoggedIn(user.UserID)
		if isLogged {
			fmt.Println("User has earlier active sessions, deactivating earlier sessions")
			for _, id := range userSessions {
				app.database.RemoveSession(id)
			}
		} else {
			fmt.Println("No other active sessions detected for user")
		}

		// create session
		sID := uuid.NewV4().String()
		cookie := &http.Cookie{
			Name:   "forum",
			Value:  sID,
			Path:   "/",
			MaxAge: keepSession,
		}
		creationTime := time.Now().Format("2006-01-02 15:04:05")

		http.SetCookie(w, cookie)

		app.database.StoreSession(sID, user.UserID, creationTime)

		fmt.Printf("User %v logged in\n", user.UserName)

		// Redirect the user to the relevant page

		http.Redirect(w, r, "./", http.StatusFound)
	}
}

func (app *application) newpost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("newpost handler")

	if !app.database.IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	session := app.database.GetUser(w, r)

	data := &templateData{
		SessionData: session,
	}

	files := []string{
		"./ui/html/newpost.html",
		"./ui/html/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		err = ts.Execute(w, data)

		if err != nil {
			app.serveError(w, err)
		}
	case http.MethodPost:
		// limit the POST body size to 20.5Mb and throw error if client attempt to send more than that
		r.Body = http.MaxBytesReader(w, r.Body, 20<<20+512)

		PostID := uuid.NewV4().String()
		fmt.Println("creating new PostID ", PostID, "with len", len(PostID))
		UserID := session.UserID
		PostTitle := r.FormValue("inputPostTitle")
		PostContent := r.FormValue("inputPostContent")
		creationTime := time.Now().Format("2006-01-02 15:04:05")
		TagsSelected := r.Form["tagsClearThreshold[]"]
		parentPost := r.FormValue("ParentID")
		PostImage := ""
		redirectPath := "./"

		// FormFile returns the first file for the given key `inputImage` it also returns
		// the FileHeader so we can get the Filename, the Header and the size of the file
		postFile, fileHandler, err := r.FormFile("inputImage")

		// 20 << 20 specifies a maximum upload of 20 MB files
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			fmt.Println("Error: user tried to upload a file over 20mb")
			http.Error(w, "files over 20mb in size are not allowed", http.StatusRequestEntityTooLarge)
			return
		}

		var IsComment = true
		fmt.Println("parentPost:", parentPost)
		if parentPost == "0" {
			parentPost = PostID
			IsComment = false
		} else {
			redirectPath = "./thread?ID=" + parentPost
		}

		if err == nil {

			defer postFile.Close()

			contentType := fileHandler.Header.Get("Content-Type")
			if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
				fmt.Println("Error: user tried to upload a wrong file type:", contentType)
				http.Error(w, "file type not allowed", http.StatusUnsupportedMediaType)
				return
			}

			fmt.Printf("User Attached a File '%+v' with File Size %+v and Content Type %+v\n", fileHandler.Filename, fileHandler.Size, contentType)

			imgPath := filepath.Join(".", "ui/static/forum-images")
			err = os.MkdirAll(imgPath, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}

			// Create a temporary file within our temp-images directory that follows the naming pattern
			tempFile, err := ioutil.TempFile(imgPath, "upload-*-"+fileHandler.Filename)
			if err != nil {
				fmt.Println(err)
			}
			defer tempFile.Close()

			// read all of the contents of our uploaded file into a byte array
			fileBytes, err := ioutil.ReadAll(postFile)
			if err != nil {
				fmt.Println(err)
			}
			// write this byte array to our temporary file
			tempFile.Write(fileBytes)
			PostImage = tempFile.Name()[2:]

		} /*else if err != http.ErrMissingFile {
			// Error handling in case the error is not a missing file
			fmt.Println("Error Retrieving the File:", err)
			http.Error(w, err.Error(), 406)
			return
		}*/

		// Redirect the user to the relevant page for the post.
		app.database.Insert(PostID, parentPost, UserID, PostTitle, PostContent, PostImage, creationTime, TagsSelected)

		//TODO
		//WRITE A QUERY TO DETERMINE IF IT IS A COMMENT OR A POST
		// if app.database.IsComment(parentPost) {
		// 	app.sendNotification(parentPost)
		// }
		if IsComment {
			app.sendNotification(parentPost, UserID)
		}
		//TO DO
		//get the original post UserID from a query of the posts ParentID
		//app.database.AddNotification(PostID, UserID, 0)

		http.Redirect(w, r, redirectPath, http.StatusFound)
	}
}

func (app *application) sendNotification(parentPost string, commentAuthor string) {

	//See userid mille leian on vale? aga miks...
	// UserID := app.database.FindPostAuthor(parentPost)

	fmt.Println("commentAuthor:", commentAuthor)

	app.database.AddNotification(parentPost, commentAuthor, 0)
}

func (app *application) userpage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("userpage handler")
	if !app.database.IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	files := []string{
		"./ui/html/userpage.html", // path relative to the root of the project cateory
		"./ui/html/posts.html",
		"./ui/html/base.layout.html",
		"./ui/html/notifications.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	session := app.database.GetUser(w, r)
	data := &templateData{}
	var posts []*models.PostData
	var notifications []*models.NotificationsData

	switch r.Method {
	case http.MethodGet:
		data = &templateData{
			SessionData: session,
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error while handling user page filters", r.PostForm)
		}
		userPosts := r.FormValue("userpage")
		switch userPosts {
		case "Posts":
			posts, err = app.database.UserPosts(session)
		case "Reactions":
			posts, err = app.database.UserLikes(session)
		case "Notifications":
			notifications, err = app.database.GetUserNotifications(session)

		}

		fmt.Println("handlers userPosts,", userPosts)
		var threadPostID []string

		for _, element := range posts {
			if element.ParentID != "0" {
				threadPostID = append(threadPostID, "/thread?ID="+element.ParentID)
			} else {
				threadPostID = append(threadPostID, "/thread?ID="+element.PostID)
			}

		}

		if err != nil {
			app.serveError(w, err)
			return
		}
		data = &templateData{
			PostsData:        posts,
			SessionData:      session,
			UserPageData:     userPosts,
			ThreadData:       threadPostID,
			NotificationData: notifications,
			IsThread:         true,
		}
	}

	err = ts.Execute(w, data)
	if err != nil {
		app.serveError(w, err)
	}
}

func (app *application) react(w http.ResponseWriter, r *http.Request) {
	fmt.Println("react handler")
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error while handling reactions", r.PostForm)
		}
		reactionPostID := r.FormValue("ID")
		reactionLike, err := strconv.ParseInt(r.FormValue("LikeStatus"), 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		returnPage := r.FormValue("page")
		session := app.database.GetUser(w, r)

		fmt.Println("session.UserId", session.UserID)

		app.database.ProcessReaction(reactionPostID, session.UserID, reactionLike)
		app.database.AddNotification(reactionPostID, session.UserID, reactionLike)

		http.Redirect(w, r, returnPage+"#"+reactionPostID, http.StatusSeeOther)
		return
	}
}

func (app *application) thread(w http.ResponseWriter, r *http.Request) {
	fmt.Println("thread handler")
	files := []string{
		"./ui/html/thread.html", // path relative to the root of the project cateory
		"./ui/html/posts.html",
		"./ui/html/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	if r.Method == "GET" {

		PostID := r.URL.Query().Get("ID")
		session := app.database.GetUser(w, r)

		posts, err := app.database.GetThread(session, PostID)
		if err != nil {
			app.serveError(w, err)
			return
		}

		var threadPostID []string

		for _, element := range posts {
			threadPostID = append(threadPostID, "/thread?ID="+element.ParentID)
		}

		data := &templateData{
			PostsData:   posts,
			SessionData: session,
			ThreadData:  threadPostID,
			IsThread:    true,
		}

		err = ts.Execute(w, data)
		if err != nil {
			app.serveError(w, err)
		}
	}
}

func (app *application) deleteContent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete content handler")
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error while handling reactions", r.PostForm)
		}

		PostID := r.FormValue("PostID")
		fmt.Println(PostID)
		fmt.Println("len PostID", len(PostID))
		if PostID == "" {
			session := app.database.GetUser(w, r)

			fmt.Println("session.UserId", session.UserID)

			app.database.DeleteNotification(session.UserID)

			http.Redirect(w, r, "/userpage", http.StatusSeeOther)
			return
		} else {
			app.database.DeletePost(PostID)
			http.Redirect(w, r, "/userpage", http.StatusSeeOther)
			return
		}
	}
}

func (app *application) editContent(w http.ResponseWriter, r *http.Request) {

	if !app.database.IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	session := app.database.GetUser(w, r)

	files := []string{
		"./ui/html/editpost.html",
		"./ui/html/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serveError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		r.ParseForm()
		PostID := r.FormValue("PostID")
		var posts []*models.PostData
		posts, err = app.database.FindPostWithID(PostID)
		if err != nil {
			app.serveError(w, err)
			return
		}

		data := &templateData{
			SessionData: session,
			PostsData:   posts,
		}

		err = ts.Execute(w, data)

		if err != nil {
			app.serveError(w, err)
		}

	case http.MethodPost:

		PostID := r.FormValue("PostID")

		PostTitle := r.FormValue("inputPostTitle")
		PostContent := r.FormValue("inputPostContent")
		//creationTime := time.Now().Format("2006-01-02 15:04:05")
		parentPost := r.FormValue("ParentID")
		redirectPath := "./"

		//var IsComment = true

		if parentPost == "0" {
			parentPost = PostID
			//	IsComment = false
		} else {
			redirectPath = "./thread?ID=" + parentPost
		}

		// Redirect the user to the relevant page for the post.
		app.database.UpdatePost(PostID, PostTitle, PostContent)

		http.Redirect(w, r, redirectPath, http.StatusFound)

	}
}
