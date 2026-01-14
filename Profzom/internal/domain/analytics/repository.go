package analytics

import "context"

type Repository interface {
	Create(ctx context.Context, event Event) error
}
