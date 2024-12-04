package models

import (
	"fmt"
	"strings"
)

// TableMetadata contains table structure information
type TableMetadata struct {
	Name       string
	Schema     string
	Owner      string
	Columns    []ColumnMetadata
	PrimaryKey []string
	Indexes    []IndexMetadata
}

// ColumnMetadata contains column information
type ColumnMetadata struct {
	Name     string
	DataType string
	Length   int
	Nullable bool
	Default  string
}

// IndexMetadata contains index information
type IndexMetadata struct {
	Name        string
	Columns     []string
	IsUnique    bool
	IsClustered bool
}

// ProcedureMetadata contains stored procedure information
type ProcedureMetadata struct {
	Name       string
	Schema     string
	Owner      string
	Parameters []ParameterMetadata
	ReturnType string
	Body       string
}

// ParameterMetadata contains procedure parameter information
type ParameterMetadata struct {
	Name      string
	DataType  string
	Direction string
	Default   string
}

// MetadataContext contains metadata information for translation
type MetadataContext struct {
	Tables     map[string]*TableMetadata
	Procedures map[string]*ProcedureMetadata
}

// String returns a string representation of the metadata context
func (m *MetadataContext) String() string {
	var result strings.Builder

	// Add table metadata
	result.WriteString("Tables:\n")
	for name, table := range m.Tables {
		result.WriteString(fmt.Sprintf("  %s:\n", name))
		result.WriteString("    Columns:\n")
		for _, col := range table.Columns {
			result.WriteString(fmt.Sprintf("      %s %s (nullable: %v, default: %s)\n",
				col.Name, col.DataType, col.Nullable, col.Default))
		}
		if len(table.PrimaryKey) > 0 {
			result.WriteString(fmt.Sprintf("    Primary Key: %s\n",
				strings.Join(table.PrimaryKey, ", ")))
		}
		for _, idx := range table.Indexes {
			result.WriteString(fmt.Sprintf("    Index %s: %s (unique: %v)\n",
				idx.Name, strings.Join(idx.Columns, ", "), idx.IsUnique))
		}
	}

	// Add procedure metadata
	result.WriteString("\nProcedures:\n")
	for name, proc := range m.Procedures {
		result.WriteString(fmt.Sprintf("  %s:\n", name))
		result.WriteString("    Parameters:\n")
		for _, param := range proc.Parameters {
			result.WriteString(fmt.Sprintf("      %s %s %s (default: %s)\n",
				param.Name, param.Direction, param.DataType, param.Default))
		}
		if proc.ReturnType != "" {
			result.WriteString(fmt.Sprintf("    Returns: %s\n", proc.ReturnType))
		}
	}

	return result.String()
}
