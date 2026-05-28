# 家庭内部儿童内容浏览工具 技术方案

## 1. 文档信息
- 文档名称：家庭内部儿童内容浏览工具 技术方案
- 版本：v0.2
- 日期：2026-05-26
- 状态：草案
- 对应 PRD：`docs/prd-family-child-content-app.md`

## 2. 目标
本文档将 PRD 中定义的 MVP 范围转换为可实现的技术方案，重点覆盖：
- 系统边界
- 前后端职责
- 关键模块拆分
- 抖音内容处理链路
- WebView 播放与私有视频兜底策略
- 数据模型
- 接口草案
- MVP 实施顺序

MVP 只解决一个最小闭环：

> 家长从抖音、X、微信视频号、网页等来源分享内容，系统处理后分发给孩子，孩子在受控 App 内浏览和播放。

## 3. 总体架构
采用"一个后端 + 一个家长分享入口 + 一个孩子端 App"的最小结构。

### 3.1 MVP 组件划分
1. **家长端分享壳**
   - 首期优先作为系统分享面板中的 share target / share extension
   - 支持剪贴板链接自动识别
   - 只负责接收外部分享内容、展示最小预览、确认发送给孩子

2. **孩子端 App**
   - 首期仅做 Android
   - 只负责内容列表和受控播放

3. **Go API 服务**
   - 提供导入、分发、feed、播放配置接口
   - 负责鉴权与家庭身份管理
   - 同时负责任务编排

4. **异步 Worker**
   - 负责链接解析、元信息抓取、可播放性探测、兜底处理

5. **存储层**
   - PostgreSQL：业务数据
   - 对象存储：封面、私有视频文件、处理产物
   - Redis：任务队列与短期状态

## 4. 推荐技术栈
以下是偏稳妥的 MVP 选择，不追求"最优"，追求"尽快闭环"。

### 4.1 家长端轻量壳
首期不做完整管理台，优先做一个能出现在系统分享面板中的轻量壳。

建议：
- Android：优先支持系统分享目标（share target）+ 剪贴板监听
- iOS：后续补 share extension 或等价分享入口

首期能力只需覆盖：
- 接收外部分享进来的链接或内容引用
- 监听系统剪贴板，识别已复制的链接
- 提供手动粘贴入口
- 做最小预览
- 允许家长确认发送给孩子

技术选型可尽量轻量，目标是先打通分享入口，而不是做复杂运营后台。

### 4.2 孩子端 App
MVP 不再并行讨论多端方案，直接收敛为：

- **原生 Android 优先**

原因：
- 更容易做"安全壳"
- 更容易控制 WebView、深链、跳转、播放器容器
- 可以减少 MVP 阶段的技术分叉

iOS 暂不纳入首期范围。

### 4.3 后端
- Go
- 首期建议使用标准 `net/http` + 路由框架，保持可控和简单
- 可选：
  - Gin
  - Chi
  - Echo

更推荐：**Chi 或 Gin**
- 上手快
- 结构足够清晰
- 方便后续拆分 ingest / playback / assignment 等模块

### 4.4 数据库与存储
- PostgreSQL：业务主库
- S3 兼容对象存储：封面、缓存素材、私有视频
- Redis：任务状态缓存、去重、短期队列辅助

### 4.5 异步任务
- Redis + Go 队列框架
- 可选：
  - Asynq
  - Machinery
  - 自建轻量任务队列
- 用于：
  - 链接解析
  - 内容元信息抓取
  - 可播放性探测
  - 私有视频兜底处理

## 5. 系统边界与职责

### 5.1 家长端职责
- 作为系统分享面板中的可选目标
- 监听剪贴板，识别链接
- 接受分享/粘贴的内容链接或内容引用
- 展示最小导入预览
- 分配内容给孩子
- 设置家长控制参数（时长限制等）

