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
