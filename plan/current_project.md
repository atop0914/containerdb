# ContainerDB — 2周开发计划

## 项目概述
A lightweight containerized database toolkit for Go development and testing. Spin up real databases in containers with a single function call.

## 技术栈
- Go 1.22+
- testcontainers-go (container management)
- go-sql-driver/mysql (MySQL driver)
- lib/pq (PostgreSQL driver)
- mattn/go-sqlite3 (SQLite driver)

## 2周计划

### Week 1 — 基础架构与核心功能

| 日期 | 任务 | 状态 |
|------|------|------|
| Day 1 | 项目初始化，搭建基础架构 | ✅ done |
| Day 2 | 实现 MySQL 容器封装，添加配置管理 | ✅ done |
| Day 3 | 实现 PostgreSQL 容器封装 | ✅ done |
| Day 4 | 实现 SQLite 辅助工具（in-memory/temp file） | ✅ done |
| Day 5 | 编写基础单元测试，覆盖核心 API | ✅ done |
| Day 6 | 添加 CLI 工具，支持启动/停止/状态查看 | ✅ done |
| Day 7 | 休息日 | — |

### Week 2 — 高级功能与完善

| 日期 | 任务 | 状态 |
|------|------|------|
| Day 8 | 添加连接池配置、健康检查增强 | ✅ done |
| Day 9 | 实现数据迁移辅助工具（migrate integration） | ✅ done |
| Day 10 | 添加 Docker Compose 兼容模式 | ✅ done |
| Day 11 | 完善文档，编写使用指南 | ✅ done |
| Day 12 | 添加性能基准测试 | ✅ done |
| Day 13 | 代码优化，清理 TODO，提交 v1.0.0 | ✅ done |
| Day 14 | 发布 Release，完善 CI/CD | ✅ done |

## GitHub 仓库
https://github.com/atop0914/containerdb-bootcamp

## 当前阶段
**🎉 2周计划全部完成！v1.0.0 已发布**

Day 14 完成内容：
- ✅ 修复 MySQL/PostgreSQL 测试不尊重 -short 标志的 bug（CI 无 Docker 环境可正确跳过）
- ✅ 添加 GitHub Actions CI 工作流（Go 1.22-1.25 矩阵测试、lint、5 平台交叉编译）
- ✅ 添加 GitHub Actions Release 工作流（打 tag 自动构建二进制、生成 changelog、创建 Release）
- ✅ 添加 golangci-lint 配置文件
- ✅ 更新 Makefile：添加 build/build-all/lint 目标，支持 ldflags 版本注入
- ✅ 更新 .gitignore 排除 dist/ 目录
- ✅ 合并 dev → main，创建 v1.0.0 tag
- ✅ 所有测试通过，代码已推送到 dev 和 main 分支

## 发布信息
- **v1.0.0** tag 已推送到 GitHub
- CI/CD 流水线已就绪，后续 push 到 main/dev 自动触发测试
- 打 v* tag 自动触发 Release 构建
