package datastorage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

type redisDA struct {
	redisClient *redis.Client
}

var _ IMDataStorage = &redisDA{}

func NewRedisDA(redisClient *redis.Client) *redisDA {
	return &redisDA{
		redisClient: redisClient,
	}
}

func (s *redisDA) AddMessage(ctx context.Context, chatroom string, message string, timestamp int64) error {
	z := redis.Z{
		Score:  float64(timestamp),
		Member: message,
	}
	return s.redisClient.ZAdd(chatroom, z).Err()
}

// GetChat will return the list of raw data of the message.
// Since the cursor will be the last item in the previous pagination, it will exclude the message at the cursor,
func (s *redisDA) GetChat(ctx context.Context, chat string, cursor int64, limit int32, isAsc bool) ([]string, int64, error) {
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

func (s *redisDA) getRedisZRangeByScoreWithScores(chat string, cursor int64, limit int32) ([]redis.Z, error) {
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

func (s *redisDA) getRedisZRevRangeByScoreWithScores(chat string, cursor int64, limit int32) ([]redis.Z, error) {
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