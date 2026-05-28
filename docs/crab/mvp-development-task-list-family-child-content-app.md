# 家庭内部儿童内容浏览工具 MVP 开发任务清单

## 1. 文档信息
- 文档名称：家庭内部儿童内容浏览工具 MVP 开发任务清单
- 版本：v0.2
- 日期：2026-05-26
- 对应 PRD：`docs/prd-family-child-content-app.md`
- 对应技术方案：`docs/technical-design-family-child-content-app.md`

## 2. 目标
把当前 MVP 技术方案拆成可执行的开发任务，确保第一版只完成下面这条主链路：

> 家长从抖音 / X / 微信视频号 / 网页分享内容 → 系统导入并处理 → 家长确认发送给孩子 → 孩子在 Android App 中安全浏览和播放。

## 3. 开发原则
1. 先打通主链路，再补体验。
2. 先支持 Android，iOS 暂不进入首期。
3. 先做分享壳，不做家长 Web 管理台。
4. 先做 MVP 核心表，不扩展非 MVP schema。
5. 先重点验证抖音播放与兜底，其他来源先完成导入和基础展示。
6. 每日时长限制依赖设备系统设置，本产品 MVP 不实现。

## 4. 里程碑

### Milestone 1：后端最小骨架可用
完成后，系统可以接收分享链接、建内容记录、返回 feed 与播放配置。

### Milestone 2：家长端分享壳可用
完成后，家长可以从系统分享面板打开分享壳，看到最小预览并把内容发送给孩子。

### Milestone 3：孩子端 Android 可浏览播放
完成后，孩子可以看到 feed，并在受控容器中播放内容。

### Milestone 4：抖音兜底能力可用
完成后，抖音 WebView 失败时可以切换到私有视频方案。

## 5. 风险等级说明
| 等级 | 说明 |
|------|------|
| 🟢 低风险 | 标准实现，预期顺利 |
| 🟡 中风险 | 需要关注，可能遇到问题 |
| 🔴 高风险 | 技术难点，需提前验证或准备方案 |

## 6. 任务拆解

### Phase 1: 基础工程

#### T1. 初始化 Go 后端工程
**风险等级**: 🟢 低风险

**内容**
- 创建 Go 服务目录结构（cmd/api, internal/models, internal/handlers, internal/middleware 等）
- 接入路由框架（建议 Chi 或 Gin）
- 配置基础启动方式
- 接入配置加载（环境变量）
- 增加健康检查接口

**产出**
- 可启动的 API 服务
- `GET /healthz`

**验收标准**
- 本地可启动
- 健康检查返回 200
- 配置可通过环境变量注入

---

#### T2. 初始化 PostgreSQL / Redis / 对象存储连接
**风险等级**: 🟢 低风险

**内容**
- 接入 PostgreSQL（gorm 或 sqlx）
- 接入 Redis
- 抽象对象存储客户端（兼容 S3）
- 完成本地开发环境 docker-compose 配置

**产出**
- 数据库连接层
- Redis 连接层
- 对象存储封装
- docker-compose.yml

**验收标准**
- 服务启动时能完成依赖连接
- 连接失败有清晰报错

---

#### T3. 建立 MVP 核心表
**风险等级**: 🟢 低风险

**内容**
创建以下表（使用 golang-migrate 管理 migration）：

**families 表**
- id (UUID, PK)
- pairing_code (VARCHAR(8), 唯一索引)
- created_at (TIMESTAMP)

**children 表**
- id (UUID, PK)
- family_id (UUID, FK → families)
- name (VARCHAR)
- avatar_url (VARCHAR, nullable)
- device_id (VARCHAR, nullable)
- status (VARCHAR: active/inactive)
- created_at (TIMESTAMP)

**items 表**
- id (UUID, PK)
- family_id (UUID, FK → families)
- title (VARCHAR)
- cover_url (VARCHAR, nullable)
- summary (TEXT, nullable)
- source_type (VARCHAR: douyin/x/wechat_channels/web_article)
- source_url (TEXT)
- normalized_url (TEXT, 索引)
- playback_mode (VARCHAR: webview/private_video)
- processing_status (VARCHAR: pending/ready/failed)
- playback_status (VARCHAR: unknown/playable/fallback_required)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

