package initiator

import (
	"context"
	"log/slog"

	"yadro.com/course/search/core"
)

func RunEventBasedIndexUpdate(ctx context.Context, searcher core.Searcher, broker core.Broker, log *slog.Logger) {
	if err := searcher.BuildIndex(ctx); err != nil {
		log.Error("failed to build index on start", "error", err)
	}

	if err := broker.Subscribe("xkcd.db.updated", func(data []byte) {
		log.Info("received db update event, rebuilding index", "data", string(data))
		if err := searcher.BuildIndex(ctx); err != nil {
			log.Error("failed to rebuild index after db update", "error", err)
		}
	}); err != nil {
		log.Error("failed to subscribe to db update events", "error", err)
	}

	if err := broker.Subscribe("xkcd.db.dropped", func(data []byte) {
		log.Info("received db drop event, clearing index", "data", string(data))
		searcher.ClearIndex()
	}); err != nil {
		log.Error("failed to subscribe to db drop events", "error", err)
	}
}
