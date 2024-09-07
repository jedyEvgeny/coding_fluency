//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 9 мин 32 сек

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	nameDB   = "go-db.db"
	driverDB = "sqlite3"
)

func main() {
	err := initDataBase()
	if err != nil {
		log.Fatal("не удалось создать БД: %w", err)
	}
	err = insertItem("Orange", 159.99)
	if err != nil {
		log.Fatal("не удалось создать запись в БД: %w", err)
	}
}

func initDataBase() error {
	db, err := sql.Open(driverDB, nameDB)
	if err != nil {
		return fmt.Errorf("не удалось открыть БД: %w", err)
	}
	defer func() { _ = db.Close() }()

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRYMARI KEY, name VARCHAR, cost FLOAT);
	`

	if _, err = db.Exec(request); err != nil {
		return fmt.Errorf("не удалось создать таблицу в БД %s: %w", nameDB, err)
	}
	log.Println("Таблица в БД создана")
	return nil
}

func insertItem(name string, cost float64) error {
	db, err := sql.Open(driverDB, nameDB)
	if err != nil {
		return fmt.Errorf("не удалось открыть БД: %w", err)
	}
	defer func() { _ = db.Close() }()

	request := `
	INSERT INTO items
	(name, cost)
	VALUES(?,?)
	`
	stmtSql, err := db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не удалось подготовить запрос к БД %s: %w", name, err)
	}

	if _, err = stmtSql.Exec(name, cost); err != nil {
		return fmt.Errorf("не удалось выполнить запрос в БД %s: %w", nameDB, err)
	}
	log.Printf("вставка в БД %s выполнена\n", nameDB)
	return nil
}

// type Item struct {
// 	ID   int
// 	Name string
// 	Cost float64
// }

// func read() error {
// 	db, err := sql.Open(driver, dbName)
// 	if err != nil {
// 		return fmt.Errorf("не смогли открыть БД %s: %w", dbName, err)
// 	}
// 	defer func() { _ = db.Close() }()
// 	request := `
// 	SELECT id, name, cost FROM items
// 	`
// 	rows, err := db.Query(request)
// 	if err != nil {
// 		return fmt.Errorf("не удалось получить данные из БД %s: %w", dbName, err)
// 	}
// 	defer func() { _ = rows.Close() }()
// 	var items []Item
// 	for rows.Next() {
// 		var item Item
// 		if err = rows.Scan(&item.ID, &item.Name, &item.Cost); err != nil {
// 			return fmt.Errorf("не удалось получить данные из БД %s: %w", dbName, err)
// 		}
// 		items = append(items, item)
// 	}
// 	for _, item := range items {
// 		fmt.Printf("ID: %d; наименование: %s; цена: %.2f\n", item.ID, item.Name, item.Cost)
// 	}
// 	return nil
// }