**assignments 表**
- id (UUID, PK)
- item_id (UUID, FK → items)
- child_id (UUID, FK → children)
- assigned_by (VARCHAR: parent)
- assigned_at (TIMESTAMP)
- state (VARCHAR: active/revoked)

**playback_assets 表**
- id (UUID, PK)
- item_id (UUID, FK → items)
- kind (VARCHAR: webview/private_video)
- asset_url (TEXT)
- status (VARCHAR: pending/ready/failed)
- meta_json (JSONB, nullable)
- created_at (TIMESTAMP)

**play_events 表**
- id (UUID, PK)
- child_id (UUID, FK → children)
- item_id (UUID, FK → items)
- event_type (VARCHAR: open/start/pause/complete/favorite)
- position_sec (INTEGER)
- occurred_at (TIMESTAMP)
- created_at (TIMESTAMP)

**daily_usage 表**
- id (UUID, PK)
- child_id (UUID, FK → children)
- date (DATE, 索引)
- total_play_sec (INTEGER)
- updated_at (TIMESTAMP)

**验收标准**
- migration 可重复执行
- 本地数据库建表成功
- 所有外键约束正确

---

#### T4. 准备最小种子数据
**风险等级**: 🟢 低风险

**内容**
- 插入 1 个测试家庭，配对码如 `TEST01`
- 插入 1~2 个测试孩子账号
- 准备 3~5 条测试 Item 数据（包含不同 source_type）

**验收标准**
- 本地环境可直接看到 feed 数据
- 不依赖手工写 SQL 才能联调

**验收标准**
- 本地可启动
- 健康检查返回 200
- 配置可通过环境变量注入

---

### Phase 2: 家庭管理与认证

#### T5. 实现 family/create 接口
**风险等级**: 🟢 低风险

**接口**: `POST /api/family/create`

**内容**
- 创建新家庭，生成随机 8 位配对码
- 返回 familyId 和 parentToken（JWT）
- JWT 包含 familyId 和 role: parent

**验收标准**
- 同一家庭不会重复创建（根据请求特征）
- 返回有效的 JWT token

---

#### T6. 实现 family/pair 接口
**风险等级**: 🟢 低风险

**接口**: `POST /api/family/pair`

**内容**
- 接收配对码、孩子名称、设备 ID
- 校验配对码有效性
- 创建 child 记录
- 返回 childToken（JWT）

**验收标准**
- 正确配对码可成功绑定
- 错误配对码返回明确错误
- 同一设备可重新绑定（更新 child 记录）

---

#### T7. 实现 children 管理接口
**风险等级**: 🟢 低风险

**接口**:
- `GET /api/children` - 获取家庭下所有孩子列表
- `POST /api/children` - 添加孩子（可选，MVP 可在 pair 时自动创建）
- `GET /api/children/:id` - 获取孩子详情
- `DELETE /api/children/:id` - 解绑孩子（软删除）

**验收标准**
- 家长可查看和管理自己家庭下的孩子
- 孩子端无法访问其他家庭的数据

---

### Phase 3: 核心业务 API

#### T8. 实现导入接口 POST /api/items/import
**风险等级**: 🟡 中风险（URL 标准化有边界情况）

**接口**: `POST /api/items/import`

**内容**
1. **T8a URL 标准化**
   - 展开短链（HTTP 请求获取最终 URL）
   - 去除追踪参数（utm_*, fbclid, spm 等）
   - 生成 normalized_url

2. **T8b 去重检查**
   - 根据 normalized_url 查询 items 表
   - 存在则返回已有 Item，标记 isDuplicate: true

3. **T8c 来源识别 + 创建**
   - 根据域名/URL 模式识别来源类型
   - 创建 Item 记录（状态: pending）
   - 投递异步处理任务

**验收标准**
- 传入有效链接可返回 item id
- 非支持来源返回明确错误（LINK_UNSUPPORTED）
- 重复 URL 可正确去重或复用已有 Item
- normalized_url 在去重中正确生效

---

#### T9. 实现家庭 feed 接口 GET /api/items
**风险等级**: 🟢 低风险

**接口**: `GET /api/items`

**内容**
- 查询该家庭可见内容（按 familyId 查询 items）
- 只返回 items.processing_status = ready 的内容
- 返回 Item 基础信息（id, title, cover_url, source_type, playback_mode）
- 按 created_at 倒序排序

