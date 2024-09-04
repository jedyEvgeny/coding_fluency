// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Основа кода - поиск наиболее часто встречающихся слов среди нескольких заранее подготовленных текстов
// Запускаем с аргументом, примерно так: go run main.go -name Алиса ./short_files

// Лучшее время набора обоих файлов = 1 час 10 мин
// ЗЫ - нужно заранее создать файлы с текстом, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с удаления предыдущих файлов .go, котороые следом создаю через консоль touch main.go 
package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

type App struct {
	wg                 sync.WaitGroup
	mu                 sync.Mutex
	filesDir           string
	maxTopWord         int
	hostClient         string
	methodClient       string
	endpointClient     string
	schemeClient       string
	portServer         string
	enpointServerWords string
	enpointServerJson  string
	enpointServerImage string
}

type frequencyWord struct {
	word     string
	frequncy int
}

type Response struct {
	Info     string `json:"Информация о запросе"`
	FileName string `json:"Имя файла"`
	Content  string `json:"Содержимое файла"`
}

const perm = 0744

var (
	maxTopWord         = 10
	filesDir           = "./files"
	hostClient         = "jsonplaceholder.typicode.com"
	methodClient       = ""
	endpointClient     = "users"
	schemeClient       = "https"
	portServer         = ":8080"
	enpointServerWords = "/words"
	enpointServerJson  = "/json"
	enpointServerImage = "/image"
)

func New() App {
	return App{
		filesDir:           findFilesDir(),
		maxTopWord:         maxTopWord,
		hostClient:         hostClient,
		methodClient:       methodClient,
		endpointClient:     endpointClient,
		schemeClient:       schemeClient,
		portServer:         portServer,
		enpointServerWords: enpointServerWords,
		enpointServerJson:  enpointServerJson,
		enpointServerImage: enpointServerImage,
	}
}

func findFilesDir() string {
	l := len(os.Args)
	if l > 3 {
		filesDir = os.Args[l-1]
	}
	return filesDir
}

func main() {
	initConfig()
	a := New()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	name := flag.String(
		"name",
		"гость",
		"имя пользователя",
	)
	flag.Parse()
	template := "02.01.2006 15:03"
	now := time.Now()

	nameStrUpChars := strings.ToUpper(*name)
	fmt.Printf("\nПривет, %s! Сегодня %v\n\n", nameStrUpChars, now.Format(template))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s, введи $ для старта сервиса:\n", *name)
		if !scanner.Scan() {
			log.Println("не удалось считать данные:", scanner.Err())
		}
		phrase := scanner.Text()
		ok := findKey(phrase)
		if ok {
			break
		}
	}
	symbols := []string{"/", "|", "-", "\\"}
	end := time.Now().Add(1 * time.Second)
	for time.Now().Before(end) {
		for _, el := range symbols {
			fmt.Print("\rЗагружаем сервис:" + el)
			time.Sleep(150 * time.Millisecond)
		}
	}
	fmt.Print("\rСервис загружен!     \n")
	time.Sleep(500 * time.Millisecond)
}

func findKey(phrase string) bool {
	target := "$"
	idxTarget := strings.IndexRune(phrase, []rune(target)[0])
	log.Println(idxTarget)
	return strings.Contains(phrase, target)
}

func (a *App) Run() error {
	filesList, err := os.ReadDir(a.filesDir)
	if err != nil {
		return fmt.Errorf("нет файлов: %w", err)
	}
	allWords := make([]string, 0, a.maxTopWord)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		a.wg.Add(1)
		go func(entry fs.DirEntry) {
			defer a.wg.Done()
			allWords = a.findWords(allWords, entry)
		}(fileEntry)
	}
	uniqueWords := make(map[string]int)
	for _, el := range allWords {
		uniqueWords[el]++
	}
	frequencyWords := make([]frequencyWord, 0, a.maxTopWord)
	for key, val := range uniqueWords {
		frequencyWords = append(frequencyWords, frequencyWord{key, val})
	}
	sort.Slice(frequencyWords, func(i, j int) bool {
		return frequencyWords[i].frequncy > frequencyWords[j].frequncy
	})

	var currFrequency, lastFrequency, topWord int
	var buf []string
	results := make([]string, 0, a.maxTopWord)

	for _, el := range frequencyWords {
		if topWord > a.maxTopWord {
			break
		}
		currFrequency = el.frequncy
		if currFrequency != lastFrequency && len(buf) > 0 {
			s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWord, len(buf), lastFrequency, buf)
			fmt.Print(s)
			results = append(results, s)
			buf = nil
		}
		if currFrequency != lastFrequency {
			lastFrequency = currFrequency
			topWord++
		}
		buf = append(buf, el.word)
	}
	if len(buf) > 0 && topWord < a.maxTopWord+1 {
		s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.\n", topWord, len(buf), lastFrequency)
		fmt.Print(s)
		results = append(results, s)
	}
	result := strings.Join(results, "")
	fPathRes, err := a.saveResult(result)
	if err != nil {
		return fmt.Errorf("не смогли сохранить результат: %w", err)
	}

	err = a.createRequest()
	if err != nil {
		return fmt.Errorf("не удалось выполнить запрос: %w", err)
	}

	err = a.createServer(fPathRes)
	if err != nil {
		return fmt.Errorf("не удалось создать сервер: %w", err)
	}
	return nil
}

