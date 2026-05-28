# MVP 部署验证计划

## 概述

本计划指导完成 Crab MVP 版本的端到端测试验证，包括：
- 后端服务部署
- 数据库迁移
- iOS 家长端安装
- Android 孩子端安装
- 完整业务流程验证

## 当前状态

| 组件 | 状态 | 说明 |
|------|------|------|
| 后端 Go 服务 | ⚠️ 部分 | 使用内存存储，需连接真实数据库 |
| PostgreSQL | ✅ 就绪 | docker-compose 已配置 |
| Redis | ✅ 就绪 | docker-compose 已配置 |
| iOS 家长端 | ✅ 已实现 | 需真机安装 |
| Android 孩子端 | ⚠️ 框架已搭建 | 需完善并安装 |

---

## Phase 1: 后端服务部署 (预计 30 分钟)

### 1.1 启动基础设施

```bash
cd code/server
docker-compose up -d
```

验证：
```bash
# PostgreSQL
docker exec crab-postgres pg_isready -U postgres

# Redis
docker exec crab-redis redis-cli ping
```

### 1.2 数据库迁移

```bash
# 方案 A: 使用 golang-migrate (推荐)
brew install golang-migrate
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/crab?sslmode=disable" up

# 方案 B: 手动执行
docker exec -i crab-postgres psql -U postgres -d crab < migrations/001_create_families.up.sql
# 依次执行其他迁移文件
```

### 1.3 配置环境变量

```bash
cd code/server
cat > .env << 'EOF'
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development
DATABASE_URL=postgres://postgres:postgres@localhost:5432/crab?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=dev-secret-change-in-production-32chars
EOF
```

### 1.4 后端代码修改 (必须)

**当前问题**: 服务使用内存存储，需要连接真实数据库。

需要修改:
1. `cmd/api/main.go` - 添加数据库连接
2. 创建 `internal/repository/postgres/` - PostgreSQL 实现

### 1.5 启动服务

```bash
cd code/server
go run cmd/api/main.go
```

验证：
```bash
curl http://localhost:8080/healthz
```

---

## Phase 2: iOS 家长端安装 (预计 20 分钟)

### 2.1 前置条件

- Mac + Xcode 15+
- Apple 开发者账号
- iPhone (iOS 15+)

### 2.2 配置步骤

```bash
cd code/parent-ios
open ParentApp.xcodeproj
```

1. Signing & Capabilities → 选择 Team
2. 对 ShareExtension 重复相同步骤
3. 修改 `APIClient.swift` 中的服务器地址
4. Product → Run 安装到真机

---

## Phase 3: Android 孩子端安装 (预计 30 分钟)

### 3.1 前置条件

- Android Studio
- Android 设备 (API 24+) 或模拟器

### 3.2 配置步骤

```bash
cd code/child-app
open -a "Android Studio" .
```

1. 修改 `NetworkModule.kt` 中的服务器地址
2. Run → Run 'app'

---

## Phase 4: 端到端验证流程

### 4.1 创建家庭
1. 打开家长端 → 创建家庭
2. 记录配对码

### 4.2 孩子端配对
1. 打开孩子端 → 输入配对码 → 配对

### 4.3 分享内容
1. Safari 打开链接 → 分享 → Crab 扩展

### 4.4 浏览内容
1. 孩子端刷新 → 点击内容 → 播放

---

## 后端必须完成项

- [ ] 实现 PostgreSQL Repository
- [ ] 实现 Redis 任务队列
- [ ] 实现 `/api/items` 内容列表接口
- [ ] 实现 `/api/items/:id/playback` 播放配置接口

## 客户端必须完成项

- [ ] iOS: 配置真实服务器地址
- [ ] Android: 实现 Feed 页和 WebView 播放页