**验收标准**
- 返回结果只包含状态为 ready 的内容
- 排序符合预期（最新投喂在前）
- 家庭内所有孩子可见

---

#### T10. 实现播放配置接口 GET /api/items/:id/playback
**风险等级**: 🟢 低风险

**接口**: `GET /api/items/:id/playback`

**内容**
- 查询当前 Item 的播放配置
- 返回 playback_mode（webview/private_video）
- 返回播放资源 URL（从 playback_assets 获取）
- 返回 WebView 安全策略参数

**WebView 安全策略参数**:
```json
{
  "allowDomains": ["域名白名单"],
  "blockExternalScheme": true,
  "blockNewWindow": true,
  "requiredUrls": ["允许的 URL 模式列表"]
}
```

**验收标准**
- webview 类型可返回 webview 资源
- private_video 类型可返回私有视频资源
- 资源不存在时返回明确的错误

---

#### T12. 实现事件上报接口 POST /api/play-events
**风险等级**: 🟢 低风险

**接口**: `POST /api/play-events`

**内容**
- 接收事件：open, start, pause, complete, favorite
- 记录 position_sec（播放位置）
- 记录 occurred_at（客户端时间）
- 累加 daily_usage.total_play_sec

**验收标准**
- 孩子端可成功上报事件
- 服务端能正确记录事件
- daily_usage 可正确累计

---

#### T13. 初始化 Worker 与队列
**风险等级**: 🟢 低风险

**内容**
- 接入 Redis 队列（建议 Asynq）
- 建立任务类型：
  - `item_ingest`: 内容解析
  - `fallback_generate`: 兜底处理
- 实现 Worker 启动和优雅退出
- 实现任务重试策略（最多 3 次，指数退避）

**验收标准**
- API 可成功投递任务
- Worker 可正常消费任务
- 任务失败可正确重试

---

#### T14. 实现内容解析任务
**风险等级**: 🔴 高风险（抖音解析不确定性高）

**内容**
1. **T14a 通用网页解析**（web_article）
   - 抓取 og:title, og:image, og:description
   - 状态更新为 ready
   - 生成 webview 类型的 playback_asset

2. **T14b 抖音解析**（P0，高优先级）
   - 抓取抖音分享页元信息
   - 解析视频 ID、标题、封面
   - 生成 webview 类型的 playback_asset
   - ⚠️ 风险：抖音可能拦截或返回非播放页面

3. **T14c X 解析**（P1，降级处理）
   - 尝试解析 X 推文
   - 降级方案：使用通用网页解析
   - ⚠️ 风险：X 可能阻止爬取

4. **T14d 微信视频号解析**（P1，降级处理）
   - 尝试解析视频号分享链接
   - 降级方案：展示标题 + 提示"建议在微信中观看"
   - ⚠️ 风险：视频号链接通常需要微信环境

**验收标准**
- 各来源链接能进入对应处理分支
- 成功时 Item 状态更新为 ready
- 失败时 Item 状态更新为 failed
- 重试耗尽后不再重试

**MVP 建议**: 首期只完整实现 T14a + T14b，其他来源使用降级方案

---

### Phase 4: 家长端 Android

#### T15. 初始化 Android 分享壳工程
**风险等级**: 🟡 中风险（ShareTarget 配置复杂）

**内容**
1. **T15a 项目脚手架**
   - 创建 Android 项目（Kotlin）
   - 配置 Gradle 依赖
   - 搭建基础架构（MVVM 或 MVP）

2. **T15b ShareTarget 配置**
   - 配置 AndroidManifest.xml 的 share-target
   - 声明支持接收的 URL 类型（http/https）
   - 配置 IntentFilter

3. **T15c 基础路由**
   - 单 Activity + Fragment 架构
   - 基础路由跳转

4. **T15d 网络层**
   - Retrofit + OkHttp
   - 基础鉴权拦截器

**产出**
- 可启动的 APK
- 能从系统分享面板唤起

**验收标准**
- 从抖音/浏览器/其他支持分享的 App 中可看到本应用入口
- 应用可被正常唤起并获取分享内容

---

#### T16. 实现分享确认页
**风险等级**: 🟢 低风险