### 5.2 孩子端职责
- 配对码输入完成家庭绑定
- 拉取家长分配给自己的内容
- 在受控容器中播放/阅读
- 记录浏览、播放、收藏、最近看过
- 执行家长控制策略（时长限制）

### 5.3 后端职责
- 内容主数据管理
- 家庭与身份管理
- 内容处理任务编排
- 播放策略决策
- 孩子端播放元数据下发
- 安全边界控制信息下发
- 家长控制策略存储与校验

### 5.4 内容处理服务职责
- 识别链接来源
- 拉取标题、封面、简介等基础信息
- 判断优先播放模式
- 标记 WebView 是否可用
- 在需要时生成私有视频兜底版本

## 6. MVP 核心业务对象
MVP 先只保留 5 个核心对象：

### 6.1 Family
家庭单元，所有内容和分配都在家庭范围内。

关键字段：
- `id`
- `pairingCode`：配对码，用于孩子端绑定
- `createdAt`

### 6.2 Item
统一抽象所有内容。

关键字段：
- `id`
- `familyId`
- `sourceType`：`douyin | x | wechat_channels | web_article`
- `sourceUrl`
- `normalizedUrl`：标准化后的 URL，用于去重
- `title`
- `coverUrl`
- `summary`
- `playbackMode`：`webview | private_video | article | audio`（article/audio 为后续扩展）
- `processingStatus`：`pending | ready | failed`
- `playbackStatus`：`unknown | playable | fallback_required`
- `createdAt`
- `updatedAt`

### 6.3 Child
关键字段：
- `id`
- `familyId`
- `name`
- `deviceId`：绑定的设备标识
- `dailyTimeLimitSec`：每日使用时长限制（秒），默认 3600
- `status`：`active | inactive`
- `createdAt`

### 6.4 Assignment
关键字段：
- `id`
- `itemId`
- `childId`
- `assignedBy`
- `assignedAt`
- `state`：`active | revoked | expired`

`state` 说明：
- `active`：内容对孩子可见
- `revoked`：家长主动撤回，孩子端不再显示
- `expired`：超过设定天数孩子未查看，自动标记（MVP 可暂不启用自动过期）

### 6.5 PlaybackAsset
关键字段：
- `id`
- `itemId`
- `kind`
- `url`
- `status`
- `createdAt`

时间字段统一使用 **ISO 8601 UTC**，例如：`2026-05-26T14:30:00Z`

## 7. 数据库表建议

### 7.1 `families`
- `id`
- `pairing_code`
- `created_at`

### 7.2 `items`
- `id`
- `family_id`
- `title`
- `cover_url`
- `summary`
- `source_type`
- `source_url`
- `normalized_url`
- `playback_mode`
- `processing_status`
- `playback_status`
- `age_band`
- `duration_sec`
- `created_at`
- `updated_at`

### 7.3 `item_tags`
- `id`
- `item_id`
- `tag`

### 7.4 `children`
- `id`
- `family_id`
- `name`
- `avatar_url`
- `device_id`
- `daily_time_limit_sec`
- `status`
- `created_at`

### 7.5 `assignments`
- `id`
- `item_id`
- `child_id`
- `assigned_by`
- `assigned_at`
- `state`

### 7.6 `playback_assets`
- `id`
- `item_id`
- `kind`
- `asset_url`
- `status`
- `meta_json`
- `created_at`

### 7.7 `play_events`
- `id`
- `child_id`
- `item_id`
- `event_type`
- `position_sec`
- `occurred_at`    -- 客户端记录的事件实际发生时间
- `created_at`     -- 服务端接收写入时间

`event_type` 可取：
- `open`
- `start`
- `pause`
- `complete`
- `favorite`

### 7.8 `daily_usage`
- `id`
- `child_id`
- `date`
- `total_play_sec`
- `updated_at`

## 8. MVP 前端范围

### 8.1 家长端分享壳
首期只保留一个确认页。

