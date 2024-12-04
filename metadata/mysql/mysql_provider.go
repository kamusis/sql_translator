package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sql_translator/metadata/base"
	"sql_translator/models"
)

type MySQLMetadataProvider struct {
	base.BaseMetadataProvider
	db *sql.DB
}

func NewMySQLProvider(db *sql.DB) *MySQLMetadataProvider {
	return &MySQLMetadataProvider{db: db}
}

func (p *MySQLMetadataProvider) GetTableMetadata(ctx context.Context, identifier *models.DBObject) (*models.TableMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	metadata := &models.TableMetadata{
		Schema: identifier.Owner,
		Name:   identifier.Name,
	}

	// Query column information
	query := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query column metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col models.ColumnMetadata
		var maxLength, precision, scale sql.NullInt64
		var isNullable, defaultValue, extra sql.NullString

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&maxLength,
			&precision,
			&scale,
			&isNullable,
			&defaultValue,
			&extra,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column metadata: %v", err)
		}

		col.Length = int(maxLength.Int64)
		// col.Precision = int(precision.Int64)
		// col.Scale = int(scale.Int64)
		// col.IsNullable = isNullable.String == "YES"
		// col.DefaultValue = defaultValue.String
		// col.IsIdentity = extra.String == "auto_increment"
		// col.IsComputed = extra.String == "VIRTUAL GENERATED" || extra.String == "STORED GENERATED"

		metadata.Columns = append(metadata.Columns, col)
	}

	// Query primary keys
	pkQuery := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
			AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION`

	pkRows, err := p.db.QueryContext(ctx, pkQuery, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary key metadata: %v", err)
	}
	defer pkRows.Close()

	for pkRows.Next() {
		var pkColumn string
		if err := pkRows.Scan(&pkColumn); err != nil {
			return nil, fmt.Errorf("failed to scan primary key metadata: %v", err)
		}
		metadata.PrimaryKey = append(metadata.PrimaryKey, pkColumn)
	}

	return metadata, nil
}

func (p *MySQLMetadataProvider) GetProcedureMetadata(ctx context.Context, identifier *models.DBObject) (*models.ProcedureMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	metadata := &models.ProcedureMetadata{
		Schema: identifier.Owner,
		Name:   identifier.Name,
	}

	query := `
		SELECT
			PARAMETER_NAME,
			PARAMETER_MODE,
			DATA_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE
		FROM INFORMATION_SCHEMA.PARAMETERS
		WHERE SPECIFIC_SCHEMA = ? AND SPECIFIC_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query procedure metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var param models.ParameterMetadata
		var maxLength, precision, scale sql.NullInt64
		var paramName, paramMode sql.NullString

		err := rows.Scan(
			&paramName,
			&paramMode,
			&param.DataType,
			&maxLength,
			&precision,
			&scale,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parameter metadata: %v", err)
		}

		param.Name = paramName.String
		param.Direction = paramMode.String
		// param.Length = int(maxLength.Int64)
		// param.Precision = int(precision.Int64)
		// param.Scale = int(scale.Int64)

		metadata.Parameters = append(metadata.Parameters, param)
	}

	return metadata, nil
}
