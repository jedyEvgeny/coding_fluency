//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 34 мин
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
	errStmt        = "не удалось подготовить выражение: %w"
	errExec        = "не удалось выполнить выражение: %w"
	errAff         = "не удалось получить количество изменённых строк: %w"
	errCounAff     = "нет изменений в БД: %w"
	errRead        = "не удалось прочитать данные: %w"
	errReadInCycle = "не удалось прочитать данные в цикле: %w"
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
		name    = "dbsql.db"
		driver  = "sqlite3"
		timeout = 1000
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
	CREATE TABLE IF NOT EXISTS
	items (id INTEGER PRIMARY KEY, name TEXT, price FLOAT);
	CREATE INDEX IF NOT EXISTS idx_items_name ON items(name)
	`
	sqlStmt, err := db.Prepare(request)
	if err != nil {
		return nil, fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	if _, err = sqlStmt.Exec(); err != nil {
		return nil, fmt.Errorf(errExec, err)
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
		{Name: "Вино", Price: 8995.95},
		{Name: "Медовуха", Price: 699.95},
		{Name: "Коньяк", Price: 12000.95},
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
			log.Printf("не удалось выполнить вставку в БД %s: %v\n", a.dbName, err)
		}
		return fmt.Errorf("количество ошибок вставки: %d", len(errors))
	}

	err := a.updateItem("Вино", 7999.80)
	if err != nil {
		return fmt.Errorf("не удалось обновить данные в БД %s: %w", a.dbName, err)
	}

	err = a.deleteItem(15)
	if err != nil {
		return fmt.Errorf("не удалось удалить данные в БД %s: %w", a.dbName, err)
	}

	err = a.readItems()
	if err != nil {
		return fmt.Errorf("не удалось прочитать данные из БД %s: %w", a.dbName, err)
	}

	return nil
}

func (a *App) insertItem(name string, price float64) error {
	request := `
	INSERT INTO items
	(name, price)
	VALUES(:name, :price)
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf(errExec, err)
	}
	log.Println("вставка выполнена")
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
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	res, err := sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf(errExec, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(errAff, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(errCounAff, err)
	}
	log.Println("обновление выполнено")
	return nil
}

func (a *App) deleteItem(id int) error {
	request := `
	DELETE FROM items
	WHERE id=:id
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	res, err := sqlStmt.Exec(
		sql.Named("id", id),
	)
	if err != nil {
		return fmt.Errorf(errExec, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(errAff, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(errCounAff, err)
	}
	log.Println("удаление выполнено")
	return nil
}

func (a *App) readItems() error {
	request := `
	SELECT id, name, price
	FROM items
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	rows, err := sqlStmt.Query()
	if err != nil {
		return fmt.Errorf(errRead, err)
	}
	for rows.Next() {
		item := Item{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			return fmt.Errorf(errReadInCycle, err)
		}
		fmt.Printf("id: %d, товар: %s, цена: %.2f\n", item.ID, item.Name, item.Price)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("целостность данных нарушена: %w", err)
	}
	log.Println("чтение выполнено")
	return nil
}
