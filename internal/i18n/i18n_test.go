package i18n

import (
	"path/filepath"
	"testing"
)

func TestLoadDir_ProjectLocales(t *testing.T) {
	resetTranslations()

	localesDir := filepath.Join("..", "..", "locales")
	if err := LoadDir(localesDir); err != nil {
		t.Fatalf("load locales: %v", err)
	}
}

func resetTranslations() {
	mu.Lock()
	defer mu.Unlock()

	translations = map[string]map[string]string{}
	defaultLang = "en"
}
