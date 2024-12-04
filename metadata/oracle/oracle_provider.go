package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"sql_translator/metadata/base"
	"sql_translator/models"
)

type OracleMetadataProvider struct {
	base.BaseMetadataProvider
	db *sql.DB
}

func NewOracleProvider(db *sql.DB) *OracleMetadataProvider {
	return &OracleMetadataProvider{db: db}
}

func (p *OracleMetadataProvider) GetTableMetadata(ctx context.Context, identifier *models.DBObject) (*models.TableMetadata, error) {
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
			DATA_LENGTH,
			DATA_PRECISION,
			DATA_SCALE,
			NULLABLE,
			DATA_DEFAULT,
			IDENTITY_COLUMN,
			VIRTUAL_COLUMN
		FROM ALL_TAB_COLUMNS
		WHERE OWNER = :1 AND TABLE_NAME = :2
		ORDER BY COLUMN_ID`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query column metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col models.ColumnMetadata
		var length, precision, scale sql.NullInt64
		var nullable, defaultValue, identityColumn, virtualColumn sql.NullString

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&length,
			&precision,
			&scale,
			&nullable,
			&defaultValue,
			&identityColumn,
			&virtualColumn,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column metadata: %v", err)
		}

		col.Length = int(length.Int64)
		// col.Precision = int(precision.Int64)
		// col.Scale = int(scale.Int64)
		// col.IsNullable = nullable.String == "Y"
		// col.DefaultValue = defaultValue.String
		// col.IsIdentity = identityColumn.String == "YES"
		// col.IsComputed = virtualColumn.String == "YES"

		metadata.Columns = append(metadata.Columns, col)
	}

	// Query primary keys
	pkQuery := `
		SELECT COLUMN_NAME
		FROM ALL_CONS_COLUMNS
		WHERE OWNER = :1
			AND TABLE_NAME = :2
			AND CONSTRAINT_NAME IN (
				SELECT CONSTRAINT_NAME
				FROM ALL_CONSTRAINTS
				WHERE OWNER = :1
					AND TABLE_NAME = :2
					AND CONSTRAINT_TYPE = 'P'
			)
		ORDER BY POSITION`

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

func (p *OracleMetadataProvider) GetProcedureMetadata(ctx context.Context, identifier *models.DBObject) (*models.ProcedureMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	metadata := &models.ProcedureMetadata{
		Schema: identifier.Owner,
		Name:   identifier.Name,
	}

	query := `
		SELECT
			ARGUMENT_NAME,
			IN_OUT,
			DATA_TYPE,
			DATA_LENGTH,
			DATA_PRECISION,
			DATA_SCALE,
			DEFAULTED
		FROM ALL_ARGUMENTS
		WHERE OWNER = :1
			AND OBJECT_NAME = :2
			AND PACKAGE_NAME IS NULL
		ORDER BY POSITION`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query procedure metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var param models.ParameterMetadata
		var length, precision, scale sql.NullInt64
		var paramName, inOut, defaulted sql.NullString

		err := rows.Scan(
			&paramName,
			&inOut,
			&param.DataType,
			&length,
			&precision,
			&scale,
			&defaulted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parameter metadata: %v", err)
		}

		param.Name = paramName.String
		param.Direction = inOut.String
		// param.Length = int(length.Int64)
		// param.Precision = int(precision.Int64)
		// param.Scale = int(scale.Int64)
		// param.HasDefault = defaulted.String == "Y"

		metadata.Parameters = append(metadata.Parameters, param)
	}

	return metadata, nil
}
