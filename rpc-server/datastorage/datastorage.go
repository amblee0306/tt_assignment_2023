package datastorage

import "context"

type IMDataStorage interface {
	GetChat(ctx context.Context, chat string, cursor int64, limit int32, isAsc bool) ([]string, int64, error)
	AddMessage(ctx context.Context, chatroom string, message string, timestamp int64) error
}
