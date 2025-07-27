package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	GormDB *gorm.DB
	sqlDB  *sql.DB
	once   sync.Once
)

// InitDB 
func InitDB(dbuser,dbpassword,dbname string) error {
	var err error
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", dbuser),
		getEnv("DB_PASS", dbpassword),
		getEnv("DB_NAME", dbname),
		getEnv("DB_PORT", "5432"),
	)

	once.Do(func() {
		GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Printf("GORM DB connection failed: %v", err)
			return
		}

		sqlDB, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("SQL DB connection failed: %v", err)
		}
	})

	return err
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

// DBManager interface for DB operations
type DBManager interface {
	Insert(ctx context.Context, table string, data map[string]any) error
	Update(ctx context.Context, table string, data map[string]any, where map[string]any) error
	Delete(ctx context.Context, table string, where map[string]any) error
	Get(ctx context.Context, table string, where map[string]any) ([]map[string]any, error)
}

// PostgresManager uses the globally initialized sqlDB
type PostgresManager struct{}

func NewPostgresManager() *PostgresManager {
	return &PostgresManager{}
}

func (p *PostgresManager) ExecQuery(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return sqlDB.ExecContext(ctx, query, args...)
}

// Insert record
func (p *PostgresManager) Insert(ctx context.Context, table string, data map[string]any) error {
	keys, values, placeholders := prepareSQLComponents(data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(keys, ", "), strings.Join(placeholders, ", "))
	_, err := sqlDB.ExecContext(ctx, query, values...)
	return err
}

func (p *PostgresManager) Update(ctx context.Context, table string, data, where map[string]any) error {

//	log.Printf("[************ - CALLING THE UPDATE CALL FOR POSTGRES DB***************]")
//	log.Printf("[***** table ****] %s", table)
//	log.Printf("[****** data *****] %s", data)
//	log.Printf("[***** WHERE **** ]%s", where)
	setParts := []string{}
	args := []any{}
	i := 1

	// SET clauses
	for k, v := range data {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}

	// WHERE clauses
	whereParts := []string{}
	for k, v := range where {
		switch vals := v.(type) {
		case []any:
			placeholders := []string{}
			for _, val := range vals {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i))
				args = append(args, val)
				i++
			}
			whereParts = append(whereParts, fmt.Sprintf("%s IN (%s)", k, strings.Join(placeholders, ", ")))
		case []string:
			placeholders := []string{}
			for _, val := range vals {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i))
				args = append(args, val)
				i++
			}
			whereParts = append(whereParts, fmt.Sprintf("%s IN (%s)", k, strings.Join(placeholders, ", ")))
		case []int:
			placeholders := []string{}
			for _, val := range vals {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i))
				args = append(args, val)
				i++
			}
			whereParts = append(whereParts, fmt.Sprintf("%s IN (%s)", k, strings.Join(placeholders, ", ")))
		default:
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", k, i))
			args = append(args, v)
			i++
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table,
		strings.Join(setParts, ", "),
		strings.Join(whereParts, " AND "),
	)

	_, err := sqlDB.ExecContext(ctx, query, args...)
	return err
}

// Delete record
func (p *PostgresManager) Delete(ctx context.Context, table string, where map[string]any) error {
	whereSQL, args := buildSQLSet(where, 1)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereSQL)
	_, err := sqlDB.ExecContext(ctx, query, args...)
	return err
}

// Get records
func (p *PostgresManager) Get(ctx context.Context, table string, where map[string]any) ([]map[string]any, error) {
	whereSQL, args := buildSQLSet(where, 1)
	query := fmt.Sprintf("SELECT * FROM %s", table)
	if len(whereSQL) > 0 {
		query += " WHERE " + whereSQL
	}
	rows, err := sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseRows(rows)
}

// --- Helpers ---
func prepareSQLComponents(data map[string]any) (keys []string, values []any, placeholders []string) {
	i := 1
	for k, v := range data {
		keys = append(keys, k)
		values = append(values, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	return
}

func buildSQLSet(data map[string]any, startIdx int) (string, []any) {
	parts := []string{}
	args := []any{}
	i := startIdx
	for k, v := range data {
		parts = append(parts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	return strings.Join(parts, " AND "), args
}

func parseRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, _ := rows.Columns()
	results := []map[string]any{}
	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, col := range cols {
			val := columnPointers[i].(*any)
			row[col] = *val
		}
		results = append(results, row)
	}
	return results, nil
}
