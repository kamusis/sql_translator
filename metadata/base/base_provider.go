package base

import (
	"context"
	"sql_translator/models"
)

// BaseMetadataProvider provides default empty implementations
type BaseMetadataProvider struct{}

// GetTableMetadata returns empty table metadata
func (p *BaseMetadataProvider) GetTableMetadata(ctx context.Context, identifier *models.DBObject) (*models.TableMetadata, error) {
	return &models.TableMetadata{}, nil
}

// GetProcedurceMetadata returns empty procedure metadata
func (p *BaseMetadataProvider) GetProcedureMetadata(ctx context.Context, identifier *models.DBObject) (*models.ProcedureMetadata, error) {
	return &models.ProcedureMetadata{}, nil
}
