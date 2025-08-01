package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// 测试连接
	fmt.Println("测试Redis连接...")
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
	fmt.Printf("Redis连接成功: %s\n", pong)

	// 测试分布式锁
	fmt.Println("\n测试分布式锁...")
	lockKey := "test:lock:product:123"
	token := "test-token-123"
	expiration := 10 * time.Second

	// 尝试获取锁
	result := client.SetNX(ctx, lockKey, token, expiration)
	if err := result.Err(); err != nil {
		log.Fatalf("获取锁失败: %v", err)
	}

	if result.Val() {
		fmt.Println("✓ 成功获取分布式锁")
		
		// 检查锁是否存在
		exists := client.Exists(ctx, lockKey).Val()
		fmt.Printf("✓ 锁存在检查: %d\n", exists)
		
		// 获取锁的值
		value := client.Get(ctx, lockKey).Val()
		fmt.Printf("✓ 锁的值: %s\n", value)
		
		// 释放锁
		script := `
			if redis.call("GET", KEYS[1]) == ARGV[1] then
				return redis.call("DEL", KEYS[1])
			else
				return 0
			end
		`
		
		delResult := client.Eval(ctx, script, []string{lockKey}, token)
		if delResult.Err() != nil {
			log.Fatalf("释放锁失败: %v", delResult.Err())
		}
		
		if delResult.Val().(int64) == 1 {
			fmt.Println("✓ 成功释放分布式锁")
		} else {
			fmt.Println("✗ 释放锁失败：token不匹配")
		}
	} else {
		fmt.Println("✗ 获取锁失败：锁已被占用")
	}

	fmt.Println("\nRedis连接和分布式锁测试完成")
}
