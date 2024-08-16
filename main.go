// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу
// ЗЫ - нужно заранее создать файлы с содержимым, а также go mod для тестов
// ЗЗЫ - начинаю отсчёт времени с создания файлов .go, котороые создаю через консоль touch main.go 

// Лучшее время: - 26 мин 45 сек, включая удаление предыдущих файлов main и main_test через терминал, запуск приложения в терминале, проверка ответа в браузере
// После четырёх дней отдыха на морях, приехав вечером и уставшим, время набора - 17 мин 42 сек
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
	wg sync.WaitGroup
	mu sync.Mutex
)

func main() {
	filesDir := "./files"
	if len(os.Args) == 2 {
		filesDir = os.Args[1]
	}
	filesList, err := os.ReadDir(filesDir)
	if err != nil {
		log.Fatal(err)
	}
	allWordsSlice := make([]string, 0, 10)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		wg.Add(1)
		go findAllWords(&allWordsSlice, filesDir, fileEntry)
	}
	wg.Wait()
	allWordsMap := make(map[string]int)
	for _, el := range allWordsSlice {
		allWordsMap[el]++
	}
	type frequencyWord struct {
		word      string
		frequency int
	}
	frequencyWordsSlice := make([]frequencyWord, 0, 10)
	for key, val := range allWordsMap {
		frequencyWordsSlice = append(frequencyWordsSlice, frequencyWord{key, val})
	}
	sort.Slice(frequencyWordsSlice, func(i, j int) bool {
		return frequencyWordsSlice[i].frequency > frequencyWordsSlice[j].frequency
	})
	var currentFrequency, lastFrequency, topWords int
	buf := make([]string, 0, 10)
	var result string
	for _, el := range frequencyWordsSlice {
		if topWords > 10 {
			break
		}
		currentFrequency = el.frequency
		if currentFrequency != lastFrequency && len(buf) > 0 {
			str := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			fmt.Print(str)
			result += str
			buf = nil
		}
		if currentFrequency != lastFrequency {
			lastFrequency = currentFrequency
			topWords++
		}
		buf = append(buf, el.word)
	}
	if len(buf) > 0 && topWords < 11 {
		str := fmt.Sprintf("Топ №%d состоит из %d слов, встречающихся по %d р.\n", topWords, len(buf), lastFrequency)
		fmt.Print(str)
		result += str
	}

	createRequest()
	createServer(result)
}

func findAllWords(slice *[]string, path string, entry fs.DirEntry) {
	defer wg.Done()
	fullFName := filepath.Join(path, entry.Name())
	content, err := os.ReadFile(fullFName)
	if err != nil {
		log.Println(err)
		return
	}
	words := strings.FieldsFunc(string(content), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	mu.Lock()
	*slice = append(*slice, words...)
	mu.Unlock()
}

func createRequest() {
	host := "api.telegram.org"
	basePath := "bot" + "1234567890"
	method := "getUpdates"
	u := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   path.Join(basePath, method),
	}
	log.Println(u.String())
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	log.Println(req)

	query := url.Values{}
	query.Add("chat_id", "1234560")
	query.Add("text", "Hello, Telegram!")
	log.Println(query)
	req.URL.RawQuery = query.Encode()
	log.Println(req)
}

func createServer(result string) {
	handler := func(w http.ResponseWriter, r *http.Request) { handWords(w, r, result) }
	http.HandleFunc("/words", handler)
	log.Println("Запускаем сервер:")
	http.ListenAndServe(":8080", nil)
}

func handWords(w http.ResponseWriter, _ *http.Request, result string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, result)
}