功能：
- 接收系统分享进来的链接
- 监听剪贴板，识别抖音/X/视频号等链接并提供快速导入提示
- 手动粘贴链接输入框
- 展示最小预览
- 选择孩子
- 点击发送

不做：
- 内容列表页
- 内容详情页
- 复杂筛选
- 独立管理后台

### 8.2 孩子端 App
首期只保留两个核心页面：

#### 8.2.1 Feed 页
- 拉取已分配内容（仅 `state = active`）
- 按最近投喂时间排序
- 展示封面、标题
- 显示当日剩余可用时长

#### 8.2.2 受控播放页
- 根据 `playbackMode` 决定走 WebView、原生视频或文章页
- 拦截外链
- 禁止跳出 App
- 屏蔽非目标页面跳转
- 上报播放事件（带上客户端时间戳）
- 播放前校验当日时长余额，超时展示提示页

MVP 暂不做：
- 分类页
- 收藏页
- 最近看过

## 9. MVP 后端模块

### 9.1 IngestModule
- 接收分享链接
- 标准化 URL
- 对 `normalized_url` 做去重检查
- 判断来源类型
- 创建 `Item`
- 投递异步处理任务

### 9.2 AssignmentModule
- 把内容分配给孩子
- 支持撤回（`state = revoked`）
- 支持重新上线（`state = active`）
- 查询孩子 feed（过滤 `state = active`）

### 9.3 PlaybackModule
- 返回当前最佳播放方式
- 返回播放资源 URL
- 返回 WebView 安全策略参数
- 播放前校验每日时长余额

### 9.4 EventModule
- 记录打开、开始播放、完成等基础事件
- 记录每日使用时长累计

### 9.5 WorkerModule
- 元信息抓取
- 可播放性探测
- 私有视频兜底处理

**Worker 容错与重试策略**：
- 单次任务失败后自动重试，最多 3 次
- 重试间隔采用指数退避：1 分钟 → 5 分钟 → 15 分钟
- 3 次重试均失败后，更新 Item 的 `processingStatus = failed`，并记录失败原因
- Worker 崩溃恢复：任务基于 Redis 队列，Worker 重启后自动从队列拉取未完成任务

## 10. 多来源内容处理链路
MVP 首期重点覆盖抖音、X、微信视频号和普通网页，其中抖音链路仍然是当前最关键的验证链路。

## 10.1 导入链路（通用）
家长提交链接后：

1. 后端标准化链接（展开短链、去除追踪参数）
2. 根据域名/URL 模式识别来源类型
3. 使用标准化后的 `normalized_url` 检查是否已存在，存在则直接返回已有 Item
4. 创建 `items` 记录，状态为 `pending`
5. 投递"内容解析任务"
6. 解析完成后写入标题、封面、摘要等基础信息
7. 创建一个 `webview` 类型的 `playback_asset`
8. 更新 `processing_status = ready`

## 10.2 抖音链路

### WebView 主路径
孩子端请求播放时：

1. 调用播放配置接口
2. 服务端优先返回 `playbackMode = webview`
3. 孩子端使用受控 WebView 打开目标页面
4. WebView 层执行导航拦截
5. 如果页面保持在目标播放上下文，则继续播放

### 失败判定
若出现以下任一情况，可进入兜底：
- 跳转到抖音 App 的深链
- 页面强制离开受控播放页
- 播放区域不可用
- 播放失败率持续超阈值

### 私有视频兜底路径
1. 系统识别当前 Item 需要兜底
2. 投递兜底处理任务
3. 生成/导入私有视频版本
4. 写入 `playback_assets(kind=private_video)`
5. 更新 Item 的 `playbackMode = private_video`
6. 孩子端下次打开时直接走原生视频播放器

## 10.3 X（Twitter）链路
与抖音链路不同，X 上的内容通常以图文或短视频为主：

