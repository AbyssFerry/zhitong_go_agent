# zhitong_go_agent

一个基于 minichain + gRPC 的 Go 示例项目，提供两个服务端流接口：

- 普通大语言模型流式接口
- Agent 流式接口，内置当前时间工具

## 目录说明

- [api](api) ：gRPC 服务对接与部署入口，负责注册服务、转发流事件、启动和关闭服务器。
- [llm](llm) ：封装 minichain 的普通流和 Agent 流创建逻辑，统一读取配置并构造模型。
- [tool](tool) ：工具注册目录，目前提供 `get_current_time` 工具。
- [proto](proto) ：proto 源文件目录，服务定义和消息结构都在这里。
- [pb](pb) ：proto 编译生成的 Go 代码目录。
- [examples](examples) ：示例客户端，演示如何消费两个流接口。
- [scripts](scripts) ：proto 编译脚本，方便以后重新生成代码。
- [doc](doc) ：参考文档，包含 gRPC 和 minichain 的使用说明。
- [main.go](main.go) ：程序入口，负责加载配置并启动 gRPC 服务。

## 环境配置

启动前先准备 `.env`，至少包含以下内容：

```env
# 基础配置
MODEL=gpt-5-nano
API_KEY=your_api_key
BASE_URL=https://api.openai.com/v1
DEBUG_MESSAGES=false
DEBUG_REQUESTS=false

# gRPC 监听地址
GRPC_ADDR=:50051

# Chat 流参数
CHAT_SYSTEM_PROMPT=你是一个简洁、可靠的助手。
CHAT_CONTEXT_TRIM_TOKEN_THRESHOLD=0
CHAT_CONTEXT_KEEP_RECENT_ROUNDS=6
CHAT_TEMPERATURE=0.3
CHAT_TOP_P=0.9
CHAT_MAX_TOKENS=2048
CHAT_PRESENCE_PENALTY=0
CHAT_FREQUENCY_PENALTY=0
CHAT_SEED=0
CHAT_REQUEST_TIMEOUT=5m

# Agent 流参数
AGENT_SYSTEM_PROMPT=你是一个会优先调用工具来回答问题的助手。
AGENT_CONTEXT_TRIM_TOKEN_THRESHOLD=0
AGENT_CONTEXT_KEEP_RECENT_ROUNDS=6
AGENT_TEMPERATURE=0.2
AGENT_TOP_P=0.9
AGENT_MAX_TOKENS=2048
AGENT_PRESENCE_PENALTY=0
AGENT_FREQUENCY_PENALTY=0
AGENT_SEED=0
AGENT_MAX_REACT_ROUNDS=20
AGENT_REQUEST_TIMEOUT=5m
```

说明：

- 没有写入的 Chat/Agent 参数会保持空值或零值，不会再由 Go 代码补默认值。
- `DEBUG_MESSAGES` 只控制 minichain 内部的消息调试输出。
- `DEBUG_REQUESTS` 只控制本项目服务端对请求和响应的结构化调试输出。
- `CHAT_CONTEXT_TRIM_TOKEN_THRESHOLD`、`AGENT_CONTEXT_TRIM_TOKEN_THRESHOLD` 设为 `0` 时表示不启用裁剪。
- `CHAT_REQUEST_TIMEOUT` 和 `AGENT_REQUEST_TIMEOUT` 支持两种写法：`90` 会按 90 秒处理，`90s`、`5m` 这类 Go duration 也可以。

`GRPC_ADDR` 支持两种写法：

- `:50051`
- `50051`，程序会自动转换成 `:50051`

当前 gRPC 请求已经收敛为只传单条文本：

- `ChatStreamRequest.message`
- `AgentStreamRequest.message`

Rust 端只需要把语音识别后的文本填到 `message`，其余模型参数全部由 Go 服务端从 `.env` 读取。

## 运行服务

### 1. 生成 proto 代码

Windows：

```powershell
./scripts/gen_proto.ps1
```

Linux 或 macOS：

```bash
./scripts/gen_proto.sh
```

### 2. 启动服务

```bash
go run .
```

默认监听 `:50051`。

如果 `50051` 端口已被占用，可以先把 `.env` 里的 `GRPC_ADDR` 改成别的端口，比如 `:50052`。

## 运行测试

这个项目已经补了单元测试，建议按下面方式验证：

### 1. 编译检查

```bash
go build ./...
```

### 2. 运行单元测试

```bash
go test ./...
```

### 3. 启动服务后运行示例客户端

先启动服务：

```bash
go run .
```

另开一个终端运行示例客户端：

```bash
go run ./examples/grpc_client.go
```

示例客户端会依次调用：

- `ChatStream`：普通流式回答
- `AgentStream`：Agent 流式回答，并触发当前时间工具

`examples/grpc_client.go` 现在只负责发送示例文本：

- `ChatStreamRequest.message`
- `AgentStreamRequest.message`

服务端会从 `.env` 读取并固定以下参数：

- 模型连接信息
- 全局调试开关
- Chat 流默认参数
- Agent 流默认参数
- gRPC 监听地址

### 4. 观察输出

你应该能看到服务端流事件逐条输出，包括：

- `content`
- `tool_start`
- `tool_end`
- `final`
- `error`

如果要单独验证启动是否正常，可以先执行：

```bash
go run .
```

再在另一个终端执行：

```bash
go run ./examples/grpc_client.go
```

建议先确认 `.env` 中的 `API_KEY`、`BASE_URL` 和 `MODEL` 可用，否则流会直接返回错误事件。

## gRPC 接口

当前 proto 定义在 [proto/zhitongAgent.proto](proto/zhitongAgent.proto)，只有一个 service：`ZhitongAgent`。

它提供两个方法：

- `ChatStream`：普通大语言模型服务端流
- `AgentStream`：Agent 服务端流

## 参考

- [gRPC 使用讲解](doc/gRPC使用讲解（go语言）.md)
- [Minichain README](doc/MinichainREADME.md)