func (a *App) findWords(slice []string, entry fs.DirEntry) []string {
	fPath := filepath.Join(a.filesDir, entry.Name())
	content, err := os.ReadFile(fPath)
	if err != nil {
		log.Println(err)
		return nil
	}
	words := strings.FieldsFunc(string(content), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	a.mu.Lock()
	slice = append(slice, words...)
	a.mu.Unlock()
	return slice
}

func (a *App) saveResult(result string) (string, error) {
	h := sha1.New()
	_, err := io.WriteString(h, result)
	if err != nil {
		return "", fmt.Errorf("не смогли создать хеш: %w", err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))
	fPath := filepath.Join(a.filesDir, hash)
	err = os.WriteFile(fPath, []byte(result), perm)
	if err != nil {
		return "", fmt.Errorf("не смогли создать файл: %w", err)
	}
	return fPath, nil
}

func (a *App) createRequest() error {
	u := url.URL{
		Scheme: a.schemeClient,
		Path:   path.Join(a.endpointClient, a.methodClient),
		Host:   a.hostClient,
	}

	query := url.Values{}
	query.Add("_limit", "3")
	u.RawQuery = query.Encode()

	log.Println(u.String())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("не смогли создать запрос: %w", err)
	}
	client := &http.Client{}

	start := time.Now()
	resp, err := client.Do(req)
	finish := time.Now()
	if err != nil {
		return fmt.Errorf("не удалось отправить запрос: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Println(resp.Status, finish.Sub(start))

	var data []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("не удалось расрифровать json: %w", err)
	}
	for _, user := range data {
		s := fmt.Sprintf("Имя: %v, телефон: %v, почта: %v\n", user["name"], user["phone"], user["email"])
		fmt.Print(s)
		fmt.Println(strings.Repeat("-", 25))
	}
	return nil
}

func (a *App) createServer(path string) error {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, path) }
	handlerJson := func(w http.ResponseWriter, r *http.Request) { handleJson(w, r, path) }

	mux := http.NewServeMux()
	mux.HandleFunc(a.enpointServerWords, handlerWords)
	mux.HandleFunc(a.enpointServerJson, handlerJson)
	mux.HandleFunc(a.enpointServerImage, handlerImage)

	log.Println("Слушаем порт:")
	err := http.ListenAndServe(a.portServer, mux)
	if err != nil {
		return fmt.Errorf("не удалось просшулать порт: %w", err)
	}
	return nil
}

func handleWords(w http.ResponseWriter, r *http.Request, path string) {
	log.Println("Получен запрос с хоста: ", r.Host)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	content, err := readFile(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func handleJson(w http.ResponseWriter, r *http.Request, path string) {
	log.Println("Получен запрос с хоста: ", r.Host)
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	content, err := readFile(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s := fmt.Sprintf("Метод: %s,\nПротокол: %s,\nХост: %s\n", r.Method, r.Proto, r.Host)

	data := Response{
		Info:     s,
		FileName: path,
		Content:  string(content),
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		log.Println("Не удалось создать json: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(dataJson)
}

func handlerImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос с хоста: ", r.Host)
	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, "./images/Ветер Северный.png")
}

func readFile(path string) ([]byte, error) {
	err := isExistFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти файл: %w", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	return content, nil
}

func isExistFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о файле: %w", err)
	}
	return nil
}
