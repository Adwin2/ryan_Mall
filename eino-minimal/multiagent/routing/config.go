package routing

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type RoutingConfig struct {
	Profiles map[string]*ProfileConfig `yaml:"profiles"`
	Rules    []RoutingRule             `yaml:"rules"`
}

type ProfileConfig struct {
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
	// 记忆策略 hint
	MaxHistoryTurns   int  `yaml:"max_history_turns"`
	UseSummary        bool `yaml:"use_summary"`
	UseSemanticRecall bool `yaml:"use_semantic_recall"`
	MaxRecallFacts    int  `yaml:"max_recall_facts"`
}

type RoutingRule struct {
	Agent    string         `yaml:"agent"`
	Tier     string         `yaml:"tier"`
	Variants []RouteVariant `yaml:"variants"`
}

type RouteVariant struct {
	Profile string `yaml:"profile"`
	Weight  int    `yaml:"weight"`
}

type ConfigManager struct {
	path    string
	config  atomic.Value // *RoutingConfig
	watcher *fsnotify.Watcher
	done    chan struct{}
	mu      sync.Mutex
}

func NewConfigManager(path string) (*ConfigManager, error) {
	cm := &ConfigManager{
		path: path,
		done: make(chan struct{}),
	}

	if err := cm.load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := cm.startWatcher(); err != nil {
		log.Printf("[ConfigManager] watcher启动失败(热更新不可用): %v", err)
	}

	return cm, nil
}

func (cm *ConfigManager) Get() *RoutingConfig {
	return cm.config.Load().(*RoutingConfig)
}

func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.load()
}

func (cm *ConfigManager) Close() {
	close(cm.done)
	if cm.watcher != nil {
		cm.watcher.Close()
	}
}

func (cm *ConfigManager) load() error {
	data, err := os.ReadFile(cm.path)
	if err != nil {
		return fmt.Errorf("read %s: %w", cm.path, err)
	}

	var cfg RoutingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	cm.config.Store(&cfg)
	log.Printf("[ConfigManager] loaded %d profiles, %d rules from %s",
		len(cfg.Profiles), len(cfg.Rules), filepath.Base(cm.path))
	return nil
}

func (cm *ConfigManager) startWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(filepath.Dir(cm.path)); err != nil {
		watcher.Close()
		return err
	}

	cm.watcher = watcher
	go cm.watchLoop()
	return nil
}

func (cm *ConfigManager) watchLoop() {
	target := filepath.Base(cm.path)
	for {
		select {
		case <-cm.done:
			return
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) == target && (event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Create != 0) {
				if err := cm.Reload(); err != nil {
					log.Printf("[ConfigManager] hot-reload failed: %v", err)
				} else {
					log.Printf("[ConfigManager] hot-reload success")
				}
			}
		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[ConfigManager] watcher error: %v", err)
		}
	}
}

func (c *RoutingConfig) validate() error {
	if len(c.Profiles) == 0 {
		return fmt.Errorf("no profiles defined")
	}
	for _, rule := range c.Rules {
		if len(rule.Variants) == 0 {
			return fmt.Errorf("rule agent=%s tier=%s has no variants", rule.Agent, rule.Tier)
		}
		for _, v := range rule.Variants {
			if _, ok := c.Profiles[v.Profile]; !ok {
				return fmt.Errorf("variant references unknown profile %q", v.Profile)
			}
		}
	}
	return nil
}
