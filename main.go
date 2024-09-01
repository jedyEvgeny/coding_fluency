// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Основа кода - поиск наиболее часто встречающихся слов среди нескольких заранее подготовленных текстов
// Запускаем с аргументом, примерно так: go run main.go -name Алиса ./short_files

// Лучшее время набора обоих файлов = 58 мин.
// ЗЫ - нужно заранее создать файлы с текстом, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с удаления предыдущих файлов .go, котороые следом создаю через консоль touch main.go 
package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"flag"
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
	"time"
	"unicode"
)

type App struct {
	filesDir            string
	wg                  sync.WaitGroup
	mu                  sync.Mutex
	maxTopWords         int
	hostClient          string
	endpointClient      string
	methodClient        string
	schemeClient        string
	portServer          string
	endpointServerWords string
	endpointServerJson  string
}

type frequencyWord struct {
	word      string
	frequency int
}

type Response struct {
	Proto    string `json:"Протокол запроса клиента"`
	FileName string `json:"Имя файла"`
	Result   string `json:"Содержимое файла"`
}

const perm = 0744

var (
	filesDir            = "./files"
	maxTopWords         = 10
	hostClient          = "jsonplaceholder.typicode.com"
	endpointClient      = "/users"
	methodClient        = ""
	schemeClient        = "https"
	portServer          = ":8080"
	endpointServerWords = "/words"
	endpointServerJson  = "/json"
)

func main() {
	initConfig()
	a := New()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	name := flag.String("name", "гость", "имя пользователя")
	flag.Parse()
	nameUpLetter := strings.ToUpper(*name)
	date := time.Now().Format("02-01-2006")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s, сегодня, %s, введите фразу с $ для старта сервиса:", nameUpLetter, date)
		if !scanner.Scan() {
			log.Println("Ошибка чтения:", scanner.Err())
		}
		phrase := scanner.Text()
		ok := findKey(phrase)
		if ok {
			break
		}
	}
	symbols := []string{"/", "|", "-", "\\"}
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		for _, el := range symbols {
			fmt.Print("\rЗагрузка сервиса:" + el)
			time.Sleep(150 * time.Millisecond)
		}
	}
	fmt.Println("\rСервис запущен!     ")
	time.Sleep(1 * time.Second)
}

func findKey(phrase string) bool {
	target := "$"
	return strings.Contains(phrase, target)
}

func New() App {
	return App{
		filesDir:            findFilesDir(),
		maxTopWords:         maxTopWords,
		hostClient:          hostClient,
		endpointClient:      endpointClient,
		methodClient:        methodClient,
		schemeClient:        schemeClient,
		portServer:          portServer,
		endpointServerWords: endpointServerWords,
		endpointServerJson:  endpointServerJson,
	}
}

func findFilesDir() string {
	l := len(os.Args)
	if l > 1 {
		filesDir = os.Args[l-1]
	}
	return filesDir
}

func (a *App) Run() error {
	filesList, err := os.ReadDir(a.filesDir)
	if err != nil {
		return err
	}
	allWords := make([]string, 0, a.maxTopWords)
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
	a.wg.Wait()
	uniqueWords := make(map[string]int)
	for _, el := range allWords {
		uniqueWords[el]++
	}
	frequencyWords := make([]frequencyWord, 0, a.maxTopWords)
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
			fmt.Print(s)
			results = append(results, s)
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
		fmt.Print(s)
		results = append(results, s)
	}
	result := strings.Join(results, "")
	fPathRes, err := a.saveResult(result)
	if err != nil {
		return err
	}
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

func (a *App) findWords(slice []string, entry fs.DirEntry) []string {
	fPath := filepath.Join(a.filesDir, entry.Name())
	content, err := os.ReadFile(fPath)
	if err != nil {
		log.Println(err)
		return slice
	}
	words := strings.FieldsFunc(string(content), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	a.mu.Lock()
	slice = append(slice, words...)
	a.mu.Unlock()
	return slice
}

func (a *App) saveResult(res string) (string, error) {
	h := sha1.New()
	h.Write([]byte(res))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	fPath := filepath.Join(a.filesDir, hash)
	err := os.WriteFile(fPath, []byte(res), perm)
	if err != nil {
		return "", err
	}
	return fPath, nil
}

func (a *App) createRequest() error {
	u := url.URL{
		Scheme: a.schemeClient,
		Host:   a.hostClient,
		Path:   path.Join(a.endpointClient, a.methodClient),
	}

	query := url.Values{}
	query.Add("_limit", "5")

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
	defer func() { _ = resp.Body.Close() }()
	log.Println(resp.Status, end.Sub(start))

	var data []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	for _, el := range data {
		s := fmt.Sprintf("Имя: %v; телефон: %v; почта: %v\n", el["name"], el["phone"], el["email"])
		fmt.Print(s)
		fmt.Println(strings.Repeat("*", 25))
	}
	return nil
}

func (a *App) createServer(path string) error {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, path) }
	handlerJson := func(w http.ResponseWriter, r *http.Request) { handleJson(w, r, path) }

	http.HandleFunc(a.endpointServerWords, handlerWords)
	http.HandleFunc(a.endpointServerJson, handlerJson)

	log.Println("Запускаем сервер:")
	err := http.ListenAndServe(a.portServer, nil)
	if err != nil {
		return err
	}
	return nil
}

func handleWords(w http.ResponseWriter, r *http.Request, path string) {
	log.Println("Получен запрос методом:", r.Method)
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
	log.Println("Получен запрос с хоста:", r.Host)
	w.Header().Set("Content-Type", "application/json")
	content, err := readFile(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := Response{
		Proto:    r.Proto,
		FileName: path,
		Result:   string(content),
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(dataJson)
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
