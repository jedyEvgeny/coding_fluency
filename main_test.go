package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindAllWord(t *testing.T) {
	testDir := t.TempDir()
	content := []string{
		"бим бом дин дон",
		"КИН КОН",
		"1 2 3 ",
		"",
		" ",
		".а и! дореми?",
	}
	expectWords := []string{"бим", "бом", "дин", "дон", "КИН", "КОН", "1", "2", "3", "а", "и", "дореми"}
	for idx, el := range content {
		fileName := fmt.Sprintf("%d.txt", idx)
		filePath := filepath.Join(testDir, fileName)
		err := os.WriteFile(filePath, []byte(el), 0666)
		if err != nil {
			t.Fatal(err)
		}
	}
	filesList, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}
	var allWordsSlice []string
	for _, el := range filesList {
		wg.Add(1)
		findAllWords(&allWordsSlice, testDir, el)
	}
	wg.Wait()
	if !reflect.DeepEqual(expectWords, allWordsSlice) {
		t.Errorf("ожидалось %s, получили %s", expectWords, allWordsSlice)
	}
}
