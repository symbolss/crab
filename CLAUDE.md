# Crab - 家庭内部儿童内容浏览工具

面向家庭内部使用的儿童内容浏览工具。家长发现内容并分享给孩子，孩子在专属 App 内安全浏览。

## 项目结构

```
code/
├── parent-app/    # 家长端 (Android 分享壳)
├── child-app/     # 孩子端 (Android 原生)
└── server/        # 后端服务 (Go)

docs/crab/         # 项目文档
├── prd-*.md       # 需求文档
├── technical-design-*.md  # 技术设计
└── mvp-development-task-list-*.md  # 任务列表
```

## 技术栈

| 端 | 技术 |
|----|------|
| 后端 | Go + Chi + PostgreSQL + Redis |
| 孩子端 | Android (Kotlin + MVVM + Hilt) |
| 家长端 | Android (分享壳) |

## 核心功能

- 家长分享内容链接 → 系统处理 → 孩子端安全播放
- WebView 安全壳：拦截深链、外跳、非白名单域名
- 私有视频兜底：WebView 失败时切换原生播放器
- 每日使用时长控制

## 开发流程

按照 ECC 开发工作流：
1. 规划先行 → TDD 开发 → 代码审查 → 提交
2. 测试覆盖率 ≥ 80%
3. 使用 planner、tdd-guide、code-reviewer 等 agents

## 相关文档

- [PRD](docs/crab/prd-family-child-content-app.md)
- [技术设计](docs/crab/technical-design-family-child-content-app.md)
- [任务列表](docs/crab/mvp-development-task-list-family-child-content-app.md)
