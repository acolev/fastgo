package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
)

type Params map[string]any

var (
	translations = map[string]map[string]string{}
	defaultLang  = "en"
	mu           sync.RWMutex
)

func SetDefaultLang(lang string) {
	defaultLang = strings.ToLower(strings.TrimSpace(lang))
}

func LoadDir(root string) error {
	langs, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, langEntry := range langs {
		if !langEntry.IsDir() {
			continue
		}

		lang := strings.ToLower(strings.TrimSpace(langEntry.Name()))
		langPath := filepath.Join(root, langEntry.Name())

		files, err := os.ReadDir(langPath)
		if err != nil {
			return err
		}

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
				continue
			}

			feature := strings.TrimSuffix(file.Name(), ".json")
			fullPath := filepath.Join(langPath, file.Name())

			raw, err := os.ReadFile(fullPath)
			if err != nil {
				return err
			}

			var data map[string]string
			if err := json.Unmarshal(raw, &data); err != nil {
				return err
			}

			for k, v := range data {
				key := feature + "." + k
				add(lang, key, v)
			}
		}
	}

	return nil
}

func add(lang string, key string, value string) {
	mu.Lock()
	defer mu.Unlock()

	if translations[lang] == nil {
		translations[lang] = map[string]string{}
	}

	translations[lang][key] = value
}

func T(c fiber.Ctx, key string, params ...Params) string {
	lang, ok := c.Locals("lang").(string)
	if !ok || lang == "" {
		lang = defaultLang
	}

	return TLang(lang, key, params...)
}

func TLang(lang string, key string, params ...Params) string {
	lang = normalizeLang(lang)
	if lang == "" {
		lang = defaultLang
	}

	mu.RLock()
	value := translations[lang][key]
	mu.RUnlock()

	if value == "" {
		mu.RLock()
		value = translations[defaultLang][key]
		mu.RUnlock()
	}

	if value == "" {
		return key
	}

	if len(params) > 0 {
		for k, v := range params[0] {
			value = strings.ReplaceAll(value, "{"+k+"}", toString(v))
		}
	}

	return value
}

func TP(c fiber.Ctx, key string, count int, params ...Params) string {
	lang, ok := c.Locals("lang").(string)
	if !ok || lang == "" {
		lang = defaultLang
	}

	return TPLang(lang, key, count, params...)
}

func TPLang(lang string, key string, count int, params ...Params) string {
	lang = normalizeLang(lang)
	if lang == "" {
		lang = defaultLang
	}

	form := pluralForm(lang, count)

	p := Params{
		"count": count,
	}

	if len(params) > 0 {
		for k, v := range params[0] {
			p[k] = v
		}
	}

	return TLang(lang, key+"."+form, p)
}

func pluralForm(lang string, count int) string {
	switch lang {
	case "ru":
		n := count % 100
		if n >= 11 && n <= 14 {
			return "many"
		}

		switch count % 10 {
		case 1:
			return "one"
		case 2, 3, 4:
			return "few"
		default:
			return "many"
		}
	default:
		if count == 1 {
			return "one"
		}
		return "other"
	}
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if len(lang) >= 2 {
		return lang[:2]
	}
	return lang
}

func toString(v any) string {
	return fmt.Sprint(v)
}
