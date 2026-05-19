package copilot

import "os"

// Config Copilot 配置
type Config struct {
	MonolithAPIBase        string
	MemoryDir              string
	MemoryWindowSize       int
	OrchestratorTimeoutSec int
	PlannerTimeoutSec      int
	TotalTimeoutSec        int
	FixturesDir            string
}

// DefaultConfig 返回默认配置，优先使用环境变量
func DefaultConfig() *Config {
	return &Config{
		MonolithAPIBase:        getEnvOrDefault("MONOLITH_API_BASE", "http://localhost:8080"),
		MemoryDir:              getEnvOrDefault("COPILOT_MEMORY_DIR", "./data/memory"),
		MemoryWindowSize:       getEnvIntOrDefault("COPILOT_MEMORY_WINDOW_SIZE", 10),
		OrchestratorTimeoutSec: getEnvIntOrDefault("COPILOT_ORCHESTRATOR_TIMEOUT", 30),
		PlannerTimeoutSec:      getEnvIntOrDefault("COPILOT_PLANNER_TIMEOUT", 45),
		TotalTimeoutSec:        getEnvIntOrDefault("COPILOT_TOTAL_TIMEOUT", 60),
		FixturesDir:            getEnvOrDefault("COPILOT_FIXTURES_DIR", "./data/fixtures"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	var result int
	for _, c := range v {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return defaultVal
		}
	}
	return result
}
