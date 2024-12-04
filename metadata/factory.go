package metadata

import (
	"fmt"
)

// DBType represents supported database types
type DBType string

const (
	MySQL      DBType = "mysql"
	Oracle     DBType = "oracle"
	PostgreSQL DBType = "postgresql"
	SQLServer  DBType = "sqlserver"
)

// NewMetadataProvider creates a new metadata provider based on database type
func NewMetadataProvider(dbType DBType) (Provider, error) {
	// switch dbType {
	// case MySQL:
	// 	return mysql.NewMySQLProvider(), nil
	// case Oracle:
	// 	return oracle.NewOracleProvider(), nil
	// case PostgreSQL:
	// 	return postgres.NewPostgresProvider(), nil
	// case SQLServer:
	// 	return sqlserver.NewSQLServerProvider(), nil
	// default:
	// 	return nil, fmt.Errorf("unsupported database type: %s", dbType)
	// }
	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}
