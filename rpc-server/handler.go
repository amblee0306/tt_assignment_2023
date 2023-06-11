package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/datastorage"
	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct {
	db datastorage.IMDataStorage
}

func NewIMServiceImpl(db datastorage.IMDataStorage) *IMServiceImpl {
	return &IMServiceImpl{
		db: db,
	}
}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	resp := rpc.NewSendResponse()
	log.Printf("req %v %v %v %v\n", req.GetMessage().GetChat(), req.GetMessage().GetSender(),
		req.GetMessage().GetSendTime(), req.GetMessage().GetText())

	b, _ := json.Marshal(req.GetMessage())
	err := s.db.AddMessage(ctx, req.GetMessage().GetChat(), string(b), req.GetMessage().GetSendTime())
	if err != nil {
		log.Println("Error: RPC Send, set redis fail, ", err)
		resp.Code = 500
		return resp, err
	}

	resp.Code = 0
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	resp := rpc.NewPullResponse()
	limit := req.GetLimit()
	if limit > 100 {
		limit = 100
	}

	dbvalJsons, nextCursor, err := s.db.GetChat(ctx, req.GetChat(), req.GetCursor(), limit, req.GetReverse())
	if err != nil {
		log.Println("Error: RPC Pull, get redis fail, ", err)
		resp.Code = 500
		resp.Msg = "Cannot get data"
		return resp, err
	}

	var dbvals []rpc.Message
	for _, dbValJson := range dbvalJsons {
		dbval := rpc.Message{}
		err := json.Unmarshal([]byte(dbValJson), &dbval)
		if err != nil {
			log.Println("Error: Cannot unmarshal, corrupted data", err)
			resp.Code = 500
			resp.Msg = "Corrupted data"
			return resp, err
		}
		dbvals = append(dbvals, dbval)
	}
	// log.Println("Returned text from redis, ", dbvals)

	hasMore := nextCursor != 0
	resp.Code = 0
	resp.NextCursor = &nextCursor
	resp.HasMore = &hasMore
	resp.Messages = make([]*rpc.Message, len(dbvals))
	// For Loop to insert data
	for i := 0; i < len(dbvals); i++ {
		// resp.Messages[i] = &rpc.Message{}
		resp.Messages[i] = &dbvals[i]
	}

	return resp, nil
}