**内容**
- 读取系统分享进来的 URL
- 展示最小预览（标题、封面占位）
- 展示孩子选择器（下拉或列表）
- 点击"发送给孩子"按钮
- 处理 loading / success / failed 状态

**验收标准**
- 家长可以完成一次完整发送
- 页面信息最少但足够确认内容
- 网络异常有友好提示

---

#### T17. 对接后端接口
**风险等级**: 🟢 低风险

**内容**
- 调用 import 接口导入内容
- 调用 assign 接口分配给孩子
- 处理接口错误和重试逻辑
- 成功后返回系统分享完成

**验收标准**
- 一次分享操作能在后端成功生成 item + assignment
- 重复分享可正确识别并复用

---

### Phase 5: 孩子端 Android

#### T18. 初始化孩子端 Android 工程
**风险等级**: 🟢 低风险

**内容**
1. **T18a 项目脚手架**
   - 创建独立 Android App（与家长端分开）
   - MVVM 架构
   - 网络层（Retrofit + OkHttp）

2. **T18b 配对绑定页**
   - 输入配对码
   - 输入孩子昵称
   - 调用 pair 接口完成绑定
   - 保存 childToken

**验收标准**
- App 可启动
- 能完成家庭绑定
- 绑定后能访问 feed

---

#### T19. 实现 Feed 页
**风险等级**: 🟢 低风险

**内容**
- 调用 `GET /api/children/:id/feed`
- 展示内容卡片列表（封面、标题、来源标签）
- 下拉刷新、上拉加载更多
- 空状态和加载状态 UI
- 点击卡片进入播放页

**验收标准**
- 孩子可以看到已分配内容列表
- 空状态和加载状态可用
- 卡片点击可正常跳转

---

#### T20. 实现受控播放页
**风险等级**: 🔴 高风险（WebView 安全壳技术难点）

**内容**
1. **T20a WebView 基础**
   - WebView 初始化
   - 加载目标 URL
   - 基本导航支持

2. **T20b 域名白名单**
   - 只允许目标域名加载
   - 根据 source_type 配置不同策略

3. **T20c Scheme 拦截**
   - 拦截 App 唤起（douyin://, weixin://, etc）
   - 拦截外部浏览器跳转
   - 拦截 file:// 协议

4. **T20d 导航拦截**
   - 拦截新窗口打开
   - 拦截非目标域名跳转
   - 页面离开检测

5. **T20e 失败判定**
   - 检测到页面离开受控环境 → 触发兜底
   - 页面加载失败 → 触发兜底
   - 播放区域不可用 → 触发兜底

**播放页状态机**:
```
loading → webview_playing → [webview_blocked] → fallback_loading → native_video_playing
                ↓
         [failed] → failed
```

**验收标准**
- WebView 可在受控范围内打开目标内容
- 深链跳出被拦截
- 播放失败可正确识别并触发兜底

---

#### T21. 接入事件上报
**风险等级**: 🟢 低风险

**内容**
- 页面打开时上报 `open`
- 开始播放时上报 `start`
- 播放完成时上报 `complete`
- 可选：收藏时上报 `favorite`

**验收标准**
- 后端能看到完整事件链路
- 事件时间戳准确

---

### Phase 6: 抖音兜底能力

#### T22. 实现私有视频兜底任务
**风险等级**: 🔴 高风险（视频下载技术难点）

**内容**
1. **T22a 兜底触发检测**
   - 接收播放页上报的兜底请求
   - 更新 item.playback_status = fallback_required
   - 投递 fallback_generate 任务

2. **T22b 私有视频生成**
   - Worker 处理兜底任务
   - 尝试获取抖音视频直链（调研第三方工具）
   - 下载视频到对象存储
   - 生成 playback_asset(kind=private_video)
   - 更新 item.playback_mode = private_video

**技术挑战**:
- 抖音视频下载需要解析签名，无法直接通过 URL 下载
- 可能需要：第三方解析 API / 模拟移动端请求 / 购买商业方案

**验收标准**
- 兜底任务可被正确触发
- 私有视频资源可生成成功
- 孩子端可获取到 private_video 资源

**MVP 建议**: 如果私有视频方案短期内无法实现，可先用降级方案：展示封面 + 提示家长该内容在抖音 App 中观看效果更好