1. 标准化 X 链接
2. 识别为 `sourceType = x`
3. 元信息抓取：提取推文正文作为摘要、提取媒体（图片/视频）作为封面或播放资源
4. 优先级：优先通过 WebView 展示推文原文（X 的嵌入式推文页面通常比较干净），如果有视频则尝试提取视频直链
5. 兜底：如果 WebView 不可用（例如被 X 拦截），降级为展示抓取到的正文摘要 + 静态封面图
6. 不需要像抖音那样做私有视频兜底（X 视频通常较短且不是核心内容形态）

## 10.4 微信视频号链路
微信视频号内容主要通过分享链接导入：

1. 识别微信视频号链接
2. 元信息抓取：标题、封面
3. 优先级：尝试 WebView 打开视频号播放页
4. 兜底：视频号 WebView 播放存在不确定性（依赖微信的页面策略），失败时降级为展示封面 + 提示家长"该内容在微信中观看效果更好"
5. MVP 阶段对视频号不做私有视频兜底，将风险标注为已知限制

## 10.5 普通网页链路
1. 识别为 `sourceType = web_article`
2. 元信息抓取：提取 og:title、og:image、og:description
3. `playbackMode = webview`（在受控 WebView 中展示网页）
4. WebView 安全策略：仅允许目标域名，拦截一切外链跳转

## 11. 孩子端受控播放页设计

## 11.1 WebView 控制策略
需要实现以下能力：
- 限制只允许目标域名 / 目标页面模式
- 拦截 scheme 跳转
- 拦截打开外部 App 请求
- 拦截新窗口
- 在非允许页面出现时立即返回安全页

## 11.2 原生视频播放器策略
兜底为私有视频时：
- 使用原生播放器组件
- 禁止跳出到外部播放器
- 记录播放进度
- 支持完成回调和继续观看

## 11.3 页面状态机
建议将播放页抽象为状态机：

- `loading`
- `webview_playing`
- `webview_blocked`
- `fallback_loading`
- `native_video_playing`
- `failed`
- `time_exceeded`：当日时长已用完

这样更方便处理播放策略切换。

## 12. API 草案

## 12.1 家庭与认证

### `POST /api/family/create`
家长端首次使用时创建家庭。

请求：
```json
{}
```

响应：
```json
{
  "familyId": "fam_123",
  "pairingCode": "A1B2C3",
  "parentToken": "eyJ..."
}
```

### `POST /api/family/pair`
孩子端输入配对码完成绑定。

请求：
```json
{
  "pairingCode": "A1B2C3",
  "childName": "小明",
  "deviceId": "android_device_xxx"
}
```

响应：
```json
{
  "childId": "child_1",
  "childToken": "eyJ..."
}
```

### 鉴权方案
- 家长端和孩子端分别持有 `parentToken` 和 `childToken`（JWT）
- 家长端 JWT 中包含 `familyId` 和 `role: parent`
- 孩子端 JWT 中包含 `childId`、`familyId` 和 `role: child`
- 所有 API 请求在 `Authorization: Bearer <token>` 中携带
- 后端中间件统一校验：孩子端只能访问已分配给自己的资源

## 12.2 家长端业务接口

### `POST /api/items/import`
请求：
```json
{
  "sourceUrl": "https://example.com/abc",
  "sourceType": "douyin"
}
```

响应：
```json
{
  "id": "item_123",
  "processingStatus": "pending",
  "isDuplicate": false
}
```

去重说明：
- 后端对 `sourceUrl` 做标准化（展开短链、去追踪参数）后得到 `normalizedUrl`
- 如果 `normalizedUrl` 已存在，返回已有 Item 并标记 `isDuplicate: true`
- 剪贴板识别和手动粘贴均调用此接口

### `GET /api/items`
返回家庭内容列表。

### `GET /api/items/:id`
返回内容详情、处理状态、播放资源。

### `POST /api/items/:id/assign`
请求：
```json
{
  "childIds": ["child_1", "child_2"]
}
```

