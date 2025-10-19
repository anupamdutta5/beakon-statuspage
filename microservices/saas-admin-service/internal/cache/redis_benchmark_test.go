package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

// BenchmarkCacheGet benchmarks cache GET operations
func BenchmarkCacheGet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()
	key := "benchmark:key"
	value := "benchmark_value"

	// Pre-populate cache
	_ = client.Set(ctx, key, value, 1*time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.Get(ctx, key)
		}
	})
}

// BenchmarkCacheSet benchmarks cache SET operations
func BenchmarkCacheSet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()
	value := "benchmark_value"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:key:%d", i)
			_ = client.Set(ctx, key, value, 1*time.Hour)
			i++
		}
	})
}

// BenchmarkCacheGetMiss benchmarks cache GET operations with cache misses
func BenchmarkCacheGetMiss(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:missing:%d", i)
			_, _ = client.Get(ctx, key)
			i++
		}
	})
}

// BenchmarkCacheCompression benchmarks compressed cache operations
func BenchmarkCacheCompression(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()
	// Create 10KB payload to trigger compression
	largeValue := make([]byte, 10*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	b.ResetTimer()
	b.Run("SetCompressed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("benchmark:compressed:%d", i)
			_ = client.SetCompressed(ctx, key, largeValue, 1*time.Hour)
		}
	})

	b.Run("GetCompressed", func(b *testing.B) {
		key := "benchmark:compressed:read"
		_ = client.SetCompressed(ctx, key, largeValue, 1*time.Hour)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = client.GetCompressed(ctx, key)
		}
	})
}

// BenchmarkCircuitBreaker benchmarks circuit breaker overhead
func BenchmarkCircuitBreaker(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()
	key := "benchmark:circuit"
	value := "test_value"

	_ = client.Set(ctx, key, value, 1*time.Hour)

	b.ResetTimer()
	b.Run("WithCircuitBreaker", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.Get(ctx, key)
		}
	})
}

// BenchmarkMetricsTracking benchmarks metrics tracking overhead
func BenchmarkMetricsTracking(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()
	key := "benchmark:metrics"
	value := "test_value"

	_ = client.Set(ctx, key, value, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Get(ctx, key)
		_ = client.GetMetrics()
	}
}

// BenchmarkCacheConcurrency benchmarks concurrent cache access
func BenchmarkCacheConcurrency(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	client := NewRedisClient(RedisConfig{
		Host:    "localhost",
		Port:    6379,
		Enabled: true,
	}, logger)

	ctx := context.Background()

	// Pre-populate 100 keys
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("benchmark:concurrent:%d", i)
		_ = client.Set(ctx, key, fmt.Sprintf("value_%d", i), 1*time.Hour)
	}

	b.ResetTimer()
	b.SetParallelism(100) // 100 concurrent goroutines
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:concurrent:%d", i%100)
			_, _ = client.Get(ctx, key)
			i++
		}
	})
}
