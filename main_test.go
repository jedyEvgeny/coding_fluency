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
		"кин, !кон",
		"apple+banana",
		"",
		" ",
		"	",
		"1 -22 :333",
	}
	expectedSlice := []string{"БИМ", "бом", "диН", "Дон", "кин", "кон", "apple", "banana", "1", "22", "333"}
	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(tempDir, fName)
		err := os.WriteFile(fPath, []byte(el), perm)
		if err != nil {
			t.Fatal(err)
		}
	}
	var allWords []string
	filesList, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	a := App{
		filesDir: tempDir,
	}
	for _, entry := range filesList {
		a.wg.Add(1)
		allWords = a.findWords(allWords, entry)
	}
	a.wg.Wait()
	if !reflect.DeepEqual(expectedSlice, allWords) {
		t.Errorf("Ожидалось \n%s,\nПолучили \n%s\n", expectedSlice, allWords)
	}
}

//	Запуск: go test -bench=. -v	
//	func BenchmarkFindWords(b *testing.B) {
// 	dir := "./files"

// 	var allWords []string
// 	filesList, err := os.ReadDir(dir)
// 	if err != nil {
// 		b.Fatal(err)
// 	}

// 	a := App{
// 		filesDir: dir,
// 	}

// 	// Запускаем бенчмарк
// 	for i := 0; i < b.N; i++ {
// 		allWords = allWords[:0] // Сбрасываем slice
// 		for _, entry := range filesList {
// 			a.wg.Add(1)
// 			allWords = a.findWords(allWords, entry)
// 		}
// 		a.wg.Wait()
// 	}
// }
