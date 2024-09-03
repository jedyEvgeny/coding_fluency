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
		"Дин доН",
		"кин, ?кон",
		"apple*banana",
		"",
		" ",
		"	",
		"1 -22 333_",
	}
	expectedWords := []string{"БИМ", "бом", "Дин", "доН", "кин", "кон", "apple", "banana", "1", "22", "333"}
	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(tempDir, fName)
		err := os.WriteFile(fPath, []byte(el), perm)
		if err != nil {
			t.Fatal(err)
		}
	}
	var allWords []string
	a := App{filesDir: tempDir}
	filesList, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range filesList {
		allWords = a.findWords(allWords, entry)
	}
	if !reflect.DeepEqual(expectedWords, allWords) {
		t.Errorf("Ожидалось:\n%s,\nполучили:\n%s\n", expectedWords, allWords)
	}
}

func BenchmarkFindWords(b *testing.B) {
	dir := "./files"
	a := App{filesDir: dir}
	filesList, err := os.ReadDir(a.filesDir)
	if err != nil {
		b.Fatal(err)
	}
	var allWords []string

	for i := 0; i < b.N; i++ {
		allWords = allWords[:0]
		for _, entry := range filesList {
			allWords = a.findWords(allWords, entry)
		}
	}
}
