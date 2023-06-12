package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/datastorage"
	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func newTestRedis() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client
}

func TestIMServiceImpl_Send(t *testing.T) {
	type args struct {
		ctx context.Context
		req *rpc.SendRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: &rpc.Message{},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewIMServiceImpl(&mockIMDataStorage{})
			got, err := s.Send(tt.args.ctx, tt.args.req)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.NotNil(t, got)
		})
	}
}

type mockIMDataStorage struct {
	dataObjects []string
	dataError   error
}

var _ datastorage.IMDataStorage = &mockIMDataStorage{}

func (m *mockIMDataStorage) AddMessage(ctx context.Context, chatroom string, message string, timestamp int64) error {
	return nil
}

func (m *mockIMDataStorage) GetChat(ctx context.Context, chat string, cursor int64, limit int32, isAsc bool) ([]string, int64, error) {
	return m.dataObjects, 0, m.dataError
}

func TestIMServiceImpl_SendPull(t *testing.T) {
	firstMsg := rpc.Message{
		Chat:     "chatroom12345",
		Text:     "first message",
		Sender:   "u1",
		SendTime: 1686283200,
	}
	secondMsg := rpc.Message{
		Chat:     "chatroom12345",
		Text:     "second message",
		Sender:   "u2",
		SendTime: 1686283205,
	}

	varTrue := true
	varFalse := false
	var zero int64 = 0

	redisClient := newTestRedis()
	// put data into redis else no data in DB to pull and test.
	secondMsgB, _ := json.Marshal(secondMsg)
	firstMsgB, _ := json.Marshal(firstMsg)
	z2 := redis.Z{
		Score:  float64(secondMsg.SendTime),
		Member: string(secondMsgB),
	}
	z1 := redis.Z{
		Score:  float64(firstMsg.SendTime),
		Member: string(firstMsgB),
	}
	redisClient.ZAdd(secondMsg.Chat, z2).Err()
	redisClient.ZAdd(firstMsg.Chat, z1).Err()

	type fields struct {
		// mockClient datastorage.IMDataStorage
		redisClient *redis.Client
	}
	type args struct {
		req *rpc.PullRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *rpc.PullResponse
		wantErr error
	}{
		{
			name:   "Basic case - return default 10 messages.",
			fields: fields{redisClient},
			args:   args{req: &rpc.PullRequest{Chat: "chatroom12345", Cursor: 0, Limit: -1, Reverse: &varTrue}},
			want: &rpc.PullResponse{
				Code: 0,
				Messages: []*rpc.Message{
					&firstMsg,
					&secondMsg,
				},
				HasMore:    &varFalse,
				NextCursor: &zero,
			},
			wantErr: nil,
		},
		{
			name:   "Pull test, reverse order.",
			fields: fields{redisClient},
			args:   args{req: &rpc.PullRequest{Chat: "chatroom12345", Cursor: 0, Limit: 1, Reverse: &varFalse}},
			want: &rpc.PullResponse{
				Code: 0,
				Messages: []*rpc.Message{
					&secondMsg,
				},
				HasMore:    &varTrue,
				NextCursor: &secondMsg.SendTime,
			},
			wantErr: nil,
		},
		{
			name:   "Simple pull test, limit too high.",
			fields: fields{redisClient},
			args:   args{req: &rpc.PullRequest{Chat: "chatroom12345", Cursor: 0, Limit: 1000, Reverse: &varTrue}},
			want: &rpc.PullResponse{
				Code: 0,
				Messages: []*rpc.Message{
					&firstMsg,
					&secondMsg,
				},
				HasMore:    &varFalse,
				NextCursor: &zero,
			},
			wantErr: nil,
		},
		{
			name:   "Empty chatroom.",
			fields: fields{redisClient},
			args:   args{req: &rpc.PullRequest{Chat: "chatroom88", Cursor: 0, Limit: -1, Reverse: &varTrue}},
			want: &rpc.PullResponse{
				Code:       0,
				Messages:   []*rpc.Message{},
				HasMore:    &varFalse,
				NextCursor: &zero,
			},
			wantErr: nil,
		},
		{
			name:   "Test cursor.",
			fields: fields{redisClient},
			args:   args{req: &rpc.PullRequest{Chat: "chatroom12345", Cursor: secondMsg.SendTime, Limit: -1, Reverse: &varFalse}},
			want: &rpc.PullResponse{
				Code: 0,
				Messages: []*rpc.Message{
					&firstMsg,
				},
				HasMore:    &varFalse,
				NextCursor: &zero,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewIMServiceImpl(datastorage.NewRedisDA(redisClient))
			got, err := s.Pull(context.Background(), tt.args.req)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.True(t, cmp.Equal(tt.want, got))
			assert.Equal(t, tt.want, got)
		})
	}
}