### `POST /api/items/:id/revoke`
撤回已分配内容。

请求：
```json
{
  "childIds": ["child_1"]
}
```

### `POST /api/items/:id/reassign`
将被撤回的内容重新上线。

请求：
```json
{
  "childIds": ["child_1"]
}
```

### `PATCH /api/children/:id/settings`
家长设置孩子控制参数。

请求：
```json
{
  "dailyTimeLimitSec": 3600
}
```

### `GET /api/children/:id/usage`
查询孩子当日使用情况。

响应：
```json
{
  "childId": "child_1",
  "date": "2026-05-26",
  "totalPlaySec": 1200,
  "dailyLimitSec": 3600,
  "remainingSec": 2400
}
```

## 12.3 孩子端业务接口

### `GET /api/children/:id/feed`
返回该孩子可见内容列表（仅 `state = active` 的 Assignment）。

### `GET /api/items/:id/playback`
响应示例：
```json
{
  "itemId": "item_123",
  "playbackMode": "webview",
  "asset": {
    "kind": "webview",
    "url": "https://example.com/play"
  },
  "webviewPolicy": {
    "allowDomains": ["example.com"],
    "blockExternalScheme": true,
    "blockNewWindow": true
  },
  "usageRemainingSec": 2400
}
```

### `POST /api/play-events`
请求示例：
```json
{
  "childId": "child_1",
  "itemId": "item_123",
  "eventType": "start",
  "positionSec": 0,
  "occurredAt": "2026-05-26T14:30:00Z"
}
```

字段说明：
- `occurredAt`：客户端记录的事件实际发生时间，用于准确计算播放时长
- 服务端写入时自动记录 `created_at` 作为接收时间，用于排查延迟问题

## 13. 状态流转建议

## 13.1 内容处理状态
- `pending` → `ready`
- `pending` → `failed`（解析失败或重试耗尽）
- `failed` → `pending`（家长手动重试，后续扩展）

## 13.2 播放状态
- `unknown`
- `playable`
- `fallback_required`

## 13.3 播放资源状态
- `pending`
- `ready`
- `failed`

## 13.4 Assignment 状态
- `active` → `revoked`（家长撤回）
- `revoked` → `active`（家长重新上线）
- `active` → `expired`（长期未查看，MVP 暂不启用）
- `expired` → `active`（家长重新上线）

## 14. 安全与约束

### 14.1 身份认证
- 所有 API 通过 JWT Bearer Token 鉴权
- 家长 Token 具有家庭管理权限
- 孩子 Token 仅能访问分配给自己的资源
- Token 包含角色（parent/child）和家庭/孩子 ID，API 中间件根据角色做权限校验

### 14.2 链接处理
- 后端必须校验输入 URL
- 只允许受支持来源进入处理链路
- 避免将任意 URL 直接下发给孩子端 WebView

### 14.3 WebView 安全控制
- 配置 allowlist 域名
- 禁止任意 scheme 唤起
- 禁止开放式重定向
- 禁止文件协议和不必要能力

### 14.4 孩子身份隔离
- 孩子端只读取已分配内容
- 孩子端 API 只能访问自己的 play-events 和 usage
- 不暴露家长管理能力
- 孩子端不能修改自己的控制参数

### 14.5 客户端安全边界（Android）
- 孩子端应作为独立签名 APK，与家长端分开
- 孩子端不应存储 parentToken
- 反调试/反 root 检测为可选增强，MVP 暂不做

## 15. 错误处理规范
建议统一 API 错误响应格式：

```json
{
  "error": {
    "code": "USAGE_EXCEEDED",
    "message": "今日使用时长已用完"
  }
}
```

