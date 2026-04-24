package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestShowOrderCodeProvider_UsesTenantSetting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	provider := NewShowOrderCodeProvider(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT value
FROM tbl_tenant_setting
WHERE tenant_id = ?
  AND name = ?
 AND city_id = ? AND position_id = ? LIMIT 1`)).
		WithArgs(int64(68), settingShowOrderCode, int64(26068), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("1"))

	got, err := provider.ShouldShowOrderCode(context.Background(), 68, 26068, 7)

	require.NoError(t, err)
	require.True(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestShowOrderCodeProvider_FallsBackToDefaultSetting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	provider := NewShowOrderCodeProvider(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT value
FROM tbl_tenant_setting
WHERE tenant_id = ?
  AND name = ?
 AND city_id = ? AND position_id = ? LIMIT 1`)).
		WithArgs(int64(68), settingShowOrderCode, int64(26068), int64(7)).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT value
FROM tbl_default_settings
WHERE name = ?
LIMIT 1
`)).
		WithArgs(settingShowOrderCode).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("0"))

	got, err := provider.ShouldShowOrderCode(context.Background(), 68, 26068, 7)

	require.NoError(t, err)
	require.False(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestShowOrderCodeProvider_UsesCacheBeforeTTL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	provider := NewShowOrderCodeProvider(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT value
FROM tbl_tenant_setting
WHERE tenant_id = ?
  AND name = ?
 AND city_id = ? AND position_id = ? LIMIT 1`)).
		WithArgs(int64(68), settingShowOrderCode, int64(26068), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("1"))

	first, err := provider.ShouldShowOrderCode(context.Background(), 68, 26068, 7)
	require.NoError(t, err)
	require.True(t, first)

	second, err := provider.ShouldShowOrderCode(context.Background(), 68, 26068, 7)
	require.NoError(t, err)
	require.True(t, second)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestShowOrderCodeProvider_ReloadsAfterTTL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	provider := NewShowOrderCodeProvider(db)
	key := "68:26068:7"
	provider.cache.Store(key, boolCacheEntry{
		value:     true,
		expiresAt: time.Now().Add(-time.Second),
	})

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT value
FROM tbl_tenant_setting
WHERE tenant_id = ?
  AND name = ?
 AND city_id = ? AND position_id = ? LIMIT 1`)).
		WithArgs(int64(68), settingShowOrderCode, int64(26068), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("0"))

	got, err := provider.ShouldShowOrderCode(context.Background(), 68, 26068, 7)

	require.NoError(t, err)
	require.False(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}
