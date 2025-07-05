# HTTP服务器优化建议

## 1. 服务器配置优化

在 `cmd/server/main.go` 中优化HTTP服务器配置：

```go
server := &http.Server{
    Addr:           ":" + cfg.Server.Port,
    Handler:        r,
    ReadTimeout:    5 * time.Second,   // 减少读取超时
    WriteTimeout:   5 * time.Second,   // 减少写入超时
    IdleTimeout:    30 * time.Second,  // 减少空闲超时
    MaxHeaderBytes: 1 << 16,           // 减少最大请求头大小 64KB
    
    // 启用HTTP/2
    TLSConfig: &tls.Config{
        NextProtos: []string{"h2", "http/1.1"},
    },
}
```

## 2. 连接池优化

优化数据库连接池配置：

```go
// 根据并发需求调整
sqlDB.SetMaxOpenConns(200)        // 减少到200
sqlDB.SetMaxIdleConns(50)         // 减少到50
sqlDB.SetConnMaxLifetime(5 * time.Minute)   // 减少生命周期
sqlDB.SetConnMaxIdleTime(2 * time.Minute)   // 减少空闲时间
```

## 3. 缓存优化

减少分片数量以降低开销：

```go
// 从32分片减少到16分片
cache.SetGlobalCache(cache.NewShardedCache(16))
```
