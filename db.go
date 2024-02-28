package main

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

type status int

const (
	todo status = iota
	inProgress
	done
)

func (s status) String() string {
	return [...]string{"todo", "in progress", "done"}[s]
}

type task struct {
	ID      uint
	Name    string
	Project string
	Status  string
	Created time.Time
}

func (t task) FilterValue() string {
	return t.Name
}

func (t task) Description() string {
	return t.Project
}

func (s status) Next() int {
	if s == done {
		return int(todo)
	}
	return int(s + 1)
}

func (s status) Prev() int {
	if s == todo {
		return int(done)
	}
	return int(s - 1)
}

func (s status) Int() int {
	return int(s)
}

type taskDB struct {
	db      *sql.DB
	dataDir string
}

func initTaskDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0o770)
		}
		return err
	}
	return nil
}

func (t *taskDB) tableExists() bool {
	if _, err := t.db.Query("SELECT * FROM tasks"); err == nil {
		return true
	}
	return false
}

func (t *taskDB) createTable() error {
	_, err := t.db.Exec(`CREATE table "tasks" ( "id" INTEGER, "name" TEXT NOT NULL, "project" TEXT, "status" TEXT, "created" DATETIME, PRIMARY KEY("id" AUTOINCREMENT))`)
	return err
}

func (t *taskDB) insert(name, project string) error {
	_, err := t.db.Exec(
		"INSERT INTO tasks(name, project, status, created) VALUES( ?, ?, ?, ?)",
		name,
		project,
		todo.String(),
		time.Now())

	return err
}

func (t *taskDB) delete(id uint) error {
	_, err := t.db.Exec("DELETE FROM tasks WHERE id = ?", id)

	return err
}

func (t *taskDB) update(task task) error {
	orig, err := t.getTask(task.ID)
	if err != nil {
		return err
	}
	orig.merge(task)
	_, err = t.db.Exec(
		"UPDATE tasks SET name = ?, project = ?, status = ? WHERE id = ?",
		orig.Name,
		orig.Project,
		orig.Status,
		orig.ID)
	return err
}

func (orig *task) merge(t task) {
	uValues := reflect.ValueOf(&t).Elem()
	oValues := reflect.ValueOf(orig).Elem()
	for i := 0; i < uValues.NumField(); i++ {
		uField := uValues.Field(i).Interface()
		if oValues.CanSet() {
			if v, ok := uField.(int64); ok && uField != 0 {
				oValues.Field(i).SetInt(v)
			}
			if v, ok := uField.(string); ok && uField != "" {
				oValues.Field(i).SetString(v)
			}
		}
	}
}

type Options struct {
	Status  uint
	Project string
}

func (t *taskDB) getTasks(opts Options) ([]task, error) {
	var tasks []task
	var args []interface{}

	queryString := "SELECT * FROM tasks"
	where := make([]string, 0)

	if opts.Status != 0 {
		where = append(where, "status = ?")
		args = append(args, opts.Status)
	}

	if opts.Project != "" {
		where = append(where, "project = ?")
		args = append(args, opts.Project)
	}

	if len(where) > 0 {
		queryString += " WHERE " + strings.Join(where, " AND ")
	}

	fmt.Printf(queryString, args...)

	rows, err := t.db.Query(queryString, args...)
	if err != nil {
		return tasks, fmt.Errorf("unable to get values: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var task task
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Project,
			&task.Status,
			&task.Created,
		)

		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t *taskDB) getTask(id uint) (task, error) {
	var task task
	err := t.db.QueryRow("SELECT * FROM tasks WHERE id = ?", id).Scan(
		&task.ID,
		&task.Name,
		&task.Project,
		&task.Status,
		&task.Created,
	)
	return task, err
}
