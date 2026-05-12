package database

import (
	"os"
	"strings"
)

// PreferSimpleProtocolEnabled disables named prepared statements (pgx simple protocol).
// Default on — required for PgBouncer transaction pooling and avoids:
// FATAL: prepared statement name is already in use (SQLSTATE 08P01)
//
// Set PG_USE_SIMPLE_PROTOCOL=false only for a direct Postgres connection when you rely on prepares.
func PreferSimpleProtocolEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("PG_USE_SIMPLE_PROTOCOL")))
	return v != "0" && v != "false" && v != "no" && v != "off"
}

// PostgresDSNWithPoolerCompat keeps prefer_simple_protocol in the DSN string as a fallback.
// PreferSimpleProtocolEnabled is applied via postgres.Config.PreferSimpleProtocol in ConnectDB — that path is authoritative for GORM.
func PostgresDSNWithPoolerCompat(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if !PreferSimpleProtocolEnabled() || dsn == "" {
		return dsn
	}
	if strings.Contains(dsn, "prefer_simple_protocol") {
		return dsn
	}

	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		return dsn + sep + "prefer_simple_protocol=true"
	}

	if strings.HasSuffix(dsn, " ") {
		return dsn + "prefer_simple_protocol=true"
	}
	return dsn + " prefer_simple_protocol=true"
}
