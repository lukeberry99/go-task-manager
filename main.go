package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	gap "github.com/muesli/go-app-paths"
)

func setupPath() string {
	// get XDG config path
	scope := gap.NewScope(gap.User, "tasks")
	dirs, err := scope.DataDirs()

	if err != nil {
		log.Fatal(err)
	}

	// Create the app base dir, if it doesn't already exist
	var taskDir string
	if len(dirs) > 0 {
		taskDir = dirs[0]
	} else {
		taskDir, _ = os.UserHomeDir()
	}
	if err := initTaskDir(taskDir); err != nil {
		log.Fatal(err)
	}
	return taskDir
}

// openDB opens a sqlite database
func openDB(path string) (*taskDB, error) {
	db, err := sql.Open("sqlite3", filepath.Join(path, "tasks.db"))
	if err != nil {
		return nil, err
	}

	t := taskDB{db, path}
	if !t.tableExists() {
		err := t.createTable()
		if err != nil {
			return nil, err
		}
	}

	return &t, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
