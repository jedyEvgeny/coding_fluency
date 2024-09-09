//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 41 мин
package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	dbName    string
	sqlDriver string
	timeout   int
	db        *sql.DB
}

type Item struct {
	ID    int
	Name  string
	Price float64
}

func main() {
	a := mustNew()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func mustNew() App {
	var (
		name    = "db-go.db"
		timeout = 3000
		driver  = "sqlite3"
	)
	db, err := initDatabase(name, timeout, driver)
	if err != nil {
		log.Fatal(err)
	}
	return App{
		db:        db,
		dbName:    name,
		timeout:   timeout,
		sqlDriver: driver,
	}

}

func initDatabase(name string, timeout int, driver string) (*sql.DB, error) {
	db, err := sql.Open(driver, createName(name, timeout))
	if err != nil {
		return nil, fmt.Errorf("не смогли открыть БД %s: %w", name, err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не смогли выплонить пинг к БД %s: %w", name, err)
	}

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY, name TEXT, price FLOAT);
	`
	if _, err = db.Exec(request); err != nil {
		return nil, fmt.Errorf("не смогли создать таблицу в БД %s: %w", name, err)
	}
	return db, nil
}

func createName(name string, timeout int) string {
	return fmt.Sprintf("%s?_timeout=%d", name, timeout)
}

func (a App) Run() error {
	defer func() { _ = a.db.Close() }()

	items := []Item{
		{Name: "Апельсин", Price: 125.99},
		{Name: "Персик", Price: 312.69},
	}
	var wg sync.WaitGroup
	for _, el := range items {
		wg.Add(1)
		go func(el Item) error {
			defer wg.Done()
			err := a.insertItem(el.Name, el.Price)
			if err != nil {
				return fmt.Errorf("не смогли добавить позицию в БД %s: %w", a.dbName, err)
			}
			return nil
		}(el)
	}
	wg.Wait()
	err := a.updateItem("Апельсин", 289.31)
	if err != nil {
		return fmt.Errorf("не смогли обновить позицию в БД %s: %w", a.dbName, err)
	}
	err = a.readData()
	if err != nil {
		return fmt.Errorf("не смогли прочитать данные в БД %s: %w", a.dbName, err)
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
		return fmt.Errorf("не смогли подготовить запрос на добавление позиции: %w", err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли исполнить запрос на добавление позиции %w", err)
	}
	return nil
}

func (a App) updateItem(name string, price float64) error {
	request := `
	UPDATE items
	SET price = :price
	WHERE name = :name
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не смогли подготовить запрос на обновление инфо: %w", err)
	}
	defer func() { _ = sqlStmt.Close() }()

	result, err := sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли исполнить запрос на изменение инфо: %w", err)
	}
	resultAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не смогли получить количество изменённых строк %w", err)
	}
	if resultAffected == 0 {
		return fmt.Errorf("нет изменённых строк %w", err)
	}
	return nil
}

func (a App) readData() error {
	request := `
	SELECT id, name, price 
	FROM items
	`
	rows, err := a.db.Query(request)
	if err != nil {
		return fmt.Errorf("не смогли прочитать данные: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		item := Item{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			return fmt.Errorf("не смогли прочитать строку в цикле: %w", err)
		}
		fmt.Printf("id: %d, товар: %s, цена: %.2f\n", item.ID, item.Name, item.Price)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("целостность данных нарушена %w", err)
	}

	return nil
}
