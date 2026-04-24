package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestStatusTranslator_UsesExactLanguage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	translator := NewStatusTranslator(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COALESCE(m.translation, sm.message)
FROM tbl_source_message sm
LEFT JOIN tbl_message m
	ON m.id = sm.id
	AND m.language = ?
WHERE sm.category = 'order_status'
  AND sm.message = ?
LIMIT 1
`)).
		WithArgs("ru", "New order").
		WillReturnRows(sqlmock.NewRows([]string{"translation"}).AddRow("Новый заказ"))

	got, err := translator.TranslateStatus(context.Background(), "ru", "New order")

	require.NoError(t, err)
	require.Equal(t, "Новый заказ", got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStatusTranslator_FallsBackToBaseLanguage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	translator := NewStatusTranslator(db)

	query := regexp.QuoteMeta(`
SELECT COALESCE(m.translation, sm.message)
FROM tbl_source_message sm
LEFT JOIN tbl_message m
	ON m.id = sm.id
	AND m.language = ?
WHERE sm.category = 'order_status'
  AND sm.message = ?
LIMIT 1
`)

	mock.ExpectQuery(query).
		WithArgs("ru-RU", "New order").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(query).
		WithArgs("ru", "New order").
		WillReturnRows(sqlmock.NewRows([]string{"translation"}).AddRow("Новый заказ"))

	got, err := translator.TranslateStatus(context.Background(), "ru-RU", "New order")

	require.NoError(t, err)
	require.Equal(t, "Новый заказ", got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStatusTranslator_ReturnsSourceNameWhenNothingFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	translator := NewStatusTranslator(db)
	query := regexp.QuoteMeta(`
SELECT COALESCE(m.translation, sm.message)
FROM tbl_source_message sm
LEFT JOIN tbl_message m
	ON m.id = sm.id
	AND m.language = ?
WHERE sm.category = 'order_status'
  AND sm.message = ?
LIMIT 1
`)

	mock.ExpectQuery(query).
		WithArgs("ru", "New order").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(query).
		WithArgs("en-US", "New order").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(query).
		WithArgs("en", "New order").
		WillReturnError(sql.ErrNoRows)

	got, err := translator.TranslateStatus(context.Background(), "ru", "New order")

	require.NoError(t, err)
	require.Equal(t, "New order", got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStatusTranslator_UsesCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	translator := NewStatusTranslator(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COALESCE(m.translation, sm.message)
FROM tbl_source_message sm
LEFT JOIN tbl_message m
	ON m.id = sm.id
	AND m.language = ?
WHERE sm.category = 'order_status'
  AND sm.message = ?
LIMIT 1
`)).
		WithArgs("ru", "New order").
		WillReturnRows(sqlmock.NewRows([]string{"translation"}).AddRow("Новый заказ"))

	first, err := translator.TranslateStatus(context.Background(), "ru", "New order")
	require.NoError(t, err)
	require.Equal(t, "Новый заказ", first)

	second, err := translator.TranslateStatus(context.Background(), "ru", "New order")
	require.NoError(t, err)
	require.Equal(t, "Новый заказ", second)

	require.NoError(t, mock.ExpectationsWereMet())
}
