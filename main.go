// Файл создан для укрепления умения беглого программирования и будет включать основные инструменты Go
// Этот файл будет постепенно пополняться до 1 часа непрерывного коддинга при умении скоростной печати
// Чтобы получить пользу от файла, вы должны начать кодить его содержимое фрагментами и усваивать что и как тут происходит
// Первый блок кода читает файлы и выводит топ-10 слов
// Запускаем с аргументом, примерно так: go run main.go /.short_files
// Символ ./ используется в bash-языке как символ относительного пути к текущему каталогу
// ЗЫ - нужно создать файлы с содержимым

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type frequencyWord struct {
	word      string
	frequency int
}

func main() {
	pathDirFiles := "./files"
	if len(os.Args) == 2 {
		pathDirFiles = os.Args[1]
	}
	filesList, err := os.ReadDir(pathDirFiles)
	if err != nil {
		log.Println("не удаётся найти каталог:", err)
		return
	}

	allWordsSlice := make([]string, 0)
	for _, fileEntry := range filesList {
		if fileEntry.IsDir() {
			continue
		}
		fullFilePath := filepath.Join(pathDirFiles, fileEntry.Name())
		contentFile, err := os.ReadFile(fullFilePath)
		if err != nil {
			log.Printf("не удалось прочитать файл %s: %s", fileEntry.Name(), err)
			continue
		}
		words := strings.Fields(string(contentFile))
		allWordsSlice = append(allWordsSlice, words...)
	}

	allWordsMap := make(map[string]int)
	for _, value := range allWordsSlice {
		allWordsMap[value]++
	}

	frequencyWordsSlice := make([]frequencyWord, 0)
	for key, value := range allWordsMap {
		frequencyWordsSlice = append(frequencyWordsSlice, frequencyWord{key, value})
	}

	sort.Slice(frequencyWordsSlice, func(i, j int) bool {
		return frequencyWordsSlice[i].frequency > frequencyWordsSlice[j].frequency
	})
	var lastFrequency, currentFrequency int
	var top, count int
	for _, elem := range frequencyWordsSlice {
		currentFrequency = elem.frequency
		if lastFrequency != currentFrequency && top != 0 {
			fmt.Println()
		}
		if currentFrequency == 1 {
			fmt.Printf("\tТоп %d: %d слов встречаются по %d разу", top+1, len(frequencyWordsSlice)-count, elem.frequency)
			break
		}
		if lastFrequency != currentFrequency {
			top++
			fmt.Printf("\tТоп %d слов встречается %d раз:", top, elem.frequency)
			lastFrequency = currentFrequency
		}
		fmt.Print(" ", elem.word)
		count++
	}
}
