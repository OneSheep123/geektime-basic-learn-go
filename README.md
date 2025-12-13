# geektime-basic-learn-go

## 项目简介

本项目是基于 Go 语言开发的后端服务示例，采用整洁架构和领域驱动设计（DDD）思想构建。项目实现了用户系统、文章管理等核心功能，展示了现代 Go 后端开发的最佳实践。

## 核心功能

- 用户注册、登录（支持邮箱和短信验证码登录）
- 用户信息编辑与查询
- 文章的创建、编辑、发布、撤回、查询
- 基于 JWT 的身份验证
- 文章互动功能（点赞、收藏等）
- 分布式任务调度
- 基于 gRPC 的微服务通信
- 配置灵活，支持本地和远程配置中心（ETCD）
- 集成 Prometheus 监控指标
- 使用 Kafka 进行事件驱动架构
- 支持分布式限流

## 技术栈

- **语言**: Go 1.23+
- **Web框架**: Gin v1.10.0
- **ORM框架**: GORM v1.30.0
- **配置管理**: Viper v1.20.1
- **日志系统**: Zap v1.27.0
- **数据库**: MySQL 8.0, MongoDB 6.0
- **缓存**: Redis
- **消息队列**: Kafka
- **监控**: Prometheus
- **配置中心**: ETCD
- **微服务通信**: gRPC v1.67.3
- **其他**: Docker, Wire DI

## 目录结构

```
.
├── api/                  // Proto 定义文件
│   └── proto/
├── config/               // 配置文件目录
│   └── dev.yaml          // 本地开发配置
├── cronjob/              // 定时任务示例
├── grpc/                 // gRPC 相关代码
├── interactive/          // 互动服务（微服务）
├── internal/             // 业务核心代码
│   ├── domain/           // 领域模型
│   ├── repository/       // 数据访问层
│   ├── service/          // 业务服务层
│   ├── web/              // API 接口与路由
│   ├── events/           // 事件处理
│   ├── job/              // 任务调度
│   └── integration/      // 集成测试
├── ioc/                  // 依赖注入配置
├── pkg/                  // 通用工具包
├── sarama/               // Kafka 示例
├── main.go               // 程序入口
├── buf.gen.yaml          // Buf 代码生成配置
├── docker-compose.yaml   // 开发环境依赖
├── go.mod                // 依赖管理
└── wire.go               // Wire 依赖注入配置
```

## 环境要求

- Go 1.23+
- Docker & Docker Compose
- Buf CLI (用于 protobuf 代码生成)

## 快速开始

1. **启动依赖服务**

   ```bash
   docker-compose up -d
   ```

2. **安装依赖**

   ```bash
   go mod tidy
   ```

3. **配置数据库**

   修改 `config/dev.yaml`，配置你的数据库连接信息：

   ```yaml
   db:
     dsn: "root:123123@tcp(localhost:3306)/webook"
   redis:
     addr: "localhost:6379"
   kafka:
     addr:
       - "127.0.0.1:9094"
   grpc:
     client:
       intr:
         addr: "localhost:8090"
         threshold: 100
   ```

4. **生成 gRPC 代码**

   ```bash
   buf generate api/proto
   ```

5. **运行主服务**

   ```bash
   go run main.go
   ```

   启动后访问 [http://localhost:8083/hello](http://localhost:8083/hello) ，看到"hello，启动成功了！"说明服务启动成功。

6. **运行互动服务（可选）**

   ```bash
   cd interactive
   go run main.go
   ```

## 主要接口示例

### 用户相关

- 注册：`POST /users/signup`
- 登录（邮箱）：`POST /users/login`
- 登录（短信验证码）：`POST /users/login_sms`
- 发送短信验证码：`POST /users/login_sms/code/send`
- 编辑信息：`POST /users/edit`
- 查询个人信息：`GET /users/profile`

### 文章相关

- 新建/编辑文章：`POST /articles/edit`
- 发布文章：`POST /articles/publish`
- 撤回文章：`POST /articles/withdraw`
- 查询文章详情：`GET /articles/detail/:id`
- 查询文章列表：`POST /articles/list`

### 互动相关

- 点赞：`POST /interactive/like`
- 收藏：`POST /interactive/collect`
- 获取互动信息：`GET /interactive/detail`

## 开发说明

### 依赖服务

项目依赖以下服务，可通过 `docker-compose up -d` 一键启动：

- MySQL 8.0 (端口 13316)
- Redis (端口 6379)
- MongoDB 6.0 (端口 27017)
- Kafka (端口 9094)
- Prometheus (端口 9090)
- ETCD (端口 12379)

### 配置文件

主要配置在 `config/dev.yaml` 文件中，支持热重载。

### 依赖注入

项目使用 Google Wire 进行依赖注入，修改依赖关系后需要运行：

```bash
go generate ./...
```

### Protobuf 代码生成

项目使用 Buf 工具进行 protobuf 代码生成，配置文件为 `buf.gen.yaml`：

- `buf.gen.yaml`: 定义了 protobuf 代码生成的插件和输出路径
- API 定义文件位于 `api/proto/` 目录下

生成命令：
```bash
buf generate api/proto
```

该命令会根据 `buf.gen.yaml` 的配置生成 Go 代码和 gRPC 代码到 `api/proto/gen` 目录。

## 贡献与反馈

如有问题或建议，欢迎提 Issue 或直接联系项目维护者。