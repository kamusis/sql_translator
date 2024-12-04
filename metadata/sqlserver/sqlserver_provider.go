package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"sql_translator/metadata/base"
	"sql_translator/models"
)

type SQLServerMetadataProvider struct {
	base.BaseMetadataProvider
	db *sql.DB
}

func NewSQLServerProvider(db *sql.DB) *SQLServerMetadataProvider {
	return &SQLServerMetadataProvider{db: db}
}

func (p *SQLServerMetadataProvider) GetTableMetadata(ctx context.Context, identifier *models.DBObject) (*models.TableMetadata, error) {
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
			c.name AS column_name,
			t.name AS data_type,
			CASE
				WHEN t.name IN ('nchar', 'nvarchar') AND c.max_length != -1 THEN c.max_length/2
				WHEN t.name IN ('char', 'varchar') AND c.max_length != -1 THEN c.max_length
				ELSE NULL
			END AS char_length,
			c.precision AS numeric_precision,
			c.scale AS numeric_scale,
			c.is_nullable,
			object_definition(c.default_object_id) AS column_default,
			c.is_identity,
			c.is_computed
		FROM sys.columns c
		INNER JOIN sys.types t ON c.user_type_id = t.user_type_id
		INNER JOIN sys.objects o ON c.object_id = o.object_id
		INNER JOIN sys.schemas s ON o.schema_id = s.schema_id
		WHERE s.name = @p1 AND o.name = @p2
		ORDER BY c.column_id`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query column metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col models.ColumnMetadata
		var length, precision, scale sql.NullInt64
		var defaultValue sql.NullString

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&length,
			&precision,
			&scale,
			// &col.IsNullable,
			&defaultValue,
			// &col.IsIdentity,
			// &col.IsComputed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column metadata: %v", err)
		}

		col.Length = int(length.Int64)
		// col.Precision = int(precision.Int64)
		// col.Scale = int(scale.Int64)
		// col.DefaultValue = defaultValue.String

		metadata.Columns = append(metadata.Columns, col)
	}

	// Query primary keys
	pkQuery := `
		SELECT c.name AS column_name
		FROM sys.index_columns ic
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.indexes i ON ic.object_id = i.object_id AND ic.index_id = i.index_id
		INNER JOIN sys.objects o ON i.object_id = o.object_id
		INNER JOIN sys.schemas s ON o.schema_id = s.schema_id
		WHERE i.is_primary_key = 1
			AND s.name = @p1
			AND o.name = @p2
		ORDER BY ic.key_ordinal`

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

func (p *SQLServerMetadataProvider) GetProcedureMetadata(ctx context.Context, identifier *models.DBObject) (*models.ProcedureMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	metadata := &models.ProcedureMetadata{
		Schema: identifier.Owner,
		Name:   identifier.Name,
	}

	query := `
		SELECT
			p.name AS parameter_name,
			CASE p.is_output
				WHEN 1 THEN 'OUT'
				ELSE 'IN'
			END AS parameter_mode,
			t.name AS data_type,
			CASE
				WHEN t.name IN ('nchar', 'nvarchar') AND p.max_length != -1 THEN p.max_length/2
				WHEN t.name IN ('char', 'varchar') AND p.max_length != -1 THEN p.max_length
				ELSE NULL
			END AS char_length,
			p.precision AS numeric_precision,
			p.scale AS numeric_scale,
			p.has_default_value,
			ISNULL(object_definition(p.default_object_id), '') AS default_value
		FROM sys.parameters p
		INNER JOIN sys.types t ON p.user_type_id = t.user_type_id
		INNER JOIN sys.objects o ON p.object_id = o.object_id
		INNER JOIN sys.schemas s ON o.schema_id = s.schema_id
		WHERE s.name = @p1
			AND o.name = @p2
		ORDER BY p.parameter_id`

	rows, err := p.db.QueryContext(ctx, query, identifier.Owner, identifier.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query procedure metadata: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var param models.ParameterMetadata
		var length, precision, scale sql.NullInt64
		var paramName, paramMode string

		err := rows.Scan(
			&paramName,
			&paramMode,
			&param.DataType,
			&length,
			&precision,
			&scale,
			// &param.HasDefault,
			// &param.DefaultValue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parameter metadata: %v", err)
		}

		param.Name = paramName
		param.Direction = paramMode
		// param.Length = int(length.Int64)
		// param.Precision = int(precision.Int64)
		// param.Scale = int(scale.Int64)

		metadata.Parameters = append(metadata.Parameters, param)
	}

	return metadata, nil
}
