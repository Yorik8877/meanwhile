package ui

import (
	"fmt"
	"os"
	"slices"
)

const (
	systemLanguage   = ""
	russianLanguage  = "ru"
	englishLanguage  = "en"
	languageVariable = "LANGUAGE"
)

// LanguageSetting keeps the chosen interface language. An empty code means
// the system language.
type LanguageSetting interface {
	Language() (string, error)
	SaveLanguage(code string) error
}

type languageOption struct {
	code  string
	label func() string
}

// languageOptions name each language in itself, so that someone who picked a
// language they cannot read can still find their own.
var languageOptions = []languageOption{
	{systemLanguage, message{"language.system", "System language"}.String},
	{russianLanguage, func() string { return "Русский" }},
	{englishLanguage, func() string { return "English" }},
}

// The LANGUAGE value the program was started with; it comes back when the
// user returns to the system language.
var startupLanguageValue, startupLanguageIsSet = os.LookupEnv(languageVariable)

// ApplyLanguage switches the interface language. Fyne reads the LANGUAGE
// environment variable every time translations are loaded, so the variable
// is changed and the translations are loaded again. Unknown codes mean the
// system language.
func ApplyLanguage(code string) error {
	if err := setLanguageVariable(code); err != nil {
		return fmt.Errorf("set %s: %w", languageVariable, err)
	}
	return LoadTranslations()
}

func setLanguageVariable(code string) error {
	if code != systemLanguage && isKnownLanguage(code) {
		return os.Setenv(languageVariable, code)
	}
	if startupLanguageIsSet {
		return os.Setenv(languageVariable, startupLanguageValue)
	}
	return os.Unsetenv(languageVariable)
}

func isKnownLanguage(code string) bool {
	return slices.ContainsFunc(languageOptions, func(option languageOption) bool { return option.code == code })
}