---

#### T23. 孩子端原生视频播放器
**风险等级**: 🟡 中风险

**内容**
- 当 playback_mode = private_video 时使用原生播放器
- 使用 ExoPlayer 或系统 MediaPlayer
- 禁止跳转到外部播放器
- 记录播放进度
- 支持继续播放

**验收标准**
- 抖音失败场景下仍能完成播放
- 播放过程可控，不跳出 App

---

## 7. 泳道拆分与依赖关系

### 7.1 依赖关系图

```
Phase 1: T1-T4 (基础设施)
    │
    ├── T1 Go 工程 ──┐
    ├── T2 数据库连接 ┤
    ├── T3 建表 ─────┼──→ Phase 2: T5-T7 (家庭管理)
    └── T4 种子数据 ──┘           │
                                 ├── T5 family/create
                                 ├── T6 family/pair
                                 └── T7 children 管理
                                        │
                                        └──→ Phase 3: T8-T14 (核心 API)
                                                      │
                              ┌───────────────────────┼───────────────────────┐
                              │                       │                       │
                         Phase 4                  Phase 4               Phase 5
                        (家长端 T15-T17)       (孩子端 T18-T21)      (兜底 T22-T23)
                              │                       │                       │
                         Checkpoint A             Checkpoint B          Checkpoint D
                         (导入链路)              (分发+播放链路)         (兜底链路)
                              │                       │
                              └──────────┬────────────┘
                                         │
                                    Checkpoint C
                                    (端到端联调)
```

### 7.2 阶段泳道

| 阶段 | 任务 | 说明 |
|------|------|------|
| **Phase 1** | T1-T4 | 基础工程（必须先完成） |
| **Phase 2** | T5-T7 | 家庭管理（必须先完成） |
| **Phase 3** | T8-T14 | 核心 API + Worker |
| **Phase 4** | T15-T17 | 家长端 Android |
| **Phase 5** | T18-T21 | 孩子端 Android |
| **Phase 6** | T22-T23 | 抖音兜底 |

### 7.3 并行建议

**可以并行的部分**:
- T1-T4 完成后，家长端 T15、孩子端 T18 可以并行启动
- T5-T7 完成后，家长端 T16、孩子端 T19 可以并行联调
- T8-T14 完成后，家长端 T17、孩子端 T19-T21 可以并行联调

**不建议并行的部分**:
- T5 之前不要启动 T17（依赖家庭创建和配对）
- T7 之前不要启动 T19（依赖孩子列表查询）
- T8 之前不要启动 T17（依赖 import 接口）
- T14 之前不要启动 T20（依赖播放接口和 Worker）

---

## 8. 联调检查点

### Checkpoint A：导入链路打通
**前置条件**: T1-T4, T5, T8, T13, T14 完成

**验证步骤**:
1. 家长端分享一个抖音链接
2. 后端正确识别来源类型
3. Item 状态从 pending → ready
4. 后台日志显示解析任务完成

**成功标准**:
- [ ] 家长端分享成功
- [ ] 后端创建 Item
- [ ] Worker 处理完成
- [ ] Item 状态更新为 ready

---

### Checkpoint B：分发链路打通
**前置条件**: Checkpoint A + T6, T9, T10 完成

**验证步骤**:
1. 家长端选择一个孩子并发送内容
2. 查看后端 assignment 记录
3. 孩子端登录并查看 feed

**成功标准**:
- [ ] 分配成功创建
- [ ] 孩子端 feed 可看到该内容
- [ ] 内容卡片展示正确

---

### Checkpoint C：播放链路打通
**前置条件**: Checkpoint B + T11, T18, T20, T21 完成

**验证步骤**:
1. 孩子端点击内容卡片
2. 播放页加载
3. WebView 打开目标内容
4. 上报播放事件

**成功标准**:
- [ ] 播放页正常加载
- [ ] WebView 在受控范围内打开
- [ ] 播放事件可上报
- [ ] 导航拦截生效

---

### Checkpoint D：抖音兜底打通
**前置条件**: T22, T23 完成

**验证步骤**:
1. 模拟 WebView 播放失败
2. 触发兜底请求
3. 观察私有视频生成
4. 验证原生播放器切换

