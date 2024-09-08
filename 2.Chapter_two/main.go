//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 29 мин
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	dbName    string
	sqlDriver string
	db        *sql.DB
}

type Item struct {
	id    int
	name  string
	price float64
}

func main() {
	a := New()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func New() App {
	var (
		dbName    = "go-db.db"
		sqlDriver = "sqlite3"
	)
	db, err := initDataBase(dbName, sqlDriver)
	if err != nil {
		log.Fatal(err)
	}
	return App{
		dbName:    dbName,
		sqlDriver: sqlDriver,
		db:        db,
	}
}

func initDataBase(dbName, sqlDriver string) (*sql.DB, error) {
	db, err := sql.Open(sqlDriver, dbName)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть БД %s: %w", dbName, err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось выполнить пинг к БД %s: %w", dbName, err)
	}

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY, name TEXT, price FLOAT)
	`
	if _, err = db.Exec(request); err != nil {
		return nil, fmt.Errorf("не удалось создать таблицу в БД %s: %w", dbName, err)
	}
	log.Println("БД и таблица успешно созданы")
	return db, nil
}

func (a App) Run() error {
	defer func() { _ = a.db.Close() }()

	err := a.insertItem("Orange", 129.99)
	if err != nil {
		log.Println("не смогли добавить запись в БД", err)
	}
	err = a.insertItem("Rastberries", 269.99)
	if err != nil {
		log.Println("не смогли добавить запись в БД", err)
	}

	err = a.data()
	if err != nil {
		log.Println("не смогли прочесть БД", err)
	}

	err = a.ipdateItem("Rastberries", 529.90)
	if err != nil {
		log.Println("не удалось изменить данные", err)
	}

	return nil
}

func (a App) insertItem(name string, price float64) error {
	request := `
	INSERT INTO items
	(name, price)
	VALUES (:name, :price)
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не создать запрос к БД %s: %w", a.dbName, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли исполнить запрос к БД %s: %w", a.dbName, err)
	}
	log.Println("вставка прошла успешно")
	return nil
}

func (a App) data() error {
	request := `
	SELECT id, name, price
	FROM items
	`
	rows, err := a.db.Query(request)
	if err != nil {
		return fmt.Errorf("не смогли извлечь инфо из БД %s: %w", a.dbName, err)
	}
	item := Item{}
	for rows.Next() {
		err = rows.Scan(&item.id, &item.name, &item.price)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(item)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("неполное чтение данных из БД %s: %w", a.dbName, err)
	}
	return nil
}

func (a App) ipdateItem(name string, price float64) error {
	request := `
	UPDATE items 
	SET price = :price
	WHERE name = :name
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не удалось подготовить запрос для изменения БД %s: %w", a.dbName, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не удалось изменить БД %s: %w", a.dbName, err)
	}
	log.Println("данные изменены")
	return nil
}
