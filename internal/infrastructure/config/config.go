package config

import (
	"time"
)

// Config 应用配置
type Config struct {
	App      AppConfig
	Task     TaskConfig
	Database DatabaseConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string
	Environment string
	Port        int
}

// TaskConfig 任务配置
type TaskConfig struct {
	LockTimeout      time.Duration // 锁超时时间
	TaskExpireDays   int           // 任务过期天数
	MaxRetry         int           // 最大重试次数
	DefaultReward    int           // 默认奖励值
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string // memory, mysql, postgres
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "mini-sirus",
			Environment: "development",
			Port:        8080,
		},
		Task: TaskConfig{
			LockTimeout:    30 * time.Second,
			TaskExpireDays: 30,
			MaxRetry:       3,
			DefaultReward:  1,
		},
		Database: DatabaseConfig{
			Type: "memory",
		},
	}
}

