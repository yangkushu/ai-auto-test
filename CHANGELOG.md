# 版本变更记录

## 0.2.0-dev.7 - 2026-08-21

- 新增必需交付物 `测试报告.md`，以中文提供测试结论、范围、用例明细、人工处理、测试数据影响和证据/限制；
- 报告生成失败或内容与状态表不一致时视为运行记录自检失败；
- 结果 Schema 升级为 `4`，避免与未生成测试报告的旧运行混合恢复。

## 0.2.0-dev.6 - 2026-08-21

- 将 `case-status.csv` 升级为中文表头、中文状态、中文布尔值和中文清理状态；JSONL 继续保持英文机器字段；
- 开发模式维持每条执行截图；正常模式只保存异常用例和 Preflight 失败的截图；
- 结果 Schema 升级为 `3`，避免与旧版英文状态表混合恢复。

## 0.2.0-dev.5 - 2026-08-21

- 将前置检查（Preflight）拆分为 Browser 能力、目标导航和应用识别阶段，并输出事实、推断和建议；
- 人类可读的运行输出统一使用中文，原始页面与用例证据保持原样；
- 执行记录和状态表增加写操作副作用、测试数据冲突与清理状态；
- 禁止已废弃的 `self_check` 事件，统一使用 `self_check_finished`；结果 Schema 升级为 `2`。

## 0.2.0-dev.4 - 2026-08-21

- 首版二进制交付范围收敛为 Windows x64 与 Linux x64；
- CI、Release、安装校验和 Skill 启动校验同步按双平台执行；
- macOS 与 ARM64 支持延后至稳定回归后评估。

## 0.2.0-dev.3 - 2026-08-20

- 将 `execute-test-cases` 设为唯一用户入口，由 Skill 内部调用 Cursor Browser；
- 在创建 run 前拒绝无效路径、占位值、无效 URL 和未定义账号；
- Preflight 增加应用健康与成功标志检查，503 等错误页不再进入业务执行；
- 增加 Go 标准库实现的跨平台 `ai-auto-test-store`，确定性追加和校验 JSONL；
- Preflight 失败时保留全部用例为 `pending`，支持环境恢复后继续。
- 增加 GitHub Actions 六平台构建、交付校验和 tag Release；
- 增加 GitHub Skill 导入、自然语言代装、安装验收和分发架构文档。

## 0.2.0-dev.2 - 2026-08-20

- 在执行前和追加前检查 execution ID 与“用例 ID + attempt”唯一性；
- 增加 `self_check` 和 `run_valid`，审计硬失败不得降级为 warning；
- 开发模式要求来源步骤与步骤事件一一对应，不得压缩为用例级汇总；
- 增加脱敏的 `browser_action_started`/`browser_action_finished` 事件；
- 明确目标页面缺少预期控件属于 `failed`，能力或数据前置不足才属于 `blocked`。

## 0.2.0-dev.1 - 2026-08-20

- 增加 `mode=normal|development` 运行参数；
- 增加正常模式与开发模式，模式只影响可观测信息；
- 增加只追加的 `run-events.jsonl`；
- 在运行结果中记录 Skill 版本、Schema 版本和生效模式；
- 增加开发模式事件示例和闭环验证步骤。

## 0.1.0 - 2026-08-20

- 建立 Cursor Browser-only MVP；
- 跑通登录、门店页、结果落盘、受控恢复和复测追加；
- 首版仅依赖 Cursor，不要求安装 Node.js、npm、Playwright 或额外 Browser MCP。
