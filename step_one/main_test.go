//Тест запускается компндой go test -v или go test
//Бенчмарк запускать командой go test -bench .
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindWords(t *testing.T) {
	tempDir := t.TempDir()
	content := []string{
		"БИМ бом",
		"диН Дон",
		"кин, ??кон",
		"apple+banana",
		"",
		" ",
		"	",
		"1 /22 )333}",
	}
	expectedWords := []string{"БИМ", "бом", "диН", "Дон", "кин", "кон", "apple", "banana", "1", "22", "333"}
	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(tempDir, fName)
		err := os.WriteFile(fPath, []byte(el), perm)
		if err != nil {
			t.Fatal(err)
		}
	}
	filesList, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	a := App{filesDir: tempDir}
	var allWords []string
	for _, entry := range filesList {
		allWords = a.findWords(allWords, entry)
	}
	if !reflect.DeepEqual(expectedWords, allWords) {
		t.Errorf("Ожидалось: \n%s,\nПолучили: \n%s\n", expectedWords, allWords)
	}
}

func BenchmarkFindWords(b *testing.B) {
	dir := "./files"
	a := App{filesDir: dir}
	var allWords []string
	filesList, err := os.ReadDir(a.filesDir)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		allWords = allWords[:0]
		for _, entry := range filesList {
			allWords = a.findWords(allWords, entry)
		}
	}
}
