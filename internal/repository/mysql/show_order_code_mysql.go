package mysql

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
)

const settingShowOrderCode = "SHOW_ORDER_CODE"

type ShowOrderCodeProvider struct {
	db    *sql.DB
	cache sync.Map
}

func NewShowOrderCodeProvider(db *sql.DB) *ShowOrderCodeProvider {
	return &ShowOrderCodeProvider{
		db: db,
	}
}

func (p *ShowOrderCodeProvider) ShouldShowOrderCode(
	ctx context.Context,
	tenantID, cityID, positionID int64,
) (bool, error) {
	cacheKey := strconv.FormatInt(tenantID, 10) + ":" +
		strconv.FormatInt(cityID, 10) + ":" +
		strconv.FormatInt(positionID, 10)
	if cached, ok := p.cache.Load(cacheKey); ok {
		return cached.(bool), nil
	}

	value, found, err := p.loadTenantSetting(ctx, tenantID, cityID, positionID)
	if err != nil {
		return false, err
	}
	if !found {
		value, err = p.loadDefaultSetting(ctx)
		if err != nil {
			return false, err
		}
	}

	result := value == "1"
	p.cache.Store(cacheKey, result)
	return result, nil
}

func (p *ShowOrderCodeProvider) loadTenantSetting(
	ctx context.Context,
	tenantID, cityID, positionID int64,
) (string, bool, error) {
	query := `
SELECT value
FROM tbl_tenant_setting
WHERE tenant_id = ?
  AND name = ?
`

	args := []any{tenantID, settingShowOrderCode}

	if cityID > 0 {
		query += " AND city_id = ?"
		args = append(args, cityID)
	}

	if positionID > 0 {
		query += " AND position_id = ?"
		args = append(args, positionID)
	}

	query += " LIMIT 1"

	var value string
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return value, true, nil
}

func (p *ShowOrderCodeProvider) loadDefaultSetting(
	ctx context.Context,
) (string, error) {
	const query = `
SELECT value
FROM tbl_default_settings
WHERE name = ?
LIMIT 1
`

	var value string
	err := p.db.QueryRowContext(ctx, query, settingShowOrderCode).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return value, nil
}
