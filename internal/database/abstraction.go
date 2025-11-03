package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql" // MySQL/MariaDB driver
)

// DatabaseType represents the type of database backend
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	MariaDB    DatabaseType = "mariadb"
)

// DatabaseConnection is the abstraction interface for all database operations
type DatabaseConnection interface {
	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Get and Select for struct mapping
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// Transaction support
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)

	// Database info
	GetType() DatabaseType
	Ping(ctx context.Context) error
	Close() error

	// Raw connection (for advanced use)
	GetDB() *sqlx.DB
}

// UniversalDB implements DatabaseConnection for any database type
type UniversalDB struct {
	db       *sqlx.DB
	dbType   DatabaseType
	driverName string
}

// DetectDatabaseType detects the database type from connection string
func DetectDatabaseType(dsn string) (DatabaseType, error) {
	switch {
	case strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://"):
		return PostgreSQL, nil
	case strings.HasPrefix(dsn, "mysql://"):
		// Need to check if it's MariaDB by connecting
		return MySQL, nil // Will be refined after connection
	default:
		// Try to detect from DSN format
		if strings.Contains(dsn, "@tcp(") || strings.Contains(dsn, "@unix(") {
			return MySQL, nil
		}
		return "", fmt.Errorf("unknown database type from DSN: %s", dsn)
	}
}

// NewDatabaseConnection creates a new database connection with auto-detection
func NewDatabaseConnection(dsn string, maxConnections int) (DatabaseConnection, error) {
	// Detect database type
	dbType, err := DetectDatabaseType(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to detect database type: %w", err)
	}

	// Determine driver name
	var driverName string
	var connectionString string

	switch dbType {
	case PostgreSQL:
		driverName = "postgres"
		connectionString = dsn

	case MySQL, MariaDB:
		driverName = "mysql"
		// Convert mysql:// to proper MySQL DSN format
		connectionString = convertMySQLDSN(dsn)
	}

	// Open connection
	db, err := sqlx.Open(driverName, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxConnections / 3)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// For MySQL, check if it's actually MariaDB
	if dbType == MySQL {
		actualType, err := detectMySQLOrMariaDB(ctx, db)
		if err == nil {
			dbType = actualType
		}
	}

	return &UniversalDB{
		db:         db,
		dbType:     dbType,
		driverName: driverName,
	}, nil
}

// Query executes a query that returns rows
func (udb *UniversalDB) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	// Translate query syntax if needed
	translatedQuery := udb.translateQuery(query)
	return udb.db.QueryxContext(ctx, translatedQuery, args...)
}

// QueryRow executes a query that returns a single row
func (udb *UniversalDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	translatedQuery := udb.translateQuery(query)
	return udb.db.QueryRowxContext(ctx, translatedQuery, args...)
}

// Exec executes a query without returning rows
func (udb *UniversalDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	translatedQuery := udb.translateQuery(query)
	return udb.db.ExecContext(ctx, translatedQuery, args...)
}

// Get executes a query and maps a single row to a struct
func (udb *UniversalDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	translatedQuery := udb.translateQuery(query)
	return udb.db.GetContext(ctx, dest, translatedQuery, args...)
}

// Select executes a query and maps multiple rows to a slice of structs
func (udb *UniversalDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	translatedQuery := udb.translateQuery(query)
	return udb.db.SelectContext(ctx, dest, translatedQuery, args...)
}

// BeginTx starts a new transaction
func (udb *UniversalDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return udb.db.BeginTxx(ctx, opts)
}

// GetType returns the database type
func (udb *UniversalDB) GetType() DatabaseType {
	return udb.dbType
}

// Ping checks database connectivity
func (udb *UniversalDB) Ping(ctx context.Context) error {
	return udb.db.PingContext(ctx)
}

// Close closes the database connection
func (udb *UniversalDB) Close() error {
	return udb.db.Close()
}

// GetDB returns the underlying sqlx.DB (for advanced use)
func (udb *UniversalDB) GetDB() *sqlx.DB {
	return udb.db
}

