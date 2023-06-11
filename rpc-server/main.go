package main

import (
	"fmt"
	"log"

	rpc "github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc/imservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/go-redis/redis"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {

	fmt.Println("We are testing Go-Redis")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "rediscache:6379",
		Password: "Atyeyt8H91NQVR1fdwPw20xZHOLF",
		DB:       0,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Println("Redis init fail, ", err)
		log.Fatal(err)
	}

	r, err := etcd.NewEtcdRegistry([]string{"etcd:2379"}) // r should not be reused.
	if err != nil {
		log.Fatal(err)
	}

	imServiceImplInstance := &IMServiceImpl{
		redisClient: redisClient,
	}
	svr := rpc.NewServer(imServiceImplInstance, server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "demo.rpc.server",
	}))

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
