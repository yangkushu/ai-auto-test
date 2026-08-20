# 版本变更记录

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
