// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу

// Лучшее время набора - 49 мин 07 сек, в т.ч. тесты, создание файлов и т.д. - но импорты вручную не прописываю.
// ЗЫ - нужно заранее создать файлы с содержимым, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с создания файлов .go, котороые создаю через консоль touch main.go 

package main

import (
	"encoding/json"
	"fmt"
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
	"unicode"
)

type App struct {
	wg            sync.WaitGroup
	mu            sync.Mutex
	baseFailsDir  string
	needsTopWords int
	host          string
	method        string
	basePath      string
	scheme        string
	endpointWords string
	endpointJSON  string
	port          string
}

type frequencyWord struct {
	word      string
	frequency int
}

var (
	filesDir      = "./files"
	needsTopWords = 10
	host          = "api.telegram.org"
	method        = "getUpdates"
	scheme        = "https"
	endpointWords = "/words"
	endpointJSON  = "/json"
	port          = ":8080"
)

func main() {
	a := New()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func New() App {
	return App{
		baseFailsDir:  baseFailsDir(),
		needsTopWords: needsTopWords,
		host:          host,
		method:        method,
		basePath:      basePath(),
		scheme:        scheme,
		endpointWords: endpointWords,
		endpointJSON:  endpointJSON,
		port:          port,
	}
}

func baseFailsDir() string {
	if len(os.Args) == 2 {
		filesDir = os.Args[1]
	}
	return filesDir
}

func basePath() string {
	return path.Join("bot" + "1234567890")
}

func (a *App) Run() error {
	filesList, err := os.ReadDir(a.baseFailsDir)
	if err != nil {
		return err
	}
	allWords := make([]string, 0, 10)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		a.wg.Add(1)
		go a.findAllWords(&allWords, fileEntry)
	}
	a.wg.Wait()

	uniqueWords := make(map[string]int)
	for _, el := range allWords {
		uniqueWords[el]++
	}

	frequencyWords := make([]frequencyWord, 0, 10)
	for key, val := range uniqueWords {
		frequencyWords = append(frequencyWords, frequencyWord{key, val})
	}

	sort.Slice(frequencyWords, func(i, j int) bool {
		return frequencyWords[i].frequency > frequencyWords[j].frequency
	})

	var currentFrequency, lastFrequency, topWords int
	buf := make([]string, 0, 10)
	var result string

	for _, el := range frequencyWords {
		if topWords > a.needsTopWords {
			break
		}
		currentFrequency = el.frequency
		if currentFrequency != lastFrequency && len(buf) > 0 {
			str := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			result += str
			fmt.Print(str)
			buf = nil
		}
		if currentFrequency != lastFrequency {
			topWords++
			lastFrequency = currentFrequency
		}
		buf = append(buf, el.word)
	}
	if len(buf) > 0 && topWords < a.needsTopWords+1 {
		str := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.\n", topWords, len(buf), lastFrequency)
		result += str
		fmt.Print(str)
	}

	a.createRequest()
	a.startServer(result)
	return nil
}

func (a *App) findAllWords(slice *[]string, entry fs.DirEntry) {
	defer a.wg.Done()
	fullFPath := filepath.Join(a.baseFailsDir, entry.Name())
	content, err := os.ReadFile(fullFPath)
	if err != nil {
		log.Println(err)
		return
	}
	words := strings.FieldsFunc(string(content), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	a.mu.Lock()
	*slice = append(*slice, words...)
	a.mu.Unlock()
}

func (a *App) createRequest() {
	u := url.URL{
		Scheme: a.scheme,
		Path:   path.Join(a.method, a.basePath),
		Host:   a.host,
	}
	log.Println(u.String())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Println(err)
	}
	log.Println(req)

	query := url.Values{}
	query.Add("chat_id", "1234560")
	query.Add("text", "Hello, Telegram!")
	log.Println(query)

	req.URL.RawQuery = query.Encode()
	log.Println(req)
}

func (a *App) startServer(result string) {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handlWords(w, r, result) }
	handlerJSON := func(w http.ResponseWriter, r *http.Request) { handleJSON(w, r, result) }
	http.HandleFunc(a.endpointWords, handlerWords)
	http.HandleFunc(a.endpointJSON, handlerJSON)

	log.Println("Начали слушать порт:")
	err := http.ListenAndServe(a.port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handlWords(w http.ResponseWriter, _ *http.Request, result string) {
	log.Println("Получен запрос")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, result)
}

func handleJSON(w http.ResponseWriter, _ *http.Request, result string) {
	log.Println("Получен запрос")
	w.Header().Set("Content-Type", "application/json")

	data := map[string]string{
		"message": "Успех!",
		"result":  result,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Не смогли преобразовать объект в json")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
