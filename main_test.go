package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindAllWords(t *testing.T) {
	testDir := t.TempDir()
	log.Println(testDir)

	// Создание фиктивных тестовых файлов с заданным содержимым
	scentencesSlice := []string{
		"apple banana orange banana", //базовая строка
		"apple 1 ,/banana 1cherry",   //с цифрой и знаками
		"jkl lkl fpfij lkjdf 30",     //с числом
		"",                           //с пустой строкой
		" ",                          //с пробелом
		"озеро\nлуг",                 //с символом переноса строки
	}
	expectedWords := []string{"apple", "banana", "orange", "banana", "apple", "1", "banana", "1cherry",
		"jkl", "lkl", "fpfij", "lkjdf", "30", "озеро", "луг",
	}

	for idx, el := range scentencesSlice {
		fileName := fmt.Sprintf("file%d.txt", idx+1)
		filePath := filepath.Join(testDir, fileName)
		err := os.WriteFile(filePath, []byte(el), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	filesList, _ := os.ReadDir(testDir)
	log.Println(filesList)

	var allWordsSlice []string
	for _, el := range filesList {
		wg.Add(1)
		findAllWords(testDir, &allWordsSlice, el)
	}
	wg.Wait()

	if len(expectedWords) != len(allWordsSlice) {
		t.Errorf("\tдлина среза %d, ожидалась длина %d", len(allWordsSlice), len(expectedWords))
	}
	log.Println(allWordsSlice)
	if !reflect.DeepEqual(allWordsSlice, expectedWords) {
		t.Errorf("\tрезультаты неверны. Ожидалось %s, получено %s", expectedWords, allWordsSlice)
	}
}
