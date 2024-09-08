//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 31 мин
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
	timeout   int
	db        *sql.DB
}

type Item struct {
	Id    int
	Name  string
	Price float64
}

func main() {
	a := mustNew()
	err := a.Run()
	if err != nil {
		log.Fatal("ошибка сервиса ", err)
	}
}

func mustNew() App {
	var (
		name    = "go-sql.db"
		driver  = "sqlite3"
		timeout = 5000
	)
	db, err := initDatabase(name, driver, timeout)
	if err != nil {
		log.Fatal(err)
	}
	return App{
		dbName:    name,
		sqlDriver: driver,
		timeout:   timeout,
		db:        db,
	}
}

func initDatabase(name, driver string, timeout int) (*sql.DB, error) {
	db, err := sql.Open(driver, createName(name, timeout))
	if err != nil {
		return nil, fmt.Errorf("не смогли открыть БД %s: %w", name, err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не смогли выполнить пинг к БД %s: %w", name, err)
	}

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY, name TEXT, price FLOAT)
	`
	if _, err = db.Exec(request); err != nil {
		return nil, fmt.Errorf("не смогли создать таблицу в БД %s: %w", name, err)
	}
	log.Println("база данных создана")
	return db, nil
}

func createName(name string, timeout int) string {
	return fmt.Sprintf("%s?_timeout=%d", name, timeout)
}

func (a App) Run() error {
	defer func() { _ = a.db.Close() }()

	err := a.insertItem("Orange", 159.69)
	if err != nil {
		return fmt.Errorf("не смогли добавить позицию в БД %s: %w", a.dbName, err)
	}

	err = a.readData()
	if err != nil {
		return fmt.Errorf("не смогли прочитать БД %s: %w", a.dbName, err)
	}

	err = a.updateData(2, 356.90)
	if err != nil {
		return fmt.Errorf("не удалось обновить даннее %w", err)
	}
	err = a.readData()
	if err != nil {
		return fmt.Errorf("не смогли прочитать БД %s: %w", a.dbName, err)
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
		return fmt.Errorf("не смогли создать запрос для вставки инфо в БД %s: %w", a.dbName, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли исполнить запрос для вставки в БД %s: %w", a.dbName, err)
	}
	log.Println("вставка выполнена")
	return nil
}

func (a App) readData() error {
	request := `
	SELECT id, name, price
	FROM items
	`
	rows, err := a.db.Query(request)
	if err != nil {
		return fmt.Errorf("не смогли получить данные из БД %s: %w", a.dbName, err)
	}
	defer func() { _ = rows.Close() }()

	item := Item{}
	for rows.Next() {
		err = rows.Scan(&item.Id, &item.Name, &item.Price)
		if err != nil {
			log.Printf("ошибка чтения из БД %s: %v\n", a.dbName, err)
			continue
		}
		fmt.Printf("id: %d, товар: %s, цена: %.2f\n", item.Id, item.Name, item.Price)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("целостность полученных из БД %s данных нарушена: %w", a.dbName, err)
	}
	return nil
}

func (a App) updateData(id int, price float64) error {
	request := `
	UPDATE items
	SET price = :price
	WHERE id = :id
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не смогли подготовить данные для обновления в БД %s: %w", a.dbName, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("id", id),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли обновить данные в БД %s: %w", a.dbName, err)
	}
	log.Println("обновили данные")
	return nil

}
