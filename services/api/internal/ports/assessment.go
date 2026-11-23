package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
)

// AssessmentRepository stores immutable assessment results.
type AssessmentRepository interface {
	// Create stores the assessment and all its findings atomically.
	Create(ctx context.Context, a compliance.Assessment, createdBy uuid.UUID) (compliance.Assessment, error)
	// List returns summaries (Findings empty), newest first.
	List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]compliance.Assessment, error)
	// ListForAsset returns summaries for every snapshot of one asset.
	ListForAsset(ctx context.Context, workspaceID, assetID uuid.UUID, limit int) ([]compliance.Assessment, error)
	Get(ctx context.Context, workspaceID, id uuid.UUID) (compliance.Assessment, error)
}
