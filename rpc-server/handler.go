package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/go-redis/redis"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct {
	redisClient *redis.Client
}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	resp := rpc.NewSendResponse()
	log.Printf("req %v %v %v %v", req.GetMessage().GetChat(), req.GetMessage().GetSender(),
		req.GetMessage().GetSendTime(), req.GetMessage().GetText())

	// redisErr := s.redisClient.Set(req.GetMessage().GetChat(), req.GetMessage().GetText(), 0).Err()
	b, _ := json.Marshal(req.GetMessage())
	redisErr := s.redisClient.LPush(req.GetMessage().GetChat(), string(b)).Err()
	if redisErr != nil {
		log.Println("RPC Send, set redis fail, ", redisErr)
		resp.Code = 500
		return nil, redisErr
	}

	resp.Code = 0
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	resp := rpc.NewPullResponse()

	log.Println("Chat", req.GetChat())

	// textVal, redisErr := s.redisClient.Get(req.GetChat()).Result()
	dbvalJsons, redisErr := s.redisClient.LRange(req.GetChat(), req.GetCursor(), int64(req.GetLimit())).Result()
	var dbvals []rpc.Message
	for _, dbValJson := range dbvalJsons {
		dbval := rpc.Message{}
		err := json.Unmarshal([]byte(dbValJson), &dbval)
		if err != nil {
			panic(fmt.Sprintf("amber boo %v", err))
		}
		log.Println("dbval", dbval)

		dbvals = append(dbvals, dbval)
	}
	log.Println("dbvals", dbvals)

	if redisErr != nil {
		log.Println("RPC Pull, get redis fail, ", redisErr)
		resp.Code = 500
		return nil, redisErr
	}
	log.Println("Returned text from redis, ", dbvals)

	resp.Code = 0

	resp.Messages = make([]*rpc.Message, int(math.Min(float64(req.GetLimit())+1, float64(len(dbvals)))))
	// For Loop to insert data
	for i := 0; i < len(dbvals) && i < int(req.GetLimit()+1); i++ {
		// resp.Messages[i] = &rpc.Message{}
		resp.Messages[i] = &dbvals[i]
	}

	return resp, nil
}

func areYouLuckySend(req *rpc.SendRequest) (int32, string) {
	hostname, _ := os.Hostname()

	if rand.Int31n(2) == 1 {
		return 0, "success " + hostname
	} else {
		return 500, "oops " + hostname
	}
}

func areYouLuckyPull() (int32, string) {
	hostname, _ := os.Hostname()
	if rand.Int31n(2) == 1 {
		return 0, "success " + hostname
	} else {
		return 500, "oops " + hostname
	}
}
