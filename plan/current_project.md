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
- ✅ 添加 `internal/compose` 包，compose 文件解析和生成
- ✅ 支持解析现有的 docker-compose.yml 文件
- ✅ 生成 MySQL 和 PostgreSQL 服务配置
- ✅ 添加 `pkg/compose` 包，高级 compose runner
- ✅ 支持 docker-compose up/down/ps/logs 操作
- ✅ 提供模板函数快速生成 compose 文件
- ✅ 编写完整的单元测试（21 个测试用例）

## 下一步
等待 Day 11 任务：完善文档，编写使用指南。
