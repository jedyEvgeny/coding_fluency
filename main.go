// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу

// Лучшее время набора обоих файлов = 49 мин.
// ЗЫ - нужно заранее создать файлы с содержимым, а также go mod для тестов
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
	wg                  sync.WaitGroup
	mu                  sync.Mutex
	filesDir            string
	maxTopWords         int
	hostClient          string
	endpointClient      string
	methodClient        string
	schemeClient        string
	portServer          string
	endpointServerJson  string
	endpointServerWords string
}

type frequencyWord struct {
	word      string
	frequency int
}

type Response struct {
	Method string `json:"Метод запроса"`
	Path   string `json:"Имя файла"`
	Msg    string `json:"Содержимое файла"`
}

const perm = 0744

var (
	filesDir            = "./files"
	maxTopWords         = 10
	hostClient          = "jsonplaceholder.typicode.com"
	endpointClient      = "users"
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
	name := flag.String("name", "гость", "ваше имя")
	flag.Parse()
	fmt.Printf("Здравствуй, %s!\n", *name)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Введи фраузу с $ для старта сервиса:")
		if !scanner.Scan() {
			log.Println("ошибка ввода:", scanner.Err())
		}
		phrase := scanner.Text()
		ok := findKey(phrase)
		if ok {
			break
		}
	}
}

func findKey(phrase string) bool {
	target := '$'
	idxTarget := strings.IndexRune(phrase, target)
	return idxTarget > -1
}

func New() App {
	return App{
		filesDir:            findFilesDir(),
		maxTopWords:         maxTopWords,
		hostClient:          hostClient,
		endpointClient:      endpointClient,
		methodClient:        methodClient,
		schemeClient:        schemeClient,
		endpointServerJson:  endpointServerJson,
		endpointServerWords: endpointServerWords,
		portServer:          portServer,
	}
}

func findFilesDir() string {
	if len(os.Args) > 1 {
		filesDir = os.Args[len(os.Args)-1]
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
		go a.findWords(&allWords, fileEntry)
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
	pathFileResult, err := a.saveResult(result)
	if err != nil {
		return err
	}
	err = a.createRequest()
	if err != nil {
		return err
	}
	err = a.createServer(pathFileResult)
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
	err = os.WriteFile(fPath, []byte(result), perm)
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
	fmt.Println(u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	client := &http.Client{}

	start := time.Now()
	resp, err := client.Do(req)
	finish := time.Now()
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	log.Println(finish.Sub(start), resp.Status)

	var data []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	for _, user := range data {
		s := fmt.Sprintf("Имя: %v; телефон: %v; почта: %v.\n", user["name"], user["phone"], user["email"])
		fmt.Print(s)
		fmt.Println(strings.Repeat("*", 25))
	}
	return nil
}

func (a *App) createServer(path string) error {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, path) }
	handlerJson := func(w http.ResponseWriter, r *http.Request) { handleJson(w, r, path) }

	http.HandleFunc(a.endpointServerJson, handlerJson)
	http.HandleFunc(a.endpointServerWords, handlerWords)

	log.Println("Запустили сервер")
	err := http.ListenAndServe(a.portServer, nil)
	if err != nil {
		return err
	}
	return nil
}

func handleJson(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Content-Type", "application/json")
	content, err := readFile(path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := Response{
		Method: r.Method,
		Path:   path,
		Msg:    string(content),
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

func handleWords(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	content, err := readFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println(r.Proto)
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
