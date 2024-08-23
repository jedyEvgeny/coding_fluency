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
	"time"
	"unicode"
)

type App struct {
	wg               sync.WaitGroup
	mu               sync.Mutex
	filesDir         string
	maxTopWords      int
	scheme           string
	host             string
	method           string
	endpointClient   string
	endpointServWord string
	endpointServJson string
	port             string
}

type frequencyWord struct {
	word      string
	frequency int
}

type Response struct {
	Message  string `json:"Сообщение"`
	FileName string `json:"Имя файла"`
	Content  string `json:"Содержимое"`
}

var (
	filesDir         = "./files"
	maxTopWords      = 10
	host             = "jsonplaceholder.typicode.com"
	sheme            = "https"
	method           = ""
	endpointClient   = "users"
	endpointServWord = "/words"
	enpointServJson  = "/json"
	port             = ":8080"
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
		filesDir:         findDir(),
		maxTopWords:      maxTopWords,
		host:             host,
		method:           method,
		endpointClient:   endpointClient,
		scheme:           sheme,
		endpointServWord: endpointServWord,
		endpointServJson: enpointServJson,
		port:             port,
	}
}

func findDir() string {
	if len(os.Args) == 2 {
		filesDir = os.Args[1]
	}
	return filesDir
}

func (a *App) Run() error {
	filesList, err := os.ReadDir(a.filesDir)
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

	var currFrequency, lastFrequency, topWords int
	buf := make([]string, 0, 10)
	var result string

	for _, el := range frequencyWords {
		if topWords > a.maxTopWords {
			break
		}
		currFrequency = el.frequency
		if currFrequency != lastFrequency && len(buf) > 0 {
			s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			result += s
			fmt.Print(s)
			buf = nil
		}
		if currFrequency != lastFrequency {
			lastFrequency = currFrequency
			topWords++
		}
		buf = append(buf, el.word)
	}
	if len(buf) > 0 && topWords < a.maxTopWords+1 {
		s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.\n", topWords, len(buf), lastFrequency)
		result += s
		fmt.Print(s)
	}
	path, err := a.saveResult(result)
	if err != nil {
		return err
	}

	err = a.createRequest()
	if err != nil {
		return err
	}
	err = a.createServer(path)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) findAllWords(slice *[]string, entry fs.DirEntry) {
	defer a.wg.Done()
	fullFPath := filepath.Join(a.filesDir, entry.Name())
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

func (a *App) saveResult(result string) (string, error) {
	h := sha1.New()
	_, err := io.WriteString(h, result)
	if err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	fPath := filepath.Join(a.filesDir, hash)
	err = os.WriteFile(fPath, []byte(result), 0744)
	if err != nil {
		return "", err
	}
	return fPath, nil
}

func (a *App) createRequest() error {
	u := url.URL{
		Scheme: a.scheme,
		Path:   path.Join(a.endpointClient, a.method),
		Host:   host,
	}

	query := url.Values{}
	query.Add("_limit", "3")

	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	log.Println(req)

	client := &http.Client{}

	start := time.Now()
	resp, err := client.Do(req)
	end := time.Now()
	log.Println("Длительность ответа:", end.Sub(start))
	log.Println(resp.StatusCode)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var users []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return err
	}

	for _, el := range users {
		str := fmt.Sprintf("Имя: %v\nТелефон: %v\ne-mail: %v\n", el["name"], el["phone"], el["email"])
		fmt.Print(str)
		fmt.Println(strings.Repeat("-", 25))
	}
	return nil
}

func (a *App) createServer(path string) error {
	handlerWord := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, path) }
	handlerJson := func(w http.ResponseWriter, r *http.Request) { handleJson(w, r, path) }
	http.HandleFunc(a.endpointServWord, handlerWord)
	http.HandleFunc(a.endpointServJson, handlerJson)

	log.Println("Слушаем порт")
	err := http.ListenAndServe(a.port, nil)
	if err != nil {
		return err
	}
	return nil
}

func handleWords(w http.ResponseWriter, _ *http.Request, path string) {
	log.Println("Получен запрос")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	content, err := readContent(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func handleJson(w http.ResponseWriter, _ *http.Request, path string) {
	log.Println("Получен запрос")
	w.Header().Set("Content-Type", "application/json")
	content, err := readContent(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := Response{
		Message:  "Успех",
		FileName: path,
		Content:  string(content),
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func readContent(path string) ([]byte, error) {
	err := isExistFile(path)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
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

