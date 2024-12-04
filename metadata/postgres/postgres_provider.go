package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sql_translator/metadata/base"
	"sql_translator/models"
)

type PostgresMetadataProvider struct {
	base.BaseMetadataProvider
	db *sql.DB
}

func NewPostgresProvider(db *sql.DB) *PostgresMetadataProvider {
	return &PostgresMetadataProvider{db: db}
}

func (p *PostgresMetadataProvider) GetTableMetadata(ctx context.Context, identifier *models.DBObject) (*models.TableMetadata, error) {
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
			a.attname as column_name,
			t.typname as data_type,
			CASE
				WHEN t.typname = 'varchar' THEN a.atttypmod - 4
				WHEN t.typname = 'bpchar' THEN a.atttypmod - 4
				ELSE NULL
			END as char_length,
			CASE
				WHEN t.typname IN ('numeric', 'decimal') THEN ((a.atttypmod - 4) >> 16) & 65535
				ELSE NULL
			END as numeric_precision,
			CASE
				WHEN t.typname IN ('numeric', 'decimal') THEN (a.atttypmod - 4) & 65535
				ELSE NULL
			END as numeric_scale,
			NOT a.attnotnull as is_nullable,
			pg_get_expr(d.adbin, d.adrelid) as column_default,
			CASE WHEN a.attidentity != '' THEN true ELSE false END as is_identity,
			CASE WHEN a.attgenerated != '' THEN true ELSE false END as is_generated
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_type t ON t.oid = a.atttypid
		LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		WHERE n.nspname = $1
			AND c.relname = $2
			AND a.attnum > 0
			AND NOT a.attisdropped
		ORDER BY a.attnum`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query column metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col models.ColumnMetadata
		var length, precision, scale sql.NullInt64
		var defaultValue sql.NullString
		var isIdentity, isGenerated bool

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&length,
			&precision,
			&scale,
			// &IsNullable,
			&defaultValue,
			&isIdentity,
			&isGenerated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column metadata: %v", err)
		}

		col.Length = int(length.Int64)
		// col.Precision = int(precision.Int64)
		// col.Scale = int(scale.Int64)
		// col.DefaultValue = defaultValue.String
		// col.IsIdentity = isIdentity
		// col.IsComputed = isGenerated

		metadata.Columns = append(metadata.Columns, col)
	}

	// Query primary keys
	pkQuery := `
		SELECT a.attname as column_name
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE i.indisprimary
			AND n.nspname = $1
			AND c.relname = $2
		ORDER BY array_position(i.indkey, a.attnum)`

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

func (p *PostgresMetadataProvider) GetProcedureMetadata(ctx context.Context, identifier *models.DBObject) (*models.ProcedureMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	metadata := &models.ProcedureMetadata{
		Schema: identifier.Owner,
		Name:   identifier.Name,
	}

	query := `
		SELECT
			p.parameter_name,
			p.parameter_mode,
			p.data_type,
			CASE
				WHEN p.data_type = 'character varying' THEN p.character_maximum_length
				WHEN p.data_type = 'character' THEN p.character_maximum_length
				ELSE NULL
			END as char_length,
			p.numeric_precision,
			p.numeric_scale,
			p.parameter_default IS NOT NULL as has_default,
			COALESCE(p.parameter_default, '') as default_value
		FROM information_schema.parameters p
		WHERE p.specific_schema = $1
			AND p.specific_name = $2
		ORDER BY p.ordinal_position`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query procedure metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var param models.ParameterMetadata
		var length, precision, scale sql.NullInt64
		var paramName, paramMode sql.NullString
		var defaultValue string

		err := rows.Scan(
			&paramName,
			&paramMode,
			&param.DataType,
			&length,
			&precision,
			&scale,
			// &param.HasDefault,
			&defaultValue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parameter metadata: %v", err)
		}

		param.Name = paramName.String
		param.Direction = paramMode.String
		// param.Length = int(length.Int64)
		// param.Precision = int(precision.Int64)
		// param.Scale = int(scale.Int64)
		// param.DefaultValue = defaultValue

		metadata.Parameters = append(metadata.Parameters, param)
	}

	return metadata, nil
}
