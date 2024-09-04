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
		",кин ?кон",
		"apple+banana",
		"",
		" ",
		"	",
		"1 -22 _333%",
	}
	expectedWords := []string{"БИМ", "бом", "диН", "Дон", "кин", "кон", "apple", "banana", "1", "22", "333"}
	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(tempDir, fName)
		err := os.WriteFile(fPath, []byte(el), perm)
		if err != nil {
			t.Fatalf("не удалось создать тестовый файл №%d: %v", idx, err)
		}
	}
	var allWords []string
	a := App{filesDir: tempDir}
	filesList, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("не удалось прочитать директорию %s: %v", tempDir, err)
	}
	for _, entry := range filesList {
		allWords = a.findWords(allWords, entry)
	}
	if !reflect.DeepEqual(expectedWords, allWords) {
		t.Errorf("Ожидалось:\n%s,\nПолучили:\n%s\n", expectedWords, allWords)
	}
}

func Benchmark(b *testing.B) {
	filesDir := "./files"
	a := App{filesDir: filesDir}
	var allWords []string
	filesList, err := os.ReadDir(a.filesDir)
	if err != nil {
		b.Fatalf("не удалось прочитать директорию %s: %v", a.filesDir, err)
	}

	for i := 0; i < b.N; i++ {
		allWords = allWords[:0]
		for _, entry := range filesList {
			allWords = a.findWords(allWords, entry)
		}
	}
}
