package tool

import (
	"time"

	"github.com/abyssferry/minichain/llm"
)

// CurrentTimeArgs 定义当前时间工具的入参。
type CurrentTimeArgs struct{}

// NewRegistry 构建包含当前时间工具的注册器。
func NewRegistry() (*llm.ToolRegistry, error) {
	registry := llm.NewToolRegistry()
	err := registry.RegisterFromHandler(
		"get_current_time",
		"获取当前本地时间，返回 RFC3339 格式字符串",
		func(_ CurrentTimeArgs) (string, error) {
			return time.Now().Format(time.RFC3339), nil
		},
	)
	if err != nil {
		return nil, err
	}
	return registry, nil
}
