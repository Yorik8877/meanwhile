package ui

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func readTranslationKeys(t *testing.T, fileName string) []string {
	t.Helper()
	content, err := translations.ReadFile("translations/" + fileName)
	if err != nil {
		t.Fatalf("read %s: %v", fileName, err)
	}
	var entries map[string]string
	if err := json.Unmarshal(content, &entries); err != nil {
		t.Fatalf("parse %s: %v", fileName, err)
	}
	return slices.Sorted(maps.Keys(entries))
}

func TestTranslationsLoad(t *testing.T) {
	if err := LoadTranslations(); err != nil {
		t.Fatalf("load translations: %v", err)
	}
}

func TestEnglishAndRussianHaveSameKeys(t *testing.T) {
	english := readTranslationKeys(t, "en.json")
	russian := readTranslationKeys(t, "ru.json")

	if !slices.Equal(english, russian) {
		t.Errorf("keys differ:\nen: %v\nru: %v", english, russian)
	}
}

var messageKeyPattern = regexp.MustCompile(`message\{"([a-z_.]+)"`)

func TestEveryKeyUsedInCodeIsTranslated(t *testing.T) {
	known := readTranslationKeys(t, "en.json")
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sources {
		if strings.HasSuffix(source, "_test.go") {
			continue
		}
		content, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range messageKeyPattern.FindAllStringSubmatch(string(content), -1) {
			if _, found := slices.BinarySearch(known, match[1]); !found {
				t.Errorf("%s uses key %q that is missing from en.json", source, match[1])
			}
		}
	}
}
