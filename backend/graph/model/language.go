package model

import (
	"io"

	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pemistahl/lingua-go"
)

type Language lingua.Language

func (l *Language) Scan(value interface{}) error {
	if value == nil {
		*l = Language(lingua.Unknown)
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into Language", value)
	}

	switch strings.ToLower(str) {
	case "english", "en":
		*l = Language(lingua.English)
	case "german", "de":
		*l = Language(lingua.German)
	case "french", "fr":
		*l = Language(lingua.French)
	case "spanish", "es":
		*l = Language(lingua.Spanish)
	case "italian", "it":
		*l = Language(lingua.Italian)
	case "portuguese", "pt":
		*l = Language(lingua.Portuguese)
	case "dutch", "nl":
		*l = Language(lingua.Dutch)
	case "polish", "pl":
		*l = Language(lingua.Polish)
	case "russian", "ru":
		*l = Language(lingua.Russian)
	case "turkish", "tr":
		*l = Language(lingua.Turkish)
	case "chinese", "zh":
		*l = Language(lingua.Chinese)
	case "japanese", "ja":
		*l = Language(lingua.Japanese)
	case "arabic", "ar":
		*l = Language(lingua.Arabic)
	case "unknown", "":
		*l = Language(lingua.Unknown)
	default:
		return fmt.Errorf("unknown language: %s", str)
	}

	return nil
}

func (l Language) Value() (driver.Value, error) {
	lang := lingua.Language(l)
	switch lang {
	case lingua.English:
		return "EN", nil
	case lingua.German:
		return "DE", nil
	case lingua.French:
		return "FR", nil
	case lingua.Spanish:
		return "ES", nil
	case lingua.Italian:
		return "IT", nil
	case lingua.Portuguese:
		return "PT", nil
	case lingua.Dutch:
		return "NL", nil
	case lingua.Polish:
		return "PL", nil
	case lingua.Russian:
		return "RU", nil
	case lingua.Turkish:
		return "TR", nil
	case lingua.Chinese:
		return "ZH", nil
	case lingua.Japanese:
		return "JA", nil
	case lingua.Arabic:
		return "AR", nil
	case lingua.Unknown:
		return "UNKNOWN", nil
	default:
		return "UNKNOWN", nil
	}
}

func (l Language) ToLingua() lingua.Language {
	return lingua.Language(l)
}

func FromLingua(lang lingua.Language) Language {
	return Language(lang)
}

func MarshalLanguage(lang Language) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		langStr := strings.ToUpper(string(lang))
		w.Write([]byte(fmt.Sprintf(`"%s"`, langStr)))
	})
}

func UnmarshalLanguage(v interface{}) (Language, error) {
	switch v := v.(type) {
	case string:
		switch strings.ToLower(v) {
		case "english", "en":
			return FromLingua(lingua.English), nil
		case "german", "de":
			return FromLingua(lingua.German), nil
		case "french", "fr":
			return FromLingua(lingua.French), nil
		case "spanish", "es":
			return FromLingua(lingua.Spanish), nil
		case "italian", "it":
			return FromLingua(lingua.Italian), nil
		case "portuguese", "pt":
			return FromLingua(lingua.Portuguese), nil
		case "dutch", "nl":
			return FromLingua(lingua.Dutch), nil
		case "polish", "pl":
			return FromLingua(lingua.Polish), nil
		case "russian", "ru":
			return FromLingua(lingua.Russian), nil
		case "turkish", "tr":
			return FromLingua(lingua.Turkish), nil
		case "chinese", "zh":
			return FromLingua(lingua.Chinese), nil
		case "japanese", "ja":
			return FromLingua(lingua.Japanese), nil
		case "arabic", "ar":
			return FromLingua(lingua.Arabic), nil
		case "unknown":
			return FromLingua(lingua.Unknown), nil
		default:
			return FromLingua(lingua.Unknown), fmt.Errorf("%s is not a supported language", v)
		}
	case nil:
		return FromLingua(lingua.Unknown), nil
	default:
		return FromLingua(lingua.Unknown), fmt.Errorf("%T is not a valid Language", v)
	}
}

// Language detector instance (lazy initialization)
var detector lingua.LanguageDetector

// InitLanguageDetector initializes the language detector with common European and global languages
func InitLanguageDetector() {
	if detector == nil {
		languages := []lingua.Language{
			lingua.English,
			lingua.German,
			lingua.French,
			lingua.Spanish,
			lingua.Italian,
			lingua.Portuguese,
			lingua.Dutch,
			lingua.Polish,
			lingua.Russian,
			lingua.Turkish,
			lingua.Chinese,
			lingua.Japanese,
			lingua.Arabic,
		}
		detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			Build()
	}
}

// DetectLanguage detects the language of the given text
// Returns the detected Language or Unknown if detection fails
func DetectLanguage(text string) Language {
	InitLanguageDetector()

	if text == "" {
		return FromLingua(lingua.Unknown)
	}

	// Detect language from the text
	if detectedLang, exists := detector.DetectLanguageOf(text); exists {
		return FromLingua(detectedLang)
	}

	return FromLingua(lingua.Unknown)
}
