# 项目名称：geektime-basic-learn-go

## 项目简介

本项目是一个基于Go语言的后端服务示例，学习极客时间初级Go训练营，采用整洁架构和领域驱动设计（DDD）思想，适合学习和实践现代后端开发。项目主要实现了用户注册、登录、文章管理等基础功能，代码结构清晰，易于维护和扩展。

## 主要功能

- 用户注册、登录（支持邮箱和短信验证码登录）
- 用户信息编辑与查询
- 文章的创建、编辑、发布、撤回、查询
- 基于Gin框架的RESTful API
- 配置灵活，支持本地和远程配置中心
- 日志、限流等中间件支持

## 目录结构

```
.
├── main.go              // 程序入口
├── config/              // 配置文件目录
│   └── dev.yaml         // 本地开发配置
├── internal/            // 业务核心代码
│   ├── domain/          // 领域模型
│   ├── repository/      // 数据访问层
│   ├── service/         // 业务服务层
│   └── web/             // API接口与路由
├── pkg/                 // 通用工具包
├── go.mod               // 依赖管理
└── ...
```

## 快速开始

1. **安装依赖**

   ```bash
   go mod tidy
   ```

2. **配置数据库和Redis**

   修改 `config/dev.yaml`，配置你的数据库和Redis连接信息：

   ```yaml
   db:
     dsn: "root:root@tcp(localhost:3306)/webook"
   redis:
     addr: "localhost:6379"
   ```

3. **运行项目**

   ```bash
   go run main.go
   ```

   启动后访问 [http://localhost:8080/hello](http://localhost:8080/hello) ，看到"hello，启动成功了！"说明服务启动成功。

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

## 配置说明

`config/dev.yaml` 示例：

```yaml
test:
  key: value1234

redis:
  addr: "localhost:6379"

db:
  dsn: "root:root@tcp(localhost:3306)/webook"
```

## 依赖技术

- Go 1.23+
- Gin Web框架
- GORM ORM
- Viper 配置管理
- Zap 日志
- MySQL、Redis、MongoDB

## 贡献与反馈

如有问题或建议，欢迎提Issue或直接联系项目维护者。 