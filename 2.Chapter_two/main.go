//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД


package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "go-engeneer.db"

type Item struct {
	ID    int
	Name  string
	Price float64
}

func main() {

	initDatabase()
}

func initDatabase() {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}
	log.Printf("Соединение с БД %s установлено\n", dbName)

	sqlStmt := `CREATE TABLE IF NOT EXISTS items (id INTEGER PRIMARY KEY, name TEXT, price FLOAT);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
