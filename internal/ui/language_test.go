package ui

import "testing"

func TestApplyLanguageSwitchesInterfaceTexts(t *testing.T) {
	t.Cleanup(func() { ApplyLanguage(systemLanguage) })
	addButton := message{"tasks.add", "Add"}

	if err := ApplyLanguage(russianLanguage); err != nil {
		t.Fatal(err)
	}
	if text := addButton.String(); text != "Добавить" {
		t.Errorf("russian text = %q, want Добавить", text)
	}

	if err := ApplyLanguage(englishLanguage); err != nil {
		t.Fatal(err)
	}
	if text := addButton.String(); text != "Add" {
		t.Errorf("english text = %q, want Add", text)
	}
}

func TestUnknownLanguageFallsBackToSystem(t *testing.T) {
	if known := isKnownLanguage("xx"); known {
		t.Error("xx is reported as a known language")
	}
	for _, option := range languageOptions {
		if !isKnownLanguage(option.code) {
			t.Errorf("option %q is not known", option.code)
		}
	}
}
