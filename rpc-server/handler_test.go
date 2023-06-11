package main

import (
	"context"
	"errors"
	"testing"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/datastorage"
	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/stretchr/testify/assert"
)

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
			s := NewIMServiceImpl(&mockClient{})
			got, err := s.Send(tt.args.ctx, tt.args.req)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.NotNil(t, got)
		})
	}
}

type mockClient struct {
	dataObjects []string
	dataError   error
}

var _ datastorage.IMDataStorage = &mockClient{}

func (m *mockClient) AddMessage(ctx context.Context, chatroom string, message string, timestamp int64) error {
	return nil
}

func (m *mockClient) GetChat(ctx context.Context, chat string, cursor int64, limit int32, isAsc bool) ([]string, int64, error) {
	return m.dataObjects, 0, m.dataError
}
