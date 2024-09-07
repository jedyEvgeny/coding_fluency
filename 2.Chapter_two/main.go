//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 19 мин 19 сек

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type data struct {
	name string
	id   int
	cost float64
}

var (
	dbName = "go-db.db"
	driver = "sqlite3"
)

func main() {
	err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = insertItem("Orange", 152.99)
	if err != nil {
		log.Fatal(err)
	}
	err = insertItem("Яблоко", 165.35)
	if err != nil {
		log.Fatal(err)
	}
	err = info()
	if err != nil {
		log.Fatal(err)
	}
}

func initDatabase() error {
	db, err := sql.Open(driver, dbName)
	if err != nil {
		return fmt.Errorf("не смогли открыть БД %s: %w", dbName, err)
	}
	defer func() { _ = db.Close() }()

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY, name TEXT, cost FLOAT)
	`
	if _, err = db.Exec(request); err != nil {
		return fmt.Errorf("не смогли создать таблицу в БД %s: %w", dbName, err)
	}
	log.Println("Создали таблицу в БД", dbName)
	return nil
}

func insertItem(name string, cost float64) error {
	db, err := sql.Open(driver, dbName)
	if err != nil {
		return fmt.Errorf("не смогли открыть БД %s: %w", dbName, err)
	}
	defer func() { _ = db.Close() }()

	request := `
	INSERT INTO items
	(name, cost)
	VALUES(?,?)
	`
	sqlStmt, err := db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не удалось подготовить запрос к БД %s: %w", dbName, err)
	}

	if _, err = sqlStmt.Exec(name, cost); err != nil {
		return fmt.Errorf("не удалось добавить информацию в БД %s: %w", dbName, err)
	}
	log.Println("информация успешно добавлена в БД", dbName)
	return nil
}

func info() error {
	db, err := sql.Open(driver, dbName)
	if err != nil {
		return fmt.Errorf("не смогли открыть БД %s: %w", dbName, err)
	}
	defer func() { _ = db.Close() }()

	request := `
	SELECT id, name, cost 
	FROM items
	`
	rows, err := db.Query(request)
	if err != nil {
		return fmt.Errorf("не смогли выполнить запрос к БД %s: %w", dbName, err)
	}
	defer func() { _ = rows.Close() }()

	items := []data{}
	for rows.Next() {
		var item data
		if err = rows.Scan(&item.id, &item.name, &item.cost); err != nil {
			return fmt.Errorf("не смогли прочитать данные из БД %s: %w", dbName, err)
		}
		items = append(items, item)
	}
	for _, item := range items {
		fmt.Printf("ID: %d, наименование: %s, цена товара: %.2f\n", item.id, item.name, item.cost)
	}
	return nil
}
