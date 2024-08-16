package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindAllWords(t *testing.T) {
	tempDir := t.TempDir()
	content := []string{
		"БИМ бом",
		"Дин доН",
		"кин, !кон",
		"",
		" ",
		"1 ,22 333?",
	}
	expextedSlice := []string{"БИМ", "бом", "Дин", "доН", "кин", "кон", "1", "22", "333"}
	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(tempDir, fName)
		err := os.WriteFile(fPath, []byte(el), 0744)
		if err != nil {
			t.Fatal(err)
		}
	}
	filesList, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	allWordsSlice := make([]string, 0, 9)
	for _, entry := range filesList {
		wg.Add(1)
		findAllWords(&allWordsSlice, tempDir, entry)
	}
	wg.Wait()
	if !reflect.DeepEqual(expextedSlice, allWordsSlice) {
		t.Errorf("Ожидалось %s, получили %s\n", expextedSlice, allWordsSlice)
	}
}
