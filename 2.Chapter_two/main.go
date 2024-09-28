//Это вторая часть codding-fluency: довожу до автоматизма навык работы с реляционными базами данных
//Начну с основ - SQLite: создание-изменение-удаление таблиц и БД
//Лучшее время набора 31 мин
// Повторно вернулся спустя примерно две недели и писал код около часа без спешки. Основная проблема была с вспоминанием sql-команд

// package main

// import (
// 	"database/sql"
// 	"log"

// 	"github.com/golang-migrate/migrate/v4"
// 	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// )

// func main() {
// 	db, err := sql.Open("sqlite3", "mydatabase.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	dbName := "sqlite3://mydatabase.db"
// 	migrationPath := "file://migrations" // Используем относительный путь

// 	// Создаем мигратор
// 	m, err := migrate.New(migrationPath, dbName) // передаем правильный путь
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Применяем миграции
// 	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
// 		log.Fatal(err)
// 	}
// 	if err == migrate.ErrNoChange {
// 		log.Println("миграции не требуются")

// 	}
// 	if err != migrate.ErrNoChange {
// 		log.Println("миграции применены успешно!")
// 	}

// 	//Пробуем работать с БД
// 	_, err = db.Exec("INSERT INTO items (name, price) VALUES (?, ?)", "Item2", 64.34)
// 	if err != nil {
// 		log.Fatalf("ошибка вставки: %v", err)
// 	}

// 	// Получить данные
// 	rows, err := db.Query("SELECT id, name, price FROM items")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var id int
// 		var name string
// 		var price float64
// 		if err := rows.Scan(&id, &name, &price); err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Printf("Item: %d, Name: %s, Price: %.2f\n", id, name, price)
// 	}

// 	// Удалить элемент
// 	_, err = db.Exec("DELETE FROM items WHERE name = ?", "Item1")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	dbName    string
	timeout   int
	sqlDriver string
	db        *sql.DB
	wg        sync.WaitGroup
	mu        sync.Mutex
}
type Item struct {
	id    int
	goods string
	price float64
}

const (
	errOpen        = "не смогли откыть БД: %w"
	errPing        = "не смогли выполнить пинг к БД: %w"
	errStmt        = "не смогли подготовить sql-выражение: %w"
	errExec        = "не смогли выполнить выражение: %w"
	errRes         = "не смогли получить результат выполнения sql-запроса: %w"
	errResAffected = "не выполнены изменения в БД: %w"
	errRead        = "не смогли получить БД: %w"
	errReadCycle   = "не смогли прочитать полученные данные: %w"
	errReadConsist = "целостность данных при чтении нарушена: %w"
)

func mustNew() App {
	var (
		name    = "go_sql.db"
		driver  = "sqlite3"
		timeout = 50
	)
	db, err := initDatabase(name, driver, timeout)
	if err != nil {
		log.Fatal(err)
	}
	return App{
		dbName:    name,
		timeout:   timeout,
		sqlDriver: driver,
		db:        db,
	}
}

