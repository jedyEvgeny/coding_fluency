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
	tempDir := t.TempDir()
	content := []string{
		"БИМ бом",
		"диН Дон",
		"кин, ?кон",
		"",
		" ",
		"	",
		"1 !22 -333",
		"apple /banana",
	}
	expectedSlice := []string{"БИМ", "бом", "диН", "Дон", "кин", "кон", "1", "22", "333", "apple", "banana"}
	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(tempDir, fName)
		err := os.WriteFile(fPath, []byte(el), 0744)
		if err != nil {
			log.Fatal(err)
		}
	}
	var allWords []string
	a := App{
		filesDir: tempDir,
	}
	filesList, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range filesList {
		a.wg.Add(1)
		a.findAllWords(&allWords, entry)
	}
	a.wg.Wait()
	if !reflect.DeepEqual(expectedSlice, allWords) {
		t.Errorf("Ожидалось: \n%s,\n получили \n%s\n", expectedSlice, allWords)
	}
}
