package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/go-redis/redis"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct {
	redisClient *redis.Client
}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	resp := rpc.NewSendResponse()
	log.Printf("req %v %v %v %v\n", req.GetMessage().GetChat(), req.GetMessage().GetSender(),
		req.GetMessage().GetSendTime(), req.GetMessage().GetText())

	b, _ := json.Marshal(req.GetMessage())
	err := s.AddMessage(ctx, req.GetMessage().GetChat(), string(b), req.GetMessage().GetSendTime())
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

	dbvalJsons, nextCursor, err := s.GetChat(ctx, req.GetChat(), req.GetCursor(), limit, req.GetReverse())
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

func (s *IMServiceImpl) AddMessage(ctx context.Context, chatroom string, message string, timestamp int64) error {
	z := redis.Z{
		Score:  float64(timestamp),
		Member: message,
	}
	return s.redisClient.ZAdd(chatroom, z).Err()
}

// GetChat will return the list of raw data of the message.
// Since the cursor will be the last item in the previous pagination, it will exclude the message at the cursor,
func (s *IMServiceImpl) GetChat(ctx context.Context, chat string, cursor int64, limit int32, isAsc bool) ([]string, int64, error) {
	var zResults []redis.Z
	var err error
	if isAsc {
		zResults, err = s.getRedisZRangeByScoreWithScores(chat, cursor, limit)
	} else {
		zResults, err = s.getRedisZRevRangeByScoreWithScores(chat, cursor, limit)
	}
	var lastCursor int64
	if len(zResults) > 0 {
		lastCursor = int64(zResults[len(zResults)-1].Score)
	} else {
		lastCursor = 0
	}

	results := make([]string, 0, len(zResults))
	for _, r := range zResults {
		// log.Printf("score %v, member %v\n", r.Score, r.Member)
		// log.Printf("type %T", r.Member)
		results = append(results, fmt.Sprintf("%v", r.Member))
	}
	// log.Println("lastCursor", lastCursor)

	return results, lastCursor, err
}

func (s *IMServiceImpl) getRedisZRangeByScoreWithScores(chat string, cursor int64, limit int32) ([]redis.Z, error) {
	var min string
	if cursor == 0 {
		min = "-inf"
	} else {
		min = "(" + strconv.FormatInt(cursor, 10) // exclude the message at the cursor
	}
	zrangeBy := redis.ZRangeBy{
		Min:    min,
		Max:    "+inf",
		Offset: 0,
		Count:  int64(limit),
	}
	return s.redisClient.ZRangeByScoreWithScores(chat, zrangeBy).Result()
}

func (s *IMServiceImpl) getRedisZRevRangeByScoreWithScores(chat string, cursor int64, limit int32) ([]redis.Z, error) {
	var max string
	if cursor == 0 {
		max = "+inf"
	} else {
		max = "(" + strconv.FormatInt(cursor, 10) // exclude the message at the cursor
	}
	zrangeBy := redis.ZRangeBy{
		Min:    "-inf",
		Max:    max,
		Offset: 0,
		Count:  int64(limit),
	}
	return s.redisClient.ZRevRangeByScoreWithScores(chat, zrangeBy).Result()
}
