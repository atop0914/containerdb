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
| Day 11 | 完善文档，编写使用指南 | todo |
| Day 12 | 添加性能基准测试 | todo |
| Day 13 | 代码优化，清理 TODO，提交 v1.0.0 | todo |
| Day 14 | 发布 Release，完善 CI/CD | todo |

## GitHub 仓库
https://github.com/atop0914/containerdb-bootcamp

## 当前阶段
**Week 2 - Day 10 完成**

Day 10 完成内容：
- ✅ 添加 Docker Compose V2 (docker compose) 和 V1 (docker-compose) 双版本支持
- ✅ 为 compose 服务添加健康检查（HealthCheck）配置支持
- ✅ 添加 depends_on 依赖关系和 restart 重启策略支持
- ✅ 新增 CLI 命令：containerdb compose init/up/down/status/logs
- ✅ 生成包含健康检查的 MySQL 和 PostgreSQL 模板
- ✅ 添加 GenerateMySQLServiceWithHealthCheck/GeneratePostgresServiceWithHealthCheck
- ✅ 在 pkg/compose 中重新导出内部类型，提供公共 API
- ✅ 添加 compose 示例（examples/compose/）
- ✅ 修复 examples 目录结构（分离 main 包）
- ✅ 所有测试通过，代码已推送到 dev 分支

## 下一步
等待 Day 11 任务：完善文档，编写使用指南。
