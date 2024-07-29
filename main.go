// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу
// ЗЫ - нужно создать файлы с содержимым

// Лучшее время - 12 мин
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
)

var (
	wg sync.WaitGroup
	mu sync.Mutex
)

func main() {
	pathDir := "./files"
	if len(os.Args) == 2 {
		pathDir = os.Args[1]
	}
	filesList, err := os.ReadDir(pathDir)
	if err != nil {
		log.Println(err)
		return
	}
	allWordsSlice := make([]string, 0, 10)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		wg.Add(1)
		func(entry fs.DirEntry) {
			fullPathFile := filepath.Join(pathDir, entry.Name())
			contentFile, err := os.ReadFile(fullPathFile)
			if err != nil {
				log.Println(err)
				return
			}
			words := strings.Fields(string(contentFile))
			mu.Lock()
			allWordsSlice = append(allWordsSlice, words...)
			mu.Unlock()
		}(fileEntry)
	}
	fmt.Println("Общее количество слов:", len(allWordsSlice))
	allWordsMap := make(map[string]int)
	for _, elem := range allWordsSlice {
		allWordsMap[elem]++
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
	var currentFrequency, lastFrequncy, topWords int
	buf := make([]string, 0, 10)
	for _, elem := range frequencyWordsSlice {
		if topWords > 10 {
			break
		}
		currentFrequency = elem.frequency
		if currentFrequency != lastFrequncy && len(buf) > 0 {
			fmt.Printf("\tТоп №%d состоит из %d слов, которые встречаются по %d р.: %s\n", topWords, len(buf), lastFrequncy, buf)
			buf = nil
		}
		if currentFrequency != lastFrequncy {
			lastFrequncy = currentFrequency
			topWords++
		}
		buf = append(buf, elem.word)
	}
	if len(buf) > 0 && topWords < 11 {
		fmt.Printf("\tТоп №%d состоит из %d слов, которые встречаются по %d р.\n", topWords, len(buf), lastFrequncy)
	}
}
		fmt.Printf("\tТоп №%d состоит из %d слов, встречающихся по %d р.\n", topWord, len(buf), lastFrequency)
	}
}

