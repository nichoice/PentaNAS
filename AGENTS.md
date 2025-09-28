# Repository Guidelines

本指南面向贡献者，汇总 PNAS 仓库的日常协作规范。贡献代码前建议通读 `README.md` 及 `docs/` 下的设计文档，确保理解系统边界与术语。

## Project Structure & Module Organization
- `main.go` 启动 HTTP 服务；核心业务逻辑集中在 `internal/`，其中 `controllers` 负责请求处理，`services` 封装领域用例，`routes` 注册 Gin 路由，持久化适配器位于 `infrastructure` 与 `database`。
- 运行时配置来自 `config.yaml`（由 `config.example.yaml` 拷贝后私有化）；数据库迁移脚本保存在 `migrations/`；长期文件与导出内容统一写入 `uploads/`。
- `docs/` 收录参考资料，`pkg/` 预留可复用模块；依赖硬件的实验代码（如 `tests/lvm_test.go`）集中在 `tests/` 目录，避免污染通用测试。静态资源或示例数据请放入 `uploads/` 的子目录以便追踪。

## Build, Test, and Development Commands
- `make build` 在 `build/pnas` 输出当前平台二进制并注入版本元数据；`make build-linux-*` 支持交叉编译多架构制品。
- `make run` 本地启动 API；`make run-dev` 追加 `--debug` 方便排查；需要发布包时使用 `make package`。若不确定命令用途，可运行 `make help` 查看简要说明。
- `make test` 执行 `go test ./...`；`make test-cover` 生成 `coverage.out` 与 `coverage.html`；提交前务必运行 `make fmt vet tidy` 保证格式、静态分析与依赖整洁。首次开发者可通过 `go env` 确认本地 Go 版本已符合 `go.mod` 的要求。

## Coding Style & Naming Conventions
- 所有 Go 文件使用 `gofmt`（`make fmt`）保持官方风格，采用制表符缩进，按标准导入顺序分组。
- 包名保持简短小写且语义明确（示例：`internal/domain`、`internal/services`），文件拆分时遵循 snake_case。公共 API 置于 `pkg/` 或 `internal/interfaces`。
- 导出结构体、接口使用 PascalCase；私有函数与变量使用 camelCase；错误变量以 `err...` 命名，并在返回前包裹上下文。日志字段统一使用英文蛇形键名，方便与现有分析工具对接。

## Testing Guidelines
- 优先编写表驱动单元测试并与源码同目录存放；跨模块或设备敏感场景收敛到 `tests/`。测试文件命名遵循 `*_test.go`，并合理使用子测试突出场景差异。
- `tests/lvm_test.go` 因依赖 LVM 设备已在 CI 中禁用，如需启用请在文件头注记前置条件并改为手动运行，避免阻塞常规流水线。
- 通过 `make test-cover` 跟踪覆盖率，对异常分支、配置降级等路径补充断言，避免回归。关键函数应模拟边界值与错误返回，确保日志与告警路径可用。

## Commit & Pull Request Guidelines
- 提交信息遵循 Conventional Commits（如 `feat:`, `build:`, `tests:`），主题限定 72 字符内，必要时在正文补充细节与影响范围。变更多文件时推荐在正文中添加 bullet 梳理亮点。
- 提交 PR 前先与最新 `develop` 同步并执行 `make fmt test`；在描述中关联需求或缺陷编号，说明变更模块、配置及潜在风险。PR 需通过 CI 检查后再请求评审。
- 用户界面或 API 行为变更需附截图或示例响应；涉及迁移或配置调整务必在 PR 中明确提醒评审者。若需要配套脚本或运维指引，请附在 `docs/` 并在 PR 中链接。

## Configuration & Environment
- 首次开发时复制 `config.example.yaml` 为本地 `config.yaml` 并填充密钥，切勿提交敏感字段；私有配置可借助 `.gitignore` 保护。
- 数据库连接封装在 `internal/database`，日志默认配置位于 `internal/logging`；部署前确认日志级别和输出路径，如需采集到外部系统请同步运维团队。
- 如需自定义构建元数据，可在运行 `make build` 前设置 `VERSION`、`COMMIT`、`DATE` 环境变量，确保发布信息准确。
