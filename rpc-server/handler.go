package main

import (
	"context"
	"log"
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

	redisErr := s.redisClient.Set(req.GetMessage().GetChat(), req.GetMessage().GetText(), 0).Err()
	if redisErr != nil {
		log.Println("RPC Send, set redis fail, ", redisErr)
		return nil, redisErr
	}

	resp.Code, resp.Msg = areYouLuckySend(req)
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	resp := rpc.NewPullResponse()

	log.Println("Chat", req.GetChat())
	textVal, redisErr := s.redisClient.Get(req.GetChat()).Result()
	if redisErr != nil {
		log.Println("RPC Pull, get redis fail, ", redisErr)
		return nil, redisErr
	}
	log.Println("Returned text from redis, ", textVal)

	resp.Code, resp.Msg = areYouLuckyPull()
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
