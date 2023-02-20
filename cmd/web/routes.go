package web

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/newpost", app.newpost)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/logout", app.database.Logout)
	mux.HandleFunc("/userpage", app.userpage)
	mux.HandleFunc("/react", app.react)
	mux.HandleFunc("/thread", app.thread)
	// Create a file server which serves files out of the "./ui/static" directo
	// Note that the path given to the http.Dir function is relative to the pro
	// directory root.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Use the mux.Handle() function to register the file server as the handler
	// all URL paths that start with "/static/". For matching paths, we strip t
	// "/static" prefix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	return mux
}
