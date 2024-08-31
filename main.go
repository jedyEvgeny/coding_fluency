// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу

// Лучшее время набора обоих файлов = 52 мин.
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
	filesDir            string
	maxTopWords         int
	wg                  sync.WaitGroup
	mu                  sync.Mutex
	hostClient          string
	methodClient        string
	endpointClient      string
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
	FileName string `json:"Имя найденного файла"`
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
	nameUpChars := strings.ToUpper(*name)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s, введи фразу со знаком $ для старта сервиса:\n", nameUpChars)
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
			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Print("\rСервис запущен!          \n")
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
		schemeClient:        schemeClient,
		endpointClient:      endpointClient,
		hostClient:          hostClient,
		methodClient:        methodClient,
		portServer:          portServer,
		endpointServerJson:  endpointServerJson,
		endpointServerWords: endpointServerWords,
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

	var currFrequecny, lastFrequency, topWords int
	var buf []string
	results := make([]string, 0, a.maxTopWords)
	for _, el := range frequencyWords {
		if topWords > a.maxTopWords {
			break
		}
		currFrequecny = el.frequency
		if currFrequecny != lastFrequency && len(buf) > 0 {
			s := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			fmt.Print(s)
			results = append(results, s)
			buf = nil
		}
		if currFrequecny != lastFrequency {
			lastFrequency = currFrequecny
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
	query.Add("_limit", "4")

	u.RawQuery = query.Encode()
	fmt.Println(u.String())

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
		s := fmt.Sprintf("Имя: %v, телефон: %v, почта: %v\n", el["name"], el["phone"], el["email"])
		fmt.Print(s)
		fmt.Println(strings.Repeat("-", 25))
	}
	return nil
}

func (a *App) createServer(path string) error {
	handlerWords := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, path) }
	handlerJson := func(w http.ResponseWriter, r *http.Request) { handleJson(w, r, path) }

	http.HandleFunc(a.endpointServerWords, handlerWords)
	http.HandleFunc(a.endpointServerJson, handlerJson)

	log.Println("Запустили сервер:")
	err := http.ListenAndServe(a.portServer, nil)
	if err != nil {
		return err
	}
	return nil
}

func handleWords(w http.ResponseWriter, r *http.Request, path string) {
	log.Println("Получен запрос методом:", r.Method)
	w.Header().Set("Conten-Type", "text/plain; charset=utf-8")
	content, err := readFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func handleJson(w http.ResponseWriter, r *http.Request, path string) {
	log.Println("Получен запрос с хоста", r.Host)
	w.Header().Set("Content-Type", "application/json")
	content, err := readFile(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	data := Response{
		Proto:    r.Proto,
		FileName: path,
		Result:   string(content),
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
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
