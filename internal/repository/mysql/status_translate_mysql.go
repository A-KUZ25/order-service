package mysql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
)

type StatusTranslator struct {
	db              *sql.DB
	defaultLanguage string
	cache           sync.Map
}

func NewStatusTranslator(db *sql.DB) *StatusTranslator {
	return &StatusTranslator{
		db:              db,
		defaultLanguage: "en-US",
	}
}

func (t *StatusTranslator) TranslateStatus(
	ctx context.Context,
	language, name string,
) (string, error) {
	for _, lang := range t.languageCandidates(language) {
		translated, found, err := t.translateStatusWithLanguage(ctx, lang, name)
		if err != nil {
			return "", err
		}
		if found {
			return translated, nil
		}
	}

	return name, nil
}

func (t *StatusTranslator) translateStatusWithLanguage(
	ctx context.Context,
	language, name string,
) (string, bool, error) {
	cacheKey := language + "\x00" + name
	if cached, ok := t.cache.Load(cacheKey); ok {
		return cached.(string), true, nil
	}

	const query = `
SELECT COALESCE(m.translation, sm.message)
FROM tbl_source_message sm
LEFT JOIN tbl_message m
	ON m.id = sm.id
	AND m.language = ?
WHERE sm.category = 'order_status'
  AND sm.message = ?
LIMIT 1
`

	var translated string
	err := t.db.QueryRowContext(ctx, query, language, name).Scan(&translated)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	t.cache.Store(cacheKey, translated)
	return translated, true, nil
}

// todo разобраться
func (t *StatusTranslator) languageCandidates(language string) []string {
	if language == "" {
		language = t.defaultLanguage
	}

	candidates := []string{language}

	if base, _, ok := strings.Cut(language, "-"); ok && base != "" && base != language {
		candidates = append(candidates, base)
	}

	if t.defaultLanguage != "" && t.defaultLanguage != language {
		candidates = append(candidates, t.defaultLanguage)

		if base, _, ok := strings.Cut(t.defaultLanguage, "-"); ok && base != "" && base != t.defaultLanguage {
			candidates = append(candidates, base)
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		result = append(result, candidate)
	}

	return result
}
