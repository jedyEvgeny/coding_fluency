//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 45 мин
package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	wg        sync.WaitGroup
	mu        sync.Mutex
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

const (
	prepareErr   = "ошибка подготовки запроса: %w"
	execErr      = "ошибка исполнения sql-команды: %w"
	affErr       = "ошибка чтения количества изменённых строк: %w"
	amountAffErr = "нет изменённых строк: %w"
	readErr      = "ошибка чтения: %w"
	readCicleErr = "ошибка чтения в цикле: %w"
)

func main() {
	a := mustNew()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func mustNew() App {
	var (
		name    = "go-sql.db"
		driver  = "sqlite3"
		timeout = 50
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
		return nil, fmt.Errorf("не удалось открыть БД %s: %w", name, err)
	}
	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY, name TEXT, price FLOAT);
	CREATE INDEX IF NOT EXISTS idx_items_name ON items(name)
	`
	sqlStmt, err := db.Prepare(request)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}
	defer func() { _ = sqlStmt.Close() }()

	if _, err = sqlStmt.Exec(); err != nil {
		return nil, fmt.Errorf("не удалось создать таблицу в БД %s: %w", name, err)
	}
	log.Println("таблица создана")
	return db, nil
}

func createName(name string, timeout int) string {
	return fmt.Sprintf("%s?_timeout=%d", name, timeout)
}

func (a *App) Run() error {
	defer func() { _ = a.db.Close() }()

	items := []Item{
		{Name: "Апельсин", Price: 129.90},
		{Name: "Ананас", Price: 239.90},
		{Name: "Авокадо", Price: 333.33},
	}
	var errors []error
	for _, item := range items {
		a.wg.Add(1)
		go func(item Item) {
			defer a.wg.Done()
			err := a.insertItem(item.Name, item.Price)
			if err != nil {
				a.mu.Lock()
				errors = append(errors, err)
				a.mu.Unlock()

			}
		}(item)
	}
	a.wg.Wait()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("ошибка вставки в БД %s: %v\n", a.dbName, err)
		}
		return fmt.Errorf("количество ошибок вставки: %d", len(errors))
	}

	err := a.updateItem("Апельсин", 99.99)
	if err != nil {
		return fmt.Errorf("не удалось обновить строку в БД %s: %w", a.dbName, err)
	}

	err = a.deleteItem(12)
	if err != nil {
		return fmt.Errorf("не удалось удалить строку в БД %s: %w", a.dbName, err)
	}

	err = a.readItems()
	if err != nil {
		return fmt.Errorf("не удалось прочитать БД %s: %w", a.dbName, err)
	}

	return nil
}

func (a *App) insertItem(name string, price float64) error {
	request := `
	INSERT INTO items
	(name, price)
	VALUES (:name, :price)
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(prepareErr, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	res, err := sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf(execErr, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(affErr, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(amountAffErr, err)
	}
	log.Println("вставка выполнена успешно")
	return nil
}

func (a *App) updateItem(name string, price float64) error {
	request := `
UPDATE items
SET price=:price
WHERE name=:name
`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(prepareErr, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	res, err := sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf(execErr, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(affErr, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(amountAffErr, err)
	}
	log.Println("изменение выполнено успешно")
	return nil
}

func (a *App) deleteItem(id int) error {
	request := `
DELETE FROM items
WHERE id=:id
`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(prepareErr, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	res, err := sqlStmt.Exec(
		sql.Named("id", id),
	)
	if err != nil {
		return fmt.Errorf(execErr, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(affErr, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(amountAffErr, err)
	}
	log.Println("удаление выполнено успешно")
	return nil
}

func (a *App) readItems() error {
	request := `
SELECT id, name, price
FROM items
`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(prepareErr, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	rows, err := sqlStmt.Query()
	if err != nil {
		return fmt.Errorf(readErr, err)
	}
	for rows.Next() {
		item := Item{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			return fmt.Errorf(readCicleErr, err)
		}
		fmt.Printf("id: %d, товар: %s, цена: %.2f\n", item.ID, item.Name, item.Price)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("ошибка целостности данных: %w", err)
	}

	log.Println("чтение выполнено успешно")
	return nil
}
