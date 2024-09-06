//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 4 мин 42 сек.


package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := initDataBase()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("База данных создана")
}

func initDataBase() error {
	db, err := sql.Open("sqlite3", "go-engenier.db")
	if err != nil {
		return fmt.Errorf("не смогли открыть БД: %w", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("не смогли подключиться к БД: %w", err)
	}

	sqlStmt := `CREATE TABLE IF NOT EXISTS items(id INTEGER PRIMARY KEY, name VARCHAR, cost FLOAT);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("не смогли создать таблицу: %w", err)
	}
	return nil
}
