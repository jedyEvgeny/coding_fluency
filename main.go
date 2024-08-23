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
	mu             sync.Mutex
	wg             sync.WaitGroup
	filesDir       string
	maxTopWords    int
	scheme         string
	basePathClient string
	method         string
	host           string
	portServer     string
	endpointWords  string
	endpointJSON   string
}

type frequencyWord struct {
	word      string
	frequency int
}

var (
	filesDir       = "./files"
	maxTopWords    = 10
	scheme         = "https"
	basePathClient = "/users"
	hostClient     = "jsonplaceholder.typicode.com"
	method         = ""
	portServer     = ":8080"
	endpointWords  = "/words"
	endpointJSON   = "/json"
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
		filesDir:       findFilesDir(),
		maxTopWords:    maxTopWords,
		scheme:         scheme,
		host:           hostClient,
		basePathClient: basePathClient,
		method:         method,
		endpointWords:  endpointWords,
		portServer:     portServer,
		endpointJSON:   endpointJSON,
	}
}

func findFilesDir() string {
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

	a.createRequest()
	fPath := a.saveResult(result)
	a.createServer(fPath)
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

func (a *App) createRequest() {
	u := url.URL{
		Scheme: a.scheme,
		Host:   a.host,
		Path:   path.Join(a.basePathClient, a.method),
	}
	log.Println("URL:", u.String())

	query := url.Values{}
	query.Add("_limit", "3")

	u.RawQuery = query.Encode()
	log.Println("URL c параметрами:", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}

	n := time.Now()
	resp, err := client.Do(req)
	e := time.Now()
	log.Println("Запрос ответа длительностью:", e.Sub(n))
	log.Println(resp.Status)

	if err != nil {
		log.Println(err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var users []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		log.Println(err)
		return
	}

	for _, el := range users {
		s := fmt.Sprintf("Имя: %v\nТелефон: %v\n, e-mail: %v\n", el["name"], el["phone"], el["email"])
		fmt.Print(s)
		fmt.Println(strings.Repeat("-", 25))
	}

}

func (a *App) saveResult(result string) string {
	h := sha1.New()
	_, err := io.WriteString(h, result)
	if err != nil {
		log.Fatal(err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	fPath := filepath.Join(a.filesDir, hash)
	err = os.WriteFile(fPath, []byte(result), 0744)
	if err != nil {
		log.Fatal(err)
	}
	return fPath
}

func (a *App) createServer(path string) {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, path) }
	handlerJSON := func(w http.ResponseWriter, r *http.Request) { handleJSON(w, r, path) }
	http.HandleFunc(a.endpointWords, handlerWords)
	http.HandleFunc(a.endpointJSON, handlerJSON)

	err := http.ListenAndServe(a.portServer, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleWords(w http.ResponseWriter, _ *http.Request, path string) {
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	result, err := readFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func handleJSON(w http.ResponseWriter, _ *http.Request, path string) {
	w.Header().Set("Content-type", "application/json")
	result, err := readFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"Message":    "Успех",
		"Имя файла":  path,
		"Содержимое": string(result),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

func readFile(path string) ([]byte, error) {
	err := isExistFile(path)
	if err != nil {
		return nil, err
	}
	result, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func isExistFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		log.Println("Файл не существует", err)
		return err
	}
	return nil

}

