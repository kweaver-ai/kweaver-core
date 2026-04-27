package writer

import (
	"database/sql"
	"database/sql/driver"

	"github.com/kweaver-ai/proton-rds-sdk-go/driver/dmdb"
	"github.com/lib/pq"
)

// DM8DriverWrapper wraps the DM8 driver to implement sql.Driver interface
type DM8DriverWrapper struct{}

func (d DM8DriverWrapper) Open(dsn string) (driver.Conn, error) {
	return dmdb.Open(dsn)
}

func (d DM8DriverWrapper) OpenConnector(dsn string) (driver.Connector, error) {
	return dmdb.OpenConnector(dsn)
}

// KDBDriverWrapper wraps the KDB driver to implement sql.Driver interface
// This implementation uses the standard PostgreSQL driver directly to avoid
// any modification of the original DSN by the kingbase package
type KDBDriverWrapper struct{}

func (d KDBDriverWrapper) Open(dsn string) (driver.Conn, error) {
	// Use the standard PostgreSQL driver directly to ensure DSN integrity
	return pq.Driver{}.Open(dsn)
}

func (d KDBDriverWrapper) OpenConnector(dsn string) (driver.Connector, error) {
	// Use the standard PostgreSQL driver directly to ensure DSN integrity
	return pq.NewConnector(dsn)
}

func init() {
	sql.Register("DM8", &DM8DriverWrapper{})
	sql.Register("KDB", &KDBDriverWrapper{})
}