**成功标准**:
- [ ] 失败可被识别
- [ ] 兜底任务被投递
- [ ] 私有视频可生成（如果技术方案可行）
- [ ] 播放器可正常切换

---

## 9. 风险缓解措施

### 🔴 高风险缓解

| 风险点 | 缓解措施 |
|--------|----------|
| **抖音 WebView 被拦截** | 1. 提前调研抖音 WebView 限制 2. 准备降级方案 3. 监控播放成功率 |
| **私有视频下载技术可行性** | 1. MVP 前期调研第三方方案 2. 准备降级文案 3. 考虑购买商业 API |
| **WebView 安全壳稳定性** | 1. 指定测试机型范围 2. 准备多种 WebView 策略 3. Android 不同版本适配 |
| **视频号解析失败** | MVP 阶段直接降级为"建议微信中观看" |

### 🟡 中风险缓解

| 风险点 | 缓解措施 |
|--------|----------|
| **URL 标准化复杂性** | 1. 先实现基础标准化 2. 收集边界 case 3. 逐步完善 |
| **ShareTarget 配置** | 参考成熟开源项目配置 2. 提前在测试机验证 |
| **配对码安全性** | 6 位码 + 设备 ID 绑定 2. 可考虑加时效性 |

---

## 10. 建议延期到第二阶段的内容

以下内容**不建议**进入第一版开发：
- 家长 Web 管理台（先做分享壳）
- iOS 分享入口（先做 Android）
- iOS 孩子端 App
- 分类页
- 收藏页
- 最近看过
- 标签系统
- 完整埋点事件表
- 撤回 / 重发 / 重上线完整运营能力
- 多来源精细化差异处理
- 离线缓存

---

## 11. 第一版完成定义

满足以下条件，即可认为 MVP 第一版完成：

| # | 条件 | 验证方式 |
|---|------|----------|
| 1 | 家长可以从 Android 系统分享面板把链接送进应用 | 实际演示 |
| 2 | 后端可以识别来源、创建内容、分配给孩子 | API 测试 |
| 3 | 孩子端 Android 可以看到 feed | 实际演示 |
| 4 | 孩子端可以在受控容器中播放内容 | 实际演示 |
| 5 | 抖音内容失败时有降级处理 | 演示或说明 |
| 6 | 主链路可以完成一次端到端演示 | 完整演示 |

---

## 12. 快速参考：任务清单汇总

| Phase | 任务 | 风险 | 建议工期 |
|-------|------|------|----------|
| **1** | T1 Go 工程初始化 | 🟢 | 0.5 天 |
| **1** | T2 数据库/存储连接 | 🟢 | 0.5 天 |
| **1** | T3 建立 MVP 核心表 | 🟢 | 1 天 |
| **1** | T4 种子数据 | 🟢 | 0.5 天 |
| **2** | T5 family/create | 🟢 | 0.5 天 |
| **2** | T6 family/pair | 🟢 | 0.5 天 |
| **2** | T7 children 管理 | 🟢 | 0.5 天 |
| **3** | T8 导入接口 | 🟡 | 1 天 |
| **3** | T9 分配接口 | 🟢 | 0.5 天 |
| **3** | T10 feed 接口 | 🟢 | 0.5 天 |
| **3** | T11 playback 接口 | 🟢 | 0.5 天 |
| **3** | T12 事件上报 | 🟢 | 0.5 天 |
| **3** | T13 Worker 初始化 | 🟢 | 1 天 |
| **3** | T14 内容解析任务 | 🔴 | 2-3 天 |
| **4** | T15 Android 工程 | 🟡 | 1 天 |
| **4** | T16 分享确认页 | 🟢 | 1 天 |
| **4** | T17 对接接口 | 🟢 | 0.5 天 |
| **5** | T18 孩子端工程 | 🟢 | 1 天 |
| **5** | T19 Feed 页 | 🟢 | 1 天 |
| **5** | T20 WebView 播放页 | 🔴 | 2-3 天 |
| **5** | T21 事件上报 | 🟢 | 0.5 天 |
| **6** | T22 私有视频兜底 | 🔴 | 2-3 天 |
| **6** | T23 原生播放器 | 🟡 | 1 天 |

**预估总工期**: 约 17-21 天（不含风险缓冲）

**建议增加 20-30% 风险缓冲时间**