常见错误码：
| 错误码 | 说明 |
|--------|------|
| `INVALID_TOKEN` | Token 无效或已过期 |
| `FORBIDDEN` | 无权访问该资源（孩子访问非自己内容等） |
| `NOT_FOUND` | 资源不存在 |
| `DUPLICATE_URL` | 链接已存在（导入时返回，非错误） |
| `USAGE_EXCEEDED` | 当日使用时长已用完 |
| `LINK_UNSUPPORTED` | 链接来源不在支持范围内 |
| `LINK_RESOLVE_FAILED` | 链接解析失败 |
| `INVALID_PAIRING_CODE` | 配对码无效 |
| `ITEM_ALREADY_ASSIGNED` | 内容已分配给该孩子 |

## 16. 部署方案（MVP）
MVP 阶段推荐 Docker Compose 单机部署：

```
services:
  api:
    build: ./api
    ports: ["8080:8080"]
    depends_on: [postgres, redis]

  worker:
    build: ./worker
    depends_on: [postgres, redis]

  postgres:
    image: postgres:16-alpine

  redis:
    image: redis:7-alpine
```

后续可扩展：
- Kubernetes 部署
- 数据库读写分离
- CDN 加速静态资源

### 数据库 Migration
推荐使用 `golang-migrate` 管理数据库 schema 变更，保持版本化、可回滚。

## 17. 测试策略

### 单元测试
- 后端核心 logic 层覆盖率目标 > 70%
- 重点覆盖：URL 标准化与去重、来源类型识别、时长计算、Assignment 状态流转

### 集成测试
- 关键链路端到端验证：
  1. 创建家庭 → 配对孩子 → 导入内容 → 分配 → 孩子拉取 feed → 播放 → 上报事件
  2. 撤回 → 孩子 feed 不再出现
  3. 时长超限 → 播放被拒绝

### 孩子端测试
- WebView 导航拦截测试（模拟各类跳转场景）
- 播放页状态机流转测试

## 18. 监控与验证指标
建议首期就埋以下指标：
- 导入成功率
- Item 处理成功率
- WebView 首次播放成功率
- WebView 被阻断率
- 兜底触发率
- 原生视频播放成功率
- 孩子端内容打开率
- 收藏率

监控工具建议：Prometheus + Grafana（指标）+ 结构化日志（JSON 格式输出到 stdout，方便后续接入日志平台）。

## 19. MVP 实施顺序

### Phase 1：后端最小骨架
- 建 `families / items / children / assignments / playback_assets / play_events / daily_usage` 表
- `family/create + pair + import + assign + feed + playback + revoke + usage` 核心接口
- JWT 鉴权中间件
- Redis 队列框架

### Phase 2：家长端分享壳
- 接入系统分享入口
- 剪贴板监听
- 预览页
- 选择孩子并发送

### Phase 3：孩子端最小浏览能力
- 配对码输入绑定
- Feed 页
- 受控播放页
- WebView 导航拦截
- 时长限制校验

### Phase 4：抖音兜底能力
- 失败判定
- 私有视频资源接入
- 原生播放器切换

## 20. 当前推荐实现路径
如果目标是尽快落地，我建议：

1. **孩子端只做 Android**
2. **家长端只做分享壳，不做 Web 管理台**
3. **后端用 Go + PostgreSQL**
4. **异步任务用 Redis + Asynq（或同类方案）**
5. **首期支持抖音 / X / 微信视频号 / 网页导入**
6. **先重点验证抖音 WebView + 私有视频兜底**
7. **JWT 鉴权从 Phase 1 就带上，避免后续改造**

这样可以把产品风险压缩在最关键的地方：
- 家长分享入口是否顺手
- 孩子端受控播放是否够稳
- 抖音兜底是否足够无感
- 家庭绑定与鉴权是否简单可靠

## 21. 下一步可继续产出
基于这份技术方案，下一步可以继续补：
- 数据库 ER 图
- 前后端接口详细定义（含请求/响应字段类型与校验规则）
- 孩子端播放页状态机图
- MVP 开发任务拆解
