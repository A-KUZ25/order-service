//go:build integration

package mysql

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"orders-service/internal/app/order"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOrdersRepository_FetchOrdersByStatusGroup(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForStatusGroup(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	statusTimeFrom := int64(1711062000)
	statusTimeTo := int64(1711148399)

	got, err := repo.FetchOrdersByStatusGroup(context.Background(), order.BaseFilter{
		TenantID:       68,
		CityIDs:        []int64{26068},
		Status:         []int64{1, 6},
		Tariffs:        []int64{1033},
		UserPositions:  []int64{1},
		SelectForDate:  true,
		StatusTimeFrom: &statusTimeFrom,
		StatusTimeTo:   &statusTimeTo,
		SortField:      "o.order_id",
		SortOrder:      "asc",
	})

	require.NoError(t, err)
	require.Equal(t, []int64{1001, 1002}, got)
}

func TestOrdersRepository_FetchUnpaid(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForWarningQueries(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	got, err := repo.FetchUnpaid(context.Background(), order.UnpaidFilter{
		BaseFilter:             baseWarningFilter(),
		StatusCompletedNotPaid: 52,
	})

	require.NoError(t, err)
	require.Equal(t, []int64{2001}, got)
}

func TestOrdersRepository_FetchBadReview(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForWarningQueries(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	got, err := repo.FetchBadReview(context.Background(), order.BadReviewFilter{
		BaseFilter:   baseWarningFilter(),
		BadRatingMax: 3,
	})

	require.NoError(t, err)
	require.Equal(t, []int64{2004, 2005}, got)
}

func TestOrdersRepository_FetchExceededPrice(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForWarningQueries(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	got, err := repo.FetchExceededPrice(context.Background(), order.ExceededPriceFilter{
		BaseFilter:     baseWarningFilter(),
		MinRealPrice:   100,
		FinishedStatus: []int64{37, 38},
	})

	require.NoError(t, err)
	require.Equal(t, []int64{2001, 2004, 2011}, got)
}

func TestOrdersRepository_CountOrdersWithWarning_UsesBaseFilter(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForCountQueries(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	count, err := repo.CountOrdersWithWarning(context.Background(), order.BaseFilter{
		TenantID:      68,
		CityIDs:       []int64{26068},
		Status:        []int64{1, 6},
		Tariffs:       []int64{1033},
		UserPositions: []int64{1},
	}, nil)

	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestOrdersRepository_CountOrdersWithWarning_AddsWarningIDsWithOR(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForCountQueries(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	count, err := repo.CountOrdersWithWarning(context.Background(), order.BaseFilter{
		TenantID:      68,
		CityIDs:       []int64{26068},
		Status:        []int64{1},
		Tariffs:       []int64{1033},
		UserPositions: []int64{1},
	}, []int64{3003, 3004})

	require.NoError(t, err)
	require.Equal(t, int64(3), count)
}

func TestOrdersRepository_FetchOrdersWithWarning_AppliesORSortingAndPagination(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOrdersForFetchWithWarning(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	got, err := repo.FetchOrdersWithWarning(context.Background(), order.BaseFilter{
		TenantID:      68,
		CityIDs:       []int64{26068},
		Status:        []int64{1},
		Tariffs:       []int64{1033},
		UserPositions: []int64{1},
		SortField:     "o.order_id",
		SortOrder:     "asc",
	}, []int64{4003, 4004}, 1, 2)

	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, int64(4003), got[0].OrderID)
	require.Equal(t, int64(4004), got[1].OrderID)
	require.Equal(t, int64(69), got[0].TenantID)
	require.Equal(t, int64(99999), got[1].CityID.Int64)
	require.Equal(t, "New order", got[0].StatusName)
	require.Equal(t, "New order", got[1].StatusName)
}

func TestOrdersRepository_GetOptionsForOrders(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedOptionsData(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	got, err := repo.GetOptionsForOrders(context.Background(), []int64{5001, 5002})

	require.NoError(t, err)
	require.Equal(t, map[int64][]order.OptionDTO{
		5001: {
			{OptionID: 10, Name: "WiFi", Quantity: 1},
			{OptionID: 11, Name: "Child seat", Quantity: 2},
		},
		5002: {
			{OptionID: 11, Name: "Child seat", Quantity: 1},
		},
	}, got)
}

func TestOrdersRepository_GetStatusChangeTimes(t *testing.T) {
	db, cleanup := setupIntegrationMySQL(t)
	defer cleanup()

	createOrderSchema(t, db)
	seedStatusChangeData(t, db)

	repo, err := NewOrdersRepository(db)
	require.NoError(t, err)

	got, err := repo.GetStatusChangeTimes(context.Background(), []order.StatusKey{
		{OrderID: 6001, StatusID: 1},
		{OrderID: 6002, StatusID: 6},
		{OrderID: 6003, StatusID: 1},
	})

	require.NoError(t, err)
	require.Equal(t, map[order.StatusKey]int64{
		{OrderID: 6001, StatusID: 1}: 1711065600,
		{OrderID: 6002, StatusID: 6}: 1711065700,
	}, got)
}

func setupIntegrationMySQL(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	ctx := context.Background()

	container, err := tcmysql.Run(ctx,
		"mysql:8.0.36",
		tcmysql.WithDatabase("orders_test"),
		tcmysql.WithUsername("test"),
		tcmysql.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server - GPL").
				WithStartupTimeout(90*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "parseTime=true", "multiStatements=true")
	require.NoError(t, err)

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return db.PingContext(ctx) == nil
	}, 30*time.Second, 300*time.Millisecond)

	cleanup := func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}

	return db, cleanup
}

func createOrderSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	statements := []string{
		`CREATE TABLE tbl_order (
			order_id BIGINT PRIMARY KEY,
			tenant_id BIGINT NOT NULL,
			worker_id BIGINT NULL,
			car_id BIGINT NULL,
			active TINYINT NOT NULL,
			status_id BIGINT NOT NULL,
			status_time BIGINT NOT NULL,
			user_create BIGINT NULL,
			user_modifed BIGINT NULL,
			company_id BIGINT NULL,
			parking_id BIGINT NULL,
			address TEXT NULL,
			comment TEXT NULL,
			predv_price DECIMAL(10,2) NULL,
			predv_price_no_discount DECIMAL(10,2) NULL,
			device VARCHAR(32) NULL,
			order_number BIGINT NOT NULL DEFAULT 0,
			payment VARCHAR(32) NULL,
			show_phone BIGINT NULL,
			create_time BIGINT NULL,
			time_to_client BIGINT NULL,
			client_device_token VARCHAR(255) NULL,
			app_id BIGINT NULL,
			order_time BIGINT NOT NULL,
			predv_distance DECIMAL(10,2) NULL,
			predv_time BIGINT NULL,
			call_warning_id BIGINT NULL,
			phone VARCHAR(32) NULL,
			client_id BIGINT NOT NULL DEFAULT 0,
			bonus_payment BIGINT NULL,
			currency_id BIGINT NOT NULL DEFAULT 0,
			time_offset BIGINT NOT NULL DEFAULT 0,
			is_fix BIGINT NOT NULL DEFAULT 0,
			update_time BIGINT NULL,
			deny_refuse_order BIGINT NULL,
			city_id BIGINT NOT NULL,
			tariff_id BIGINT NOT NULL,
			position_id BIGINT NOT NULL,
			promo_code_id BIGINT NULL,
			tenant_company_id BIGINT NULL,
			mark BIGINT NULL,
			processed_exchange_program_id BIGINT NULL,
			client_passenger_id BIGINT NULL,
			client_passenger_phone VARCHAR(32) NULL,
			is_pre_order BIGINT NULL,
			app_version VARCHAR(64) NULL,
			agent_commission DECIMAL(10,2) NULL,
			is_fix_by_dispatcher BIGINT NULL,
			finish_time BIGINT NULL,
			comment_for_dispatcher TEXT NULL,
			worker_manual_surcharge DECIMAL(10,2) NULL,
			realtime_price DECIMAL(10,2) NOT NULL DEFAULT 0,
			unit_quantity DECIMAL(10,2) NULL,
			shop_id BIGINT NULL,
			require_prepayment BIGINT NULL,
			order_code VARCHAR(64) NULL,
			client_offered_price DECIMAL(10,2) NULL,
			idempotent_key VARCHAR(64) NULL,
			additional_tariff_id BIGINT NULL,
			initial_price DECIMAL(10,2) NULL,
			time_to_order BIGINT NULL,
			sort BIGINT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_client_review (
			order_id BIGINT NOT NULL,
			rating BIGINT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_client (
			client_id BIGINT PRIMARY KEY,
			phone VARCHAR(32) NULL,
			name VARCHAR(255) NULL,
			last_name VARCHAR(255) NULL,
			second_name VARCHAR(255) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_order_status (
			status_id BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_worker (
			worker_id BIGINT PRIMARY KEY,
			callsign BIGINT NULL,
			name VARCHAR(255) NULL,
			last_name VARCHAR(255) NULL,
			second_name VARCHAR(255) NULL,
			phone VARCHAR(32) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_car (
			car_id BIGINT PRIMARY KEY,
			name VARCHAR(255) NULL,
			color BIGINT NULL,
			gos_number VARCHAR(32) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_taxi_tariff (
			tariff_id BIGINT PRIMARY KEY,
			tariff_type VARCHAR(64) NULL,
			name VARCHAR(255) NULL,
			quantitative_title VARCHAR(255) NULL,
			price_for_unit DECIMAL(10,2) NULL,
			unit_name VARCHAR(64) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_order_detail_cost (
			order_id BIGINT PRIMARY KEY,
			summary_cost VARCHAR(64) NULL,
			summary_cost_no_discount VARCHAR(64) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_order_has_option (
			order_id BIGINT NOT NULL,
			option_id BIGINT NOT NULL,
			quantity BIGINT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_car_option (
			option_id BIGINT PRIMARY KEY,
			name VARCHAR(255) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_order_change_data (
			order_id BIGINT NOT NULL,
			change_field VARCHAR(64) NOT NULL,
			change_val BIGINT NOT NULL,
			change_time BIGINT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_user (
			user_id BIGINT PRIMARY KEY,
			name VARCHAR(255) NULL,
			last_name VARCHAR(255) NULL,
			second_name VARCHAR(255) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE tbl_currency (
			currency_id BIGINT PRIMARY KEY,
			name VARCHAR(255) NULL,
			code VARCHAR(32) NULL,
			symbol VARCHAR(16) NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		require.NoError(t, err)
	}
}

func seedOrdersForStatusGroup(t *testing.T, db *sql.DB) {
	t.Helper()

	insert := strings.Join([]string{
		`INSERT INTO tbl_order (
			order_id,
			tenant_id,
			active,
			status_id,
			status_time,
			order_time,
			city_id,
			tariff_id,
			position_id,
			realtime_price
		) VALUES`,
		`(1001, 68, 1, 1, 1711065600, 1711065600, 26068, 1033, 1, 0),`,
		`(1002, 68, 1, 6, 1711070000, 1711070000, 26068, 1033, 1, 0),`,
		`(1003, 68, 1, 1, 1711071000, 1711071000, 99999, 1033, 1, 0),`,
		`(1004, 68, 0, 1, 1711072000, 1711072000, 26068, 1033, 1, 0),`,
		`(1005, 69, 1, 1, 1711073000, 1711073000, 26068, 1033, 1, 0),`,
		`(1006, 68, 1, 17, 1711074000, 1711074000, 26068, 1033, 1, 0),`,
		`(1007, 68, 1, 1, 1711152000, 1711152000, 26068, 1033, 1, 0),`,
		`(1008, 68, 1, 1, 1711075000, 1711075000, 26068, 2000, 1, 0),`,
		`(1009, 68, 1, 1, 1711076000, 1711076000, 26068, 1033, 99, 0)`,
	}, "\n")

	_, err := db.Exec(insert)
	require.NoError(t, err)
}

func seedOrdersForWarningQueries(t *testing.T, db *sql.DB) {
	t.Helper()

	insertOrders := strings.Join([]string{
		`INSERT INTO tbl_order (
			order_id,
			tenant_id,
			active,
			status_id,
			status_time,
			order_time,
			city_id,
			tariff_id,
			position_id,
			realtime_price
		) VALUES`,
		`(2001, 68, 1, 52, 1711065600, 1711065600, 26068, 1033, 1, 150),`,
		`(2002, 68, 1, 52, 1711065700, 1711065700, 99999, 1033, 1, 150),`,
		`(2003, 68, 1, 52, 1711065800, 1711065800, 26068, 2000, 1, 150),`,
		`(2004, 68, 1, 17, 1711065900, 1711065900, 26068, 1033, 1, 120),`,
		`(2005, 68, 1, 38, 1711066000, 1711066000, 26068, 1033, 1, 170),`,
		`(2006, 68, 1, 17, 1711066100, 1711066100, 26068, 1033, 99, 120),`,
		`(2007, 69, 1, 52, 1711066200, 1711066200, 26068, 1033, 1, 150),`,
		`(2008, 68, 0, 52, 1711066300, 1711066300, 26068, 1033, 1, 150),`,
		`(2009, 68, 1, 17, 1711152000, 1711152000, 26068, 1033, 1, 120),`,
		`(2010, 68, 1, 17, 1711066400, 1711066400, 26068, 1033, 1, 90),`,
		`(2011, 68, 1, 17, 1711066500, 1711066500, 26068, 1033, 1, 120)`,
	}, "\n")

	_, err := db.Exec(insertOrders)
	require.NoError(t, err)

	insertReviews := strings.Join([]string{
		`INSERT INTO tbl_client_review (order_id, rating) VALUES`,
		`(2004, 2),`,
		`(2005, 1),`,
		`(2010, 4)`,
	}, "\n")

	_, err = db.Exec(insertReviews)
	require.NoError(t, err)
}

func seedOrdersForCountQueries(t *testing.T, db *sql.DB) {
	t.Helper()

	insertOrders := strings.Join([]string{
		`INSERT INTO tbl_order (
			order_id,
			tenant_id,
			active,
			status_id,
			status_time,
			order_time,
			city_id,
			tariff_id,
			position_id,
			realtime_price
		) VALUES`,
		`(3001, 68, 1, 1, 1711065600, 1711065600, 26068, 1033, 1, 0),`,
		`(3002, 68, 1, 6, 1711065700, 1711065700, 26068, 1033, 1, 0),`,
		`(3003, 68, 1, 17, 1711065800, 1711065800, 99999, 2000, 99, 0),`,
		`(3004, 69, 1, 1, 1711065900, 1711065900, 26068, 1033, 1, 0),`,
		`(3005, 68, 0, 1, 1711066000, 1711066000, 26068, 1033, 1, 0),`,
		`(3006, 68, 1, 1, 1711066100, 1711066100, 26068, 2000, 1, 0),`,
		`(3007, 68, 1, 1, 1711066200, 1711066200, 26068, 1033, 99, 0)`,
	}, "\n")

	_, err := db.Exec(insertOrders)
	require.NoError(t, err)
}

func seedOrdersForFetchWithWarning(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO tbl_order_status (status_id, name) VALUES
		(1, 'New order')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO tbl_taxi_tariff (tariff_id, tariff_type, name) VALUES
		(1033, 'main', 'Tariff 1033')
	`)
	require.NoError(t, err)

	insertOrders := strings.Join([]string{
		`INSERT INTO tbl_order (
			order_id,
			tenant_id,
			active,
			status_id,
			status_time,
			order_time,
			address,
			city_id,
			tariff_id,
			position_id,
			realtime_price,
			order_number,
			client_id,
			currency_id
		) VALUES`,
		`(4001, 68, 1, 1, 1711065600, 1711065600, '', 26068, 1033, 1, 0, 5001, 0, 0),`,
		`(4002, 68, 1, 1, 1711065700, 1711065700, '', 26068, 1033, 1, 0, 5002, 0, 0),`,
		`(4003, 69, 1, 1, 1711065800, 1711065800, '', 26068, 1033, 1, 0, 5003, 0, 0),`,
		`(4004, 68, 1, 1, 1711065900, 1711065900, '', 99999, 1033, 1, 0, 5004, 0, 0)`,
	}, "\n")

	_, err = db.Exec(insertOrders)
	require.NoError(t, err)
}

func seedOptionsData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO tbl_car_option (option_id, name) VALUES
		(10, 'WiFi'),
		(11, 'Child seat')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO tbl_order_has_option (order_id, option_id, quantity) VALUES
		(5001, 10, 1),
		(5001, 11, 2),
		(5002, 11, 1),
		(9999, 10, 9)
	`)
	require.NoError(t, err)
}

func seedStatusChangeData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO tbl_order_change_data (order_id, change_field, change_val, change_time) VALUES
		(6001, 'status_id', 1, 1711065600),
		(6002, 'status_id', 6, 1711065700),
		(6002, 'worker_id', 77, 1711065800),
		(7001, 'status_id', 1, 1711065900)
	`)
	require.NoError(t, err)
}

func baseWarningFilter() order.BaseFilter {
	statusTimeFrom := int64(1711062000)
	statusTimeTo := int64(1711148399)

	return order.BaseFilter{
		TenantID:       68,
		CityIDs:        []int64{26068},
		Tariffs:        []int64{1033},
		UserPositions:  []int64{1},
		SelectForDate:  true,
		StatusTimeFrom: &statusTimeFrom,
		StatusTimeTo:   &statusTimeTo,
		SortField:      "o.order_id",
		SortOrder:      "asc",
	}
}