func main() {
	a := mustNew()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func initDatabase(name, driver string, timeout int) (*sql.DB, error) {
	db, err := sql.Open(driver, createName(name, timeout))
	if err != nil {
		return nil, fmt.Errorf(errOpen, err)
	}
	startPing := time.Now()
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf(errPing, err)
	}
	endPing := time.Now()
	log.Printf("Пинг выполнен за %v\n", endPing.Sub(startPing))

	request := `
	CREATE TABLE IF NOT EXISTS items
	(id INTEGER PRIMARY KEY,
	goods TEXT NOT NULL,
	price FLOAT);
	CREATE INDEX IF NOT EXISTS idx_items_goods 
	ON items(goods);
	`

	sqlStmt, err := db.Prepare(request)
	if err != nil {
		return nil, fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	if _, err = sqlStmt.Exec(); err != nil {
		return nil, fmt.Errorf(errExec, err)
	}
	log.Println("БД и таблица созданы")
	return db, err
}

func createName(name string, timeout int) string {
	return fmt.Sprintf("%s?_timeout=%d", name, timeout)
}

func (a *App) Run() error {
	defer func() { _ = a.db.Close() }()

	items := []Item{
		{goods: "Апельсин", price: 229.50},
		{goods: "Манго", price: 165.70},
		{goods: "Персик", price: 364.20},
	}
	var errors []error
	for _, el := range items {
		a.wg.Add(1)
		go func(item Item) {
			defer a.wg.Done()
			err := a.insertItem(el.goods, el.price)
			if err != nil {
				a.mu.Lock()
				errors = append(errors, err)
				a.mu.Unlock()
			}
		}(el)
	}
	a.wg.Wait()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Println("Ошибка вставки: ", err)
		}
		return fmt.Errorf("количество ошибок вставки: %d", len(errors))
	}

	err := a.updateItem("Апельсин", 225.60)
	if err != nil {
		return fmt.Errorf("не смогли обновить данные: %w", err)
	}

	err = a.deleteItem(11)
	if err != nil {
		return fmt.Errorf("не смогли удалить данные: %w", err)
	}

	err = a.readDatabase()
	if err != nil {
		return fmt.Errorf("не смогли прочитать данные: %w", err)
	}

	return nil
}

func (a *App) insertItem(goods string, price float64) error {
	request := `
	INSERT INTO items
	(goods, price)
	VALUES(:goods, :price)
	`

	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	startInsert := time.Now()
	res, err := sqlStmt.Exec(
		sql.Named("goods", goods),
		sql.Named("price", price),
	)
	endInsert := time.Now()

	if err != nil {
		return fmt.Errorf(errExec, err)
	}
	resAffected, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf(errRes, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(errResAffected, err)
	}

	log.Println("информация внесена в БД за время:", endInsert.Sub(startInsert))
	return nil
}

func (a *App) updateItem(goods string, price float64) error {
	request := `
	UPDATE items
	SET price = :price
	WHERE goods=:goods
	`

	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	startExec := time.Now()
	res, err := sqlStmt.Exec(
		sql.Named("goods", goods),
		sql.Named("price", price),
	)
	endExec := time.Now()

	if err != nil {
		return fmt.Errorf(errExec, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(errRes, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(errResAffected, err)
	}

	log.Println("БД обновлена за время:", endExec.Sub(startExec))
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

	startExec := time.Now()
	res, err := sqlStmt.Exec(
		sql.Named("id", id),
	)
	endExec := time.Now()

	if err != nil {
		return fmt.Errorf(errExec, err)
	}
	resAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(errRes, err)
	}
	if resAffected == 0 {
		return fmt.Errorf(errResAffected, err)
	}

	log.Println("Удаление выполнено за время:", endExec.Sub(startExec))
	return nil
}

func (a *App) readDatabase() error {
	request := `
	SELECT * FROM items
	`

	sqlStmt, err := a.db.Prepare(request)
	if err != nil {
		return fmt.Errorf(errStmt, err)
	}
	defer func() { _ = sqlStmt.Close() }()

	startExec := time.Now()
	rows, err := sqlStmt.Query()
	endExec := time.Now()
	if err != nil {
		return fmt.Errorf(errRead, err)
	}
	defer func() { _ = rows.Close() }()

	fmt.Println(strings.Repeat("*", 30))
	for rows.Next() {
		item := Item{}
		if err = rows.Scan(&item.id, &item.goods, &item.price); err != nil {
			return fmt.Errorf(errReadCycle, err)
		}
		fmt.Printf("id: %d, товар: %s, цена: %.2f\n", item.id, item.goods, item.price)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf(errReadConsist, err)
	}
	fmt.Println(strings.Repeat("*", 30))
	log.Println("Чтение выполнено за время:", endExec.Sub(startExec))
	return nil
}
