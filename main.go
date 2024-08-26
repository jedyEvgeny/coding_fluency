// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу

// Лучшее время набора обоих файлов = 52 мин
// ЗЫ - нужно заранее создать файлы с содержимым, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с удаления предыдущих файлов .go, котороые следом создаю через консоль touch main.go 
package main

import (
	"bufio"
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
	filesDir            string
	wg                  sync.WaitGroup
	mu                  sync.Mutex
	maxTopWords         int
	endpointClient      string
	hostClient          string
	methodClient        string
	scheme              string
	endpointServerWords string
	endpointServerJson  string
	portServer          string
}

type frequencyWord struct {
	word      string
	frequency int
}

type Response struct {
	Message  string `json:"Сообщение"`
	FileName string `json:"Имя файла"`
	Result   string `json:"Содержимое файла"`
}

const perm = 0744

var (
	filesDir            = "./files"
	maxTopWords         = 10
	endpointClient      = "users"
	methodClient        = ""
	hostClient          = "jsonplaceholder.typicode.com"
	scheme              = "https"
	endpointServerWords = "/words"
	endpointServerJson  = "/json"
	portServer          = ":8080"
)

func main() {
	initInstruction()
	a := new()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func initInstruction() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Введите строку со знаком $ для старта сервиса:")
		if !scanner.Scan() {
			log.Println("Не удалось считать данные:", scanner.Err())
		}
		phrase := scanner.Text()
		ok := findKey(phrase)
		if ok {
			break
		}
	}
}

func findKey(phrase string) bool {
	destChar := '$'
	idx := strings.IndexRune(phrase, destChar)
	return idx > -1
}

func new() App {
	return App{
		filesDir:            findFilesDir(),
		maxTopWords:         maxTopWords,
		hostClient:          hostClient,
		endpointClient:      endpointClient,
		methodClient:        methodClient,
		scheme:              scheme,
		endpointServerWords: endpointServerWords,
		endpointServerJson:  endpointServerJson,
		portServer:          portServer,
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
		go a.findWords(&allWords, fileEntry)
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
	var buf []string
	results := make([]string, 0, a.maxTopWords)
	for _, el := range frequencyWords {
		if topWords > a.maxTopWords {
			break
		}
		currFrequency = el.frequency
		if currFrequency != lastFrequency && len(buf) > 0 {
			s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			results = append(results, s)
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
		results = append(results, s)
		fmt.Print(s)
	}
	result := strings.Join(results, "")
	fPathRes, err := a.saveResult(result)
	if err != nil {
		return err
	}
	_ = fPathRes
	err = a.createRequest()
	if err != nil {
		return err
	}
	err = a.createServer(fPathRes)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) findWords(slice *[]string, entry fs.DirEntry) {
	defer a.wg.Done()
	fPath := filepath.Join(a.filesDir, entry.Name())
	content, err := os.ReadFile(fPath)
	if err != nil {
		log.Println(err)
		return
	}
	words := strings.FieldsFunc(string(content), func(r rune) bool {
		return !unicode.IsNumber(r) && !unicode.IsLetter(r)
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
	err = os.WriteFile(fPath, []byte(result), perm)
	if err != nil {
		return "", err
	}
	return fPath, nil
}

func (a *App) createRequest() error {
	u := url.URL{
		Scheme: a.scheme,
		Path:   path.Join(a.endpointClient, a.methodClient),
		Host:   a.hostClient,
	}

	query := url.Values{}
	query.Add("_limit", "3")
	u.RawQuery = query.Encode()
	log.Println(u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	client := &http.Client{}

	start := time.Now()
	resp, err := client.Do(req)
	end := time.Now()
	if err != nil {
		return err
	}
	log.Println(resp.Status, end.Sub(start))

	var data []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}
	for _, el := range data {
		s := fmt.Sprintf("Имя: %v; телефон: %v; почта: %v\n", el["name"], el["phone"], el["email"])
		fmt.Print(s)
		fmt.Println(strings.Repeat("-", 30))
	}

	return nil
}

func (a *App) createServer(path string) error {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, path) }
	handlerJson := func(w http.ResponseWriter, r *http.Request) { handleJson(w, path) }

	http.HandleFunc(a.endpointServerWords, handlerWords)
	http.HandleFunc(a.endpointServerJson, handlerJson)

	log.Println("Начинаем слушать порт:")
	err := http.ListenAndServe(a.portServer, nil)
	if err != nil {
		return err
	}
	return nil
}

func handleJson(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/json")
	content, err := readFile(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := Response{
		Message:  "Успех",
		FileName: path,
		Result:   string(content),
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

func handleWords(w http.ResponseWriter, path string) {
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

func readFile(path string) ([]byte, error) {
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
