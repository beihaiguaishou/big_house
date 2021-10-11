package main

import (
	"flag"
	"context"
	"time"
)

func init() {
	flag.StringVar(&workWeChatCorpID, "workWeChatCorpID", "", "企业微信公司ID")
	flag.StringVar(&workWeChatAgentSecret, "workWeChatAgentSecret", "", "企业微信应用秘钥")
	flag.Int64Var(&workWeChatAgentId, "workWeChatAgentId", 0, "企业微信应用ID")
	flag.StringVar(&workWeChatNotifyUsers, "workWeChatNotifyUsers", "", "企业微信通知用户")
	flag.StringVar(&redisAddress, "redisAddress", "", "Redis服务器地址")
	flag.StringVar(&redisPassword, "redisPassword", "", "Redis密码")
}

func main() {
	flag.Parse()
	initConf()
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	for {
		select {
		case <-ctx.Done():
			cancel()
			if err := NotifyNewHouses(workWeChatNotifyUsers); err == nil {
				ctx, cancel = context.WithTimeout(context.Background(), time.Hour)
			}
		}
	}
}