// translateQuery translates SQL syntax between database types
func (udb *UniversalDB) translateQuery(query string) string {
	// PostgreSQL uses $1, $2, $3 placeholders
	// MySQL uses ? placeholders

	switch udb.dbType {
	case PostgreSQL:
		// Query is already in PostgreSQL format
		return query

	case MySQL, MariaDB:
		// Convert PostgreSQL $1, $2 to MySQL ?
		// This is a simple implementation - may need enhancement
		translated := query
		for i := 20; i >= 1; i-- {
			placeholder := fmt.Sprintf("$%d", i)
			translated = strings.ReplaceAll(translated, placeholder, "?")
		}
		return translated

	default:
		return query
	}
}

// convertMySQLDSN converts mysql:// URL to MySQL DSN format
func convertMySQLDSN(dsn string) string {
	// mysql://user:password@host:port/database
	// becomes: user:password@tcp(host:port)/database

	if !strings.HasPrefix(dsn, "mysql://") {
		return dsn // Already in correct format
	}

	// Remove mysql:// prefix
	dsn = strings.TrimPrefix(dsn, "mysql://")

	// Parse components
	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		return dsn // Invalid format, return as-is
	}

	credentials := parts[0]
	hostAndDB := parts[1]

	hostParts := strings.Split(hostAndDB, "/")
	if len(hostParts) != 2 {
		return dsn // Invalid format
	}

	host := hostParts[0]
	database := hostParts[1]

	// Construct MySQL DSN
	return fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4", credentials, host, database)
}

// detectMySQLOrMariaDB determines if connected database is MySQL or MariaDB
func detectMySQLOrMariaDB(ctx context.Context, db *sqlx.DB) (DatabaseType, error) {
	var version string
	err := db.GetContext(ctx, &version, "SELECT VERSION()")
	if err != nil {
		return MySQL, err
	}

	// MariaDB version strings contain "MariaDB"
	if strings.Contains(strings.ToLower(version), "mariadb") {
		return MariaDB, nil
	}

	return MySQL, nil
}

// QueryBuilder provides database-agnostic query building
type QueryBuilder struct {
	dbType DatabaseType
}

// NewQueryBuilder creates a new query builder for the given database type
func NewQueryBuilder(dbType DatabaseType) *QueryBuilder {
	return &QueryBuilder{dbType: dbType}
}

// Placeholder returns the appropriate placeholder for the given position
func (qb *QueryBuilder) Placeholder(position int) string {
	switch qb.dbType {
	case PostgreSQL:
		return fmt.Sprintf("$%d", position)
	case MySQL, MariaDB:
		return "?"
	default:
		return "?"
	}
}

// CurrentTimestamp returns the current timestamp function
func (qb *QueryBuilder) CurrentTimestamp() string {
	switch qb.dbType {
	case PostgreSQL:
		return "NOW()"
	case MySQL, MariaDB:
		return "NOW()"
	default:
		return "NOW()"
	}
}

// AutoIncrement returns the auto-increment syntax
func (qb *QueryBuilder) AutoIncrement() string {
	switch qb.dbType {
	case PostgreSQL:
		return "SERIAL"
	case MySQL, MariaDB:
		return "AUTO_INCREMENT"
	default:
		return "AUTO_INCREMENT"
	}
}

// JSONType returns the JSON column type
func (qb *QueryBuilder) JSONType() string {
	switch qb.dbType {
	case PostgreSQL:
		return "JSONB"
	case MySQL, MariaDB:
		return "JSON"
	default:
		return "TEXT"
	}
}

// LimitOffset returns the limit/offset clause
func (qb *QueryBuilder) LimitOffset(limit, offset int) string {
	// Same syntax for all databases
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

// OnConflict returns the appropriate upsert syntax
func (qb *QueryBuilder) OnConflict(constraint string, update string) string {
	switch qb.dbType {
	case PostgreSQL:
		return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", constraint, update)
	case MySQL, MariaDB:
		return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", update)
	default:
		return ""
	}
}
