package main

import (
	_ "github.com/mattn/go-sqlite3"

	"forum-advanced-features/cmd/web"
)

func main() {
	web.Server()
}
