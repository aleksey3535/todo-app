package repository

import (
	"fmt"
	"strings"
	"todo-app"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type TodoListPostgres struct {
	db *sqlx.DB
}

func NewTodoListPostgres (db *sqlx.DB) (*TodoListPostgres) {
	return &TodoListPostgres {
		db: db,
	}
}

func (t *TodoListPostgres) Create(userId int, list todo.TodoList)  (int, error) {
	tx, err := t.db.Begin()
	if err != nil {
		return 0, err
	}
	var listId int
	createListQuery := fmt.Sprintf("INSERT INTO %s (title, description) values ($1, $2) RETURNING id", todoListsTable)
	row := tx.QueryRow(createListQuery, list.Title, list.Description)
	if err:= row.Scan(&listId); err != nil {
		return 0, err
	}
	createUsersListQuery := fmt.Sprintf("INSERT INTO %s (user_id, list_id) values ($1, $2)", usersListsTable)
	_, err = tx.Exec(createUsersListQuery, userId, listId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	return listId, tx.Commit()
}

func (t *TodoListPostgres) GetAll(userID int) ([]todo.TodoList, error) {
	var lists []todo.TodoList
	query := fmt.Sprintf("SELECT tl.id, tl.title, tl.description FROM %s tl INNER JOIN %s ul on tl.id = ul.list_id WHERE ul.user_id = $1",
		todoListsTable, usersListsTable)
	err := t.db.Select(&lists, query, userID )
	return lists, err
}

func (t *TodoListPostgres) GetById(userId, listId int) (todo.TodoList ,error) {
	var list todo.TodoList
	query := fmt.Sprintf("SELECT tl.id, tl.title, tl.description FROM %s tl INNER JOIN %s ul on tl.id = ul.list_id WHERE ul.user_id = $1 AND tl.id = $2",
		todoListsTable, usersListsTable)
	err := t.db.Get(&list, query, userId, listId)
	return list, err
}

func (t *TodoListPostgres) Delete(userId, listId int) error {
	query := fmt.Sprintf("DELETE FROM %s tl USING %s ul WHERE tl.id = ul.list_id AND ul.user_id = $1 AND ul.list_id = $2", 
	todoListsTable, usersListsTable)
	_, err := t.db.Exec(query, userId, listId)
	return err
}

func(t *TodoListPostgres) Update(userId, listId int, input todo.UpdateListInput) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0 )
	argId := 1
	if input.Title != nil {
		setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
		args = append(args, *input.Title)
		argId++
	}
	if input.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description=$%d", argId))
		args = append(args, *input.Description)
		argId++
	}
	setQuery := strings.Join(setValues, ", ")
	query := fmt.Sprintf("UPDATE %s tl SET %s FROM %s ul WHERE tl.id = ul.list_id AND ul.list_id=$%d AND ul.user_id=$%d", 
	todoListsTable, setQuery, usersListsTable, argId, argId+1)
	args = append(args, listId, userId)
	logrus.Debugf("updateQuery:%s", query)
	logrus.Debugf("args: %s", args)
	_, err := t.db.Exec(query, args...)
	return err

}