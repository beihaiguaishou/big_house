package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/xen0n/go-workwx"
)

var (
	workWeChatCorpID      string
	workWeChatAgentSecret string
	workWeChatAgentId     int64
	workWeChatNotifyUsers string
)

var (
	redisAddress  string
	redisPassword string
)

var WorkWeChatApp *workwx.WorkwxApp
var Redis *redis.Client

func initConf() {
	WorkWeChatApp = workwx.New(workWeChatCorpID).WithApp(workWeChatAgentSecret, workWeChatAgentId)
	Redis = redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword, DB: 2})
}
