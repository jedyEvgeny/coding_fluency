// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу
// ЗЫ - нужно заранее создать файлы с содержимым, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с создания файлов .go, котороые создаю через консоль touch main.go 

// Лучшее время: - 41 мин, включая удаление предыдущих файлов main и main_test через терминал, 
// запуск приложения в терминале, проверка ответа в браузере, запуск тестов. Импорты подтягиваю автоматически

package main

import (
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

var (
	filesDir = "./files"
	endpoint = "/words"
	port     = ":8080"
	topWords = 10
	host     = "api.telegram.org"
	method   = "getUpdates"
	scheme   = "https"
)

type App struct {
	FilesDir    string
	Endpoint    string
	Port        string
	Host        string
	Method      string
	Scheme      string
	mu          sync.Mutex
	wg          sync.WaitGroup
	maxTopWords int
}

func main() {
	a := New()
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func New() App {
	return App{
		FilesDir:    dir(),
		Port:        port,
		Endpoint:    endpoint,
		maxTopWords: topWords,
		Host:        host,
		Method:      method,
		Scheme:      scheme,
	}
}

func dir() string {
	if len(os.Args) == 2 {
		filesDir = os.Args[1]
	}
	return filesDir
}

func (a *App) Run() error {
	filesList, err := os.ReadDir(a.FilesDir)
	if err != nil {
		return err
	}
	allWordsSlice := make([]string, 0, 10)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		a.wg.Add(1)
		go a.findAllWords(&allWordsSlice, fileEntry)
	}
	a.wg.Wait()

	allWordsMap := make(map[string]int)
	for _, el := range allWordsSlice {
		allWordsMap[el]++
	}
	type frequencyWord struct {
		frequency int
		word      string
	}
	frequencyWordsSlice := make([]frequencyWord, 0, 10)
	for key, val := range allWordsMap {
		frequencyWordsSlice = append(frequencyWordsSlice, frequencyWord{val, key})
	}
	sort.Slice(frequencyWordsSlice, func(i, j int) bool {
		return frequencyWordsSlice[i].frequency > frequencyWordsSlice[j].frequency
	})

	var currentFrequency, lastFrequency, topWords int
	buf := make([]string, 0, 10)
	result := ""

	for _, el := range frequencyWordsSlice {
		if topWords > a.maxTopWords {
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
	if len(buf) > 0 && topWords < a.maxTopWords+1 {
		str := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.\n", topWords, len(buf), lastFrequency)
		result += str
		fmt.Print(str)
	}

	a.createRequest()
	a.createServer(result)
	return nil
}

func (a *App) findAllWords(slice *[]string, entry fs.DirEntry) {
	defer a.wg.Done()
	fullFPath := filepath.Join(a.FilesDir, entry.Name())
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
		Scheme: a.Scheme,
		Path:   a.basePath(),
		Host:   a.Host,
	}
	log.Println(u.String())
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	log.Println(req)

	query := url.Values{}
	query.Add("chat_id", "12345670")
	query.Add("text", "Hello, Telegram")
	log.Println(query)

	req.URL.RawQuery = query.Encode()
	log.Println(req)
}

func (a *App) basePath() string {
	basePath := "bot" + "123456"
	return path.Join(a.Method, basePath)
}

func (a *App) createServer(result string) {
	handler := func(w http.ResponseWriter, r *http.Request) { handleWords(w, r, result) }
	log.Println("Начинаем слушать сервер:")
	http.HandleFunc(a.Endpoint, handler)
	http.ListenAndServe(a.Port, nil)
}

func handleWords(w http.ResponseWriter, _ *http.Request, result string) {
	w.Header().Set("Content-type", "text/plaing; charset=utf-8")
	fmt.Fprint(w, result)
}
