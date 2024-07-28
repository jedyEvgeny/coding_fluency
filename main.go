// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go ./short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу
// ЗЫ - нужно создать файлы с содержимым

// Лучшее время - 9 мин
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
		fullPathFile := filepath.Join(pathDir, fileEntry.Name())
		contentFile, err := os.ReadFile(fullPathFile)
		if err != nil {
			log.Printf("содержимое файла %s не удалось прочитать: %s", fileEntry.Name(), err)
			continue
		}
		words := strings.Fields(string(contentFile))
		allWordsSlice = append(allWordsSlice, words...)
	}
	allWordsMap := make(map[string]int)
	for _, val := range allWordsSlice {
		allWordsMap[val]++
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
	var currentFrequency, lastFrequency, topWord int
	var buf []string
	for _, elem := range frequencyWordsSlice {
		if topWord > 10 {
			break
		}
		currentFrequency = elem.frequency
		if currentFrequency != lastFrequency && len(buf) > 0 {
			fmt.Printf("\tТоп №%d состоит из %d слов, которые встречаются по %d р.:\t%s\n", topWord, len(buf), lastFrequency, buf)
			buf = nil
		}
		if currentFrequency != lastFrequency {
			lastFrequency = currentFrequency
			topWord++
		}
		buf = append(buf, elem.word)
	}
	if len(buf) > 0 && topWord < 11 {
		fmt.Printf("\tТоп №%d состоит из %d слов, которые встречаются по %d р.\n", topWord, len(buf), lastFrequency)
	}
}
