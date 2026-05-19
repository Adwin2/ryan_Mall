package multiagent

import (
	"eino-minimal/multiagent/memory"
	"eino-minimal/multiagent/routing"
)

type AgentInfra struct {
	Router *routing.ModelRouter
	Memory *memory.MemoryManager
}

type Event struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Tier    string `json:"tier,omitempty"`
	Profile string `json:"profile,omitempty"`
	Model   string `json:"model,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Node    string `json:"node,omitempty"`
	Status  string `json:"status,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Params  any    `json:"params,omitempty"`
	Result  string `json:"result,omitempty"`
}
