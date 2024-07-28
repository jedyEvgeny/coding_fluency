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
	var top, currentFrequency, lastFrecuency int
	var buf []string
	for _, elem := range frequencyWordsSlice {
		if top > 10 {
			break
		}
		currentFrequency = elem.frequency
		if len(buf) > 0 && currentFrequency != lastFrecuency {
			fmt.Printf("Топ %d частоты слов встречается по %d р.:\t%v\n", top, lastFrecuency, buf)
			buf = nil
		}
		if currentFrequency != lastFrecuency {
			top++
			lastFrecuency = currentFrequency
		}
		buf = append(buf, elem.word)
	}
}
