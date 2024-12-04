package metadata

import (
	"context"
	"sql_translator/models"
)

// Provider defines the interface for retrieving database object metadata
type Provider interface {
	GetTableMetadata(ctx context.Context, identifier *models.DBObject) (*models.TableMetadata, error)
	GetProcedureMetadata(ctx context.Context, identifier *models.DBObject) (*models.ProcedureMetadata, error)
}
