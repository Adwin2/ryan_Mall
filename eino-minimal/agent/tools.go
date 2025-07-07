package agent

import (
	"eino-minimal/tools"

	"github.com/cloudwego/eino/components/tool"
)

// GetTools 获取工具列表
// 这里返回空列表，表示不使用任何工具
// 如果需要添加工具，可以在这里添加
func GetTools() []tool.BaseTool {
	return []tool.BaseTool{
		// 可以在这里添加自定义工具
		// 例如: WeatherTool(), SearchTool(), etc.
		tools.FileUrlTool(),
	}
}
