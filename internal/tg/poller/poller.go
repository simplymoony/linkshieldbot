// Package poller implements a facility to poll and retrieve updates.
package poller

import (
	"context"
	"errors"
	"time"

	"github.com/simplymoony/linkshieldbot/internal/tg"
)

type (
	UpdateCallback func(context.Context, *tg.Bot, *tg.Update) error
	ErrorCallback  func(error, bool)
)

// Poll polls and retrieves updates for the provided bot.
// The updateCb callback is called for each retrieved update, whereas the errorCb callback is called
// for any non-fatal errors happening.
func Poll(
	ctx context.Context, bot *tg.Bot,
	updateTimeout time.Duration, handlerTimeout time.Duration,
	updateCb UpdateCallback, errorCb ErrorCallback,
) error {
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			uctx, uctxStop := context.WithTimeout(ctx, updateTimeout)
			defer uctxStop()

			updates, err := bot.GetUpdates(uctx, offset, 1)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				errorCb(err, true)
				time.Sleep(1 * time.Second)
				continue
			}

			count := len(updates)
			if count == 0 {
				continue
			}

			for _, update := range updates {
				go func(update *tg.Update) {
					hctx, hctxStop := context.WithTimeout(ctx, handlerTimeout)
					defer hctxStop()

					if err := updateCb(hctx, bot, update); err != nil {
						errorCb(err, false)
					}
				}(update)
			}

			offset = updates[count-1].UpdateID + 1
		}
	}
}
