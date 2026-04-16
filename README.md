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
MODEL=gpt-5-nano
API_KEY=your_api_key
BASE_URL=https://api.openai.com/v1
DEBUG_MESSAGES=false
```

可选配置：

- `GRPC_ADDR`：gRPC 监听地址，默认 `:50051`
- `TEMPERATURE`：采样温度
- `REQUEST_TIMEOUT`：请求超时时间，例如 `90s`

`GRPC_ADDR` 支持两种写法：

- `:50051`
- `50051`，程序会自动转换成 `:50051`

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

`examples/grpc_client.go` 顶部已经集中定义了本次请求参数，后续你只需要改这些变量就能调整：

- 模型和系统提示词
- `max_tokens`
- `temperature` / `top_p`
- `request_timeout`
- `max_react_rounds`
- 是否开启 `debug_messages`

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
