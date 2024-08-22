// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу

// Лучшее время набора обоих файлов = 55 мин
// ЗЫ - нужно заранее создать файлы с содержимым, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с удаления предыдущих файлов .go, котороые следом создаю через консоль touch main.go 

package main

import (
	"crypto/sha1"
	"encoding/json"
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
	"unicode"
)

type App struct {
	wg            sync.WaitGroup
	mu            sync.Mutex
	filesPath     string
	maxTopWords   int
	host          string
	method        string
	basePathURL   string
	schema        string
	port          string
	endpointWords string
	endpointJSON  string
}

type frequencyWord struct {
	word      string
	frequency int
}

var (
	filesPath     = "./files"
	maxTopWords   = 10
	schema        = "https"
	method        = "getUpdates"
	host          = "api.telegram.org"
	port          = ":8080"
	endpointWords = "/words"
	endpointJSON  = "/json"
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
		filesPath:     findFilesPath(),
		maxTopWords:   maxTopWords,
		schema:        schema,
		host:          host,
		basePathURL:   basePath(),
		method:        method,
		port:          port,
		endpointWords: endpointWords,
		endpointJSON:  endpointJSON,
	}
}

func findFilesPath() string {
	if len(os.Args) > 1 {
		filesPath = os.Args[1]
	}
	return filesPath
}

func basePath() string {
	return "bot" + "1234567890"
}

func (a *App) Run() error {
	filesList, err := os.ReadDir(a.filesPath)
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
		if topWords > a.maxTopWords {
			break
		}
		currentFrequency = el.frequency
		if currentFrequency != lastFrequency && len(buf) > 0 {
			s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			fmt.Print(s)
			result += s
			buf = nil
		}
		if currentFrequency != lastFrequency {
			lastFrequency = currentFrequency
			topWords++
		}
		buf = append(buf, el.word)
	}
	if len(buf) > 0 && topWords < a.maxTopWords+1 {
		s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.\n", topWords, len(buf), lastFrequency)
		fmt.Print(s)
		result += s
	}
	fPath, _ := a.saveResult(result)
	a.createRequest()
	a.createServer(fPath)
	return nil
}

func (a *App) findAllWords(slice *[]string, entry fs.DirEntry) {
	defer a.wg.Done()
	fullFPath := filepath.Join(a.filesPath, entry.Name())
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
		Scheme: a.schema,
		Host:   a.host,
		Path:   path.Join(a.basePathURL, a.method),
	}
	log.Println("Параметры = ", u.String())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Запрос =", req)

	query := url.Values{}
	query.Add("chat_id", "1234567")
	query.Add("text", "Hello, Telegram!")
	log.Println("параметры запроса =", query)

	req.URL.RawQuery = query.Encode()
	log.Println("Полныйй запрос =", req)
}

func (a *App) saveResult(result string) (string, error) {
	h := sha1.New()
	_, err := io.WriteString(h, result)
	if err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	fPath := filepath.Join(a.filesPath, hash)
	err = os.WriteFile(fPath, []byte(result), 0744)
	if err != nil {
		return "", err
	}
	return fPath, err
}

func (a *App) createServer(fPath string) {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, fPath) }
	handlerJSON := func(w http.ResponseWriter, r *http.Request) { handleJSON(w, r, fPath) }
	http.HandleFunc(a.endpointWords, handlerWords)
	http.HandleFunc(a.endpointJSON, handlerJSON)

	log.Println("Начали слушать порт")
	err := http.ListenAndServe(a.port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleWords(w http.ResponseWriter, _ *http.Request, fPath string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	content, err := rFile(fPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func handleJSON(w http.ResponseWriter, _ *http.Request, fPath string) {
	w.Header().Set("Content-Type", "application/json")
	content, err := rFile(fPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := map[string]string{
		"msg":      "Успех!",
		"fileName": fPath,
		"content":  string(content),
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(dataJSON)
}

func rFile(path string) ([]byte, error) {
	err := isExistFile(path)
	if err != nil {
		log.Println("Не смогли открыть файл", err)
		return []byte{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

func isExistFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return nil
}
