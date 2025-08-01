package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string
	Port        string
	BinaryPath  string
	Description string
}

// 所有微服务配置
var services = map[string]ServiceConfig{
	"gateway": {
		Name:        "API Gateway",
		Port:        "8080",
		BinaryPath:  "./bin/gateway",
		Description: "API网关服务，统一入口",
	},
	"user": {
		Name:        "User Service",
		Port:        "8081",
		BinaryPath:  "./bin/user-service",
		Description: "用户管理服务",
	},
	"product": {
		Name:        "Product Service",
		Port:        "8082",
		BinaryPath:  "./bin/product-service",
		Description: "商品管理服务",
	},
	"order": {
		Name:        "Order Service",
		Port:        "8083",
		BinaryPath:  "./bin/order-service",
		Description: "订单管理服务",
	},
	"seckill": {
		Name:        "Seckill Service",
		Port:        "8084",
		BinaryPath:  "./bin/seckill-service",
		Description: "秒杀服务",
	},
	"payment": {
		Name:        "Payment Service",
		Port:        "8085",
		BinaryPath:  "./bin/payment-service",
		Description: "支付服务",
	},
}

func main() {
	var (
		serviceFlag = flag.String("service", "", "指定要启动的服务 (gateway,user,product,order,seckill,payment,all)")
		buildFlag   = flag.Bool("build", false, "是否先构建服务")
		helpFlag    = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	// 如果需要构建
	if *buildFlag {
		if err := buildServices(*serviceFlag); err != nil {
			log.Fatalf("构建失败: %v", err)
		}
	}

	// 启动服务
	if *serviceFlag == "" {
		fmt.Println("请指定要启动的服务，使用 -help 查看帮助")
		return
	}

	if *serviceFlag == "all" {
		startAllServices()
	} else {
		startSingleService(*serviceFlag)
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("Ryan Mall 微服务启动器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  go run cmd/main.go -service=<service_name> [-build]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -service string")
	fmt.Println("        指定要启动的服务:")
	for key, config := range services {
		fmt.Printf("          %s: %s (端口:%s)\n", key, config.Description, config.Port)
	}
	fmt.Println("          all: 启动所有服务")
	fmt.Println("  -build")
	fmt.Println("        启动前先构建服务")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run cmd/main.go -service=gateway -build")
	fmt.Println("  go run cmd/main.go -service=all")
	fmt.Println("  go run cmd/main.go -service=user")
}

// buildServices 构建服务
func buildServices(serviceFlag string) error {
	fmt.Println("🔨 开始构建服务...")

	if serviceFlag == "all" {
		// 构建所有服务
		for key := range services {
			if err := buildSingleService(key); err != nil {
				return fmt.Errorf("构建服务 %s 失败: %w", key, err)
			}
		}
	} else {
		// 构建单个服务
		if err := buildSingleService(serviceFlag); err != nil {
			return fmt.Errorf("构建服务 %s 失败: %w", serviceFlag, err)
		}
	}

	fmt.Println("✅ 构建完成")
	return nil
}

// buildSingleService 构建单个服务
func buildSingleService(serviceName string) error {
	config, exists := services[serviceName]
	if !exists {
		return fmt.Errorf("未知服务: %s", serviceName)
	}

	fmt.Printf("  构建 %s...\n", config.Name)

	var cmdPath string
	switch serviceName {
	case "gateway":
		cmdPath = "./cmd/gateway"
	case "user":
		cmdPath = "./cmd/user"
	case "seckill":
		cmdPath = "./cmd/seckill-service"
	default:
		cmdPath = fmt.Sprintf("./cmd/%s-service", serviceName)
	}

	cmd := exec.Command("go", "build", "-o", config.BinaryPath, cmdPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// startAllServices 启动所有服务
func startAllServices() {
	fmt.Println("🚀 启动所有微服务...")

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// 按顺序启动服务（网关最后启动）
	startOrder := []string{"user", "product", "order", "seckill", "payment", "gateway"}

	for _, serviceName := range startOrder {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			startServiceProcess(ctx, name)
		}(serviceName)

		// 给每个服务一点启动时间
		time.Sleep(2 * time.Second)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 正在关闭所有服务...")
	cancel()
	wg.Wait()
	fmt.Println("✅ 所有服务已关闭")
}

// startSingleService 启动单个服务
func startSingleService(serviceName string) {
	config, exists := services[serviceName]
	if !exists {
		log.Fatalf("未知服务: %s", serviceName)
	}

	fmt.Printf("🚀 启动 %s (端口:%s)...\n", config.Name, config.Port)

	ctx, cancel := context.WithCancel(context.Background())

	go startServiceProcess(ctx, serviceName)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Printf("\n🛑 正在关闭 %s...\n", config.Name)
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Printf("✅ %s 已关闭\n", config.Name)
}

// startServiceProcess 启动服务进程
func startServiceProcess(ctx context.Context, serviceName string) {
	config := services[serviceName]

	// 检查二进制文件是否存在
	if _, err := os.Stat(config.BinaryPath); os.IsNotExist(err) {
		log.Printf("❌ %s 二进制文件不存在: %s", config.Name, config.BinaryPath)
		log.Printf("   请先运行: go run cmd/main.go -service=%s -build", serviceName)
		return
	}

	// 设置环境变量
	env := os.Environ()
	env = append(env, fmt.Sprintf("%s_PORT=%s", strings.ToUpper(serviceName), config.Port))

	cmd := exec.CommandContext(ctx, config.BinaryPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("✅ %s 启动成功 (PID: %d, 端口: %s)\n", config.Name, cmd.Process.Pid, config.Port)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			// 正常关闭
			return
		}
		log.Printf("❌ %s 运行出错: %v", config.Name, err)
	}
}
