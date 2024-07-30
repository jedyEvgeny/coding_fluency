// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу
// ЗЫ - нужно создать файлы с содержимым

// Лучшее время - 11 мин 33 сек
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"
)

var (
	mu sync.RWMutex
	wg sync.WaitGroup
)

func main() {
	pathDir := "./files"
	if len(os.Args) == 2 {
		pathDir = os.Args[1]
	}
	filesList, err := os.ReadDir(pathDir)
	if err != nil {
		log.Fatal(err)
	}
	allWordsSlice := make([]string, 0, 10)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		wg.Add(1)
		go func(entry fs.DirEntry) {
			defer wg.Done()
			fullFilePath := filepath.Join(pathDir, entry.Name())
			contentFile, err := os.ReadFile(fullFilePath)
			if err != nil {
				log.Println(err)
				return
			}
			words := strings.FieldsFunc(string(contentFile), func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			mu.Lock()
			allWordsSlice = append(allWordsSlice, words...)
			mu.Unlock()
		}(fileEntry)
	}
	wg.Wait()
	allWordsMap := make(map[string]int)
	for _, el := range allWordsSlice {
		allWordsMap[el]++
	}
	type frequncyWord struct {
		word      string
		frequency int
	}
	frequencyWordsSlice := make([]frequncyWord, 0, 10)
	for key, val := range allWordsMap {
		frequencyWordsSlice = append(frequencyWordsSlice, frequncyWord{key, val})
	}
	sort.Slice(frequencyWordsSlice, func(i, j int) bool {
		return frequencyWordsSlice[i].frequency > frequencyWordsSlice[j].frequency
	})
	var currentFrequency, lastFrequency, topWords int
	buf := make([]string, 0, 10)
	for _, el := range frequencyWordsSlice {
		if topWords > 10 {
			break
		}
		currentFrequency = el.frequency
		if currentFrequency != lastFrequency && len(buf) > 0 {
			fmt.Printf("Топ №%d состоит из %d слов, которые встречаются %d р.: %s\n", topWords, len(buf), lastFrequency, buf)
			buf = nil
		}
		if currentFrequency != lastFrequency {
			lastFrequency = currentFrequency
			topWords++
		}
		buf = append(buf, el.word)
	}
	if len(buf) > 0 && topWords < 11 {
		fmt.Printf("Топ №%d состоит из %d слов, которые встречаются %d р.\n", topWords, len(buf), lastFrequency)
	}
}

