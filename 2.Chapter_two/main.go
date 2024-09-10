//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 41 мин
package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	wg        sync.WaitGroup
	mu        sync.Mutex
	dbName    string
	sqlDriver string
	db        *sql.DB
	timeout   int
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
		log.Fatal("ошибка сервиса: ", err)
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
		return nil, fmt.Errorf("не смогли открыть БД %s: %w", name, err)
	}
	start := time.Now()
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не смогли выполнить пинг к БД %s: %w", name, err)
	}
	end := time.Now()
	log.Println("Пинг выполнен за", end.Sub(start))

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY, name TEXT, price FLOAT)
	`
	if _, err = db.Exec(request); err != nil {
		return nil, fmt.Errorf("не смогли создать таблицу в БД %s: %w", name, err)
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
		{Name: "Виноград", Price: 326.60},
		{Name: "Арбуз", Price: 290.85},
	}
	var errors []error
	for _, el := range items {
		a.wg.Add(1)
		go func(el Item) {
			defer a.wg.Done()
			err := a.insertItem(el.Name, el.Price)
			if err != nil {
				a.mu.Lock()
				errors = append(errors, err)
				a.mu.Unlock()
			}
		}(el)
	}
	a.wg.Wait()
	if len(errors) > 0 {
		for _, el := range errors {
			log.Printf("ошибка добавления элемента в БД %s: %v\n", a.dbName, el)
		}
		return fmt.Errorf("количество ошибок: %d", len(errors))
	}

	err := a.updateItem("Апельсин", 256.00)
	if err != nil {
		return fmt.Errorf("не смогли обновить данные в БД %s: %w", a.dbName, err)
	}

	err = a.deleteItem(6)
	if err != nil {
		return fmt.Errorf("не смогли удалить данные в БД %s: %w", a.dbName, err)
	}

	err = a.readItems()
	if err != nil {
		return fmt.Errorf("не смогли прочитать данные в БД %s: %w", a.dbName, err)
	}
	return nil
}

func (a *App) insertItem(name string, price float64) error {
	request := `
	INSERT INTO items
	(price, name)
	VALUES (:price, :name)
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не смогли подготовить запрос для вставки: %w", err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли выполнить запрос для вставки: %w", err)
	}
	log.Println("вставка выполнена успешно")
	return nil
}

func (a *App) updateItem(name string, price float64) error {
	request := `
	UPDATE items
	SET price = :price
	WHERE name = :name
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не смогли подготовить запрос на изменение %w", err)
	}
	defer func() { _ = sqlStmt.Close() }()

	_, err = sqlStmt.Exec(
		sql.Named("name", name),
		sql.Named("price", price),
	)
	if err != nil {
		return fmt.Errorf("не смогли обновить данные: %w", err)
	}
	log.Println("данные обновлены")
	return nil
}

func (a *App) deleteItem(id int) error {
	request := `
	DELETE FROM items
	WHERE id=:id
	`
	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf("не смогли подготовить запрос на удаление: %w", err)
	}
	defer func() { _ = sqlStmt.Close() }()

	result, err := sqlStmt.Exec(
		sql.Named("id", id),
	)
	if err != nil {
		return fmt.Errorf("не смогли удалить позицию: %w", err)
	}
	resultAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не смогли получить количество удалённых строк: %w", err)
	}
	if resultAffected == 0 {
		log.Println("не найдено строк для удаления")
		return nil
	}
	log.Println("удалили позицию")
	return nil
}

func (a *App) readItems() error {
	request := `
	SELECT id, name, price
	FROM items
	`
	rows, err := a.db.Query(request)
	if err != nil {
		return fmt.Errorf("не удалось прочитать данные: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		item := Item{}
		if err = rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			return fmt.Errorf("не удалось прочитать данные в цикле: %w", err)
		}
		fmt.Printf("id: %d, товар: %s, цена: %.2f\n", item.ID, item.Name, item.Price)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("целостность данных не подтверждена: %w", err)
	}
	log.Println("данные успешно прочитаны")
	return nil
}
