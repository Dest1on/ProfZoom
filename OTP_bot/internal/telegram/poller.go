package telegram

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type UpdateHandler interface {
	HandleUpdate(ctx context.Context, update Update) error
}

type UpdatesClient interface {
	GetUpdates(ctx context.Context, offset int64, timeout time.Duration, limit int) ([]Update, error)
	DeleteWebhook(ctx context.Context, dropPending bool) error
}

type Poller struct {
	client      UpdatesClient
	handler     UpdateHandler
	logger      *slog.Logger
	timeout     time.Duration
	interval    time.Duration
	limit       int
	dropPending bool
	dropWebhook bool
}

func NewPoller(client UpdatesClient, handler UpdateHandler, logger *slog.Logger, timeout, interval time.Duration, limit int, dropPending, dropWebhook bool) *Poller {
	if logger == nil {
		logger = slog.Default()
	}
	if timeout <= 0 {
		timeout = 25 * time.Second
	}
	if interval <= 0 {
		interval = time.Second
	}
	if limit <= 0 {
		limit = 50
	}
	return &Poller{
		client:      client,
		handler:     handler,
		logger:      logger,
		timeout:     timeout,
		interval:    interval,
		limit:       limit,
		dropPending: dropPending,
		dropWebhook: dropWebhook,
	}
}

func (p *Poller) Run(ctx context.Context) {
	var offset int64
	if p.dropWebhook {
		if err := p.client.DeleteWebhook(ctx, p.dropPending); err != nil {
			p.logger.Warn("telegram delete webhook failed", slog.String("error", err.Error()))
		}
	}
	if p.dropPending {
		if latest, err := p.flushPending(ctx); err != nil {
			p.logger.Warn("telegram flush pending failed", slog.String("error", err.Error()))
		} else if latest > 0 {
			offset = latest + 1
		}
	}

	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := p.client.GetUpdates(ctx, offset, p.timeout, p.limit)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			p.logger.Warn("telegram polling failed", slog.String("error", err.Error()))
			p.sleep(ctx)
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if err := p.handler.HandleUpdate(ctx, update); err != nil {
				p.logger.Error("telegram update handling failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (p *Poller) flushPending(ctx context.Context) (int64, error) {
	updates, err := p.client.GetUpdates(ctx, 0, 0, p.limit)
	if err != nil {
		return 0, err
	}
	if len(updates) == 0 {
		return 0, nil
	}
	return updates[len(updates)-1].UpdateID, nil
}

func (p *Poller) sleep(ctx context.Context) {
	timer := time.NewTimer(p.interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
