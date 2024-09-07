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
	log.Println("БД успешно создана")
}

func initDataBase() error {
	db, err := sql.Open("sqlite3", "go-db.db")
	if err != nil {
		return fmt.Errorf("не смогли открыть БД: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("не смогли установить связь с БД: %w", err)
	}

	sqlStmt := `CREATE TABLE IF NOT EXISTS items(id INTEGER PRIMARY KEY, name VARCHAR, cost FLOAT);`
	if _, err = db.Exec(sqlStmt); err != nil {
		return fmt.Errorf("не смогли создать таблицу в БД: %w", err)
	}

	return nil
}
