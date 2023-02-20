package main

import (
	web "groupforum/cmd/web"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	web.Server()
}
