package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/benwiebe/udb-core/internal/display"
	"github.com/benwiebe/udb-core/internal/plugins"
	"github.com/benwiebe/udb-plugin-library/types"
)

// Run cycles through boards in sequence until ctx is cancelled.
func Run(ctx context.Context, d display.Display, boards []plugins.BoardEntry) {
	for ctx.Err() == nil {
		for _, entry := range boards {
			if ctx.Err() != nil {
				break
			}
			runBoard(ctx, d, entry)
		}
	}
}

// runBoard dispatches to the correct render path based on board type and holds
// for the configured duration before returning.
func runBoard(ctx context.Context, d display.Display, entry plugins.BoardEntry) {
	duration := time.Duration(entry.Config.DurationSeconds) * time.Second

	switch entry.Board.GetType() {
	case types.BoardTypeStatic:
		staticBoard, ok := entry.Board.(types.StaticBoard[any])
		if !ok {
			fmt.Printf("Warning: board %s declared static but doesn't implement StaticBoard\n", entry.Config.BoardId)
			return
		}
		if err := d.Render(staticBoard.Render()); err != nil {
			fmt.Printf("Render error (board %s): %v\n", entry.Config.BoardId, err)
		}
		if duration > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(duration):
			}
		}

	case types.BoardTypeAnimated:
		animBoard, ok := entry.Board.(types.AnimatedBoard[any])
		if !ok {
			fmt.Printf("Warning: board %s declared animated but doesn't implement AnimatedBoard\n", entry.Config.BoardId)
			return
		}
		frames := animBoard.Render()
		deadline := time.Now().Add(duration)
		for ctx.Err() == nil && (duration == 0 || time.Now().Before(deadline)) {
			for _, frame := range frames {
				if ctx.Err() != nil || (duration > 0 && !time.Now().Before(deadline)) {
					break
				}
				if err := d.Render(frame.Img); err != nil {
					fmt.Printf("Render error (board %s): %v\n", entry.Config.BoardId, err)
				}
				sleep := frame.Duration
				if duration > 0 {
					if remaining := time.Until(deadline); remaining < sleep {
						sleep = remaining
					}
				}
				select {
				case <-ctx.Done():
				case <-time.After(sleep):
				}
			}
			if duration == 0 {
				break
			}
		}

	case types.BoardTypeDynamic:
		dynBoard, ok := entry.Board.(types.DynamicBoard[any])
		if !ok {
			fmt.Printf("Warning: board %s declared dynamic but doesn't implement DynamicBoard\n", entry.Config.BoardId)
			return
		}
		var changed <-chan struct{}
		if entry.Datasource != nil {
			changed = entry.Datasource.DataChanged()
		}
		deadline := time.Now().Add(duration)
		for ctx.Err() == nil && (duration == 0 || time.Now().Before(deadline)) {
			frame := dynBoard.Render()
			if err := d.Render(frame.Img); err != nil {
				fmt.Printf("Render error (board %s): %v\n", entry.Config.BoardId, err)
			}
			sleep := frame.Duration
			if duration > 0 {
				if remaining := time.Until(deadline); remaining < sleep {
					sleep = remaining
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(sleep):
			case <-changed: // nil channel blocks forever — natural no-op when datasource has no push notifications
			}
		}
	}
}
