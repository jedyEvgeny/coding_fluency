package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindAllWords(t *testing.T) {
	testDir := t.TempDir()
	a := &App{FilesDir: testDir}
	content := []string{
		"БИМ бом",
		"дин Дон",
		"кин, !кон",
		"",
		" ",
		"1 /22 333?",
	}
	expectedSlice := []string{"БИМ", "бом", "дин", "Дон", "кин", "кон", "1", "22", "333"}

	for idx, el := range content {
		fName := fmt.Sprintf("%d.txt", idx)
		fPath := filepath.Join(testDir, fName)
		err := os.WriteFile(fPath, []byte(el), 0744)
		if err != nil {
			t.Fatal(err)
		}
	}
	filesList, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}
	allSliceWords := make([]string, 0, 9)
	for _, entry := range filesList {
		a.wg.Add(1)
		a.findAllWords(&allSliceWords, entry)
	}
	a.wg.Wait()
	if !reflect.DeepEqual(expectedSlice, allSliceWords) {
		t.Errorf("Ожидалось %s, получили %s", expectedSlice, allSliceWords)
	}
}
