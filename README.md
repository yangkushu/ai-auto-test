# ai-auto-test

`ai-auto-test` 是一个供测试人员使用的 AI Web 测试编排项目。它读取已有测试用例，先进行无截图的快速验证，再按规则调用 Browser Agent 验证复杂或抽样用例，记录每条用例的结论与证据，并生成需要人工处理的清单。

当前目标不是替代人工测试，也不是建立完整测试平台，而是先用 Cursor Agent 跑通一条可重复、可审计的最小流水线。

## 快速开始

### 安装

在 Cursor 中直接执行：

```text
安装或更新这个 Skill，并严格按安装说明操作：https://github.com/yangkushu/ai-auto-test/blob/master/docs/installation.md
```

### 运行

安装完成并新建 Agent 会话后，在 Cursor 中执行：

```text
/execute-test-cases 执行这个测试用例
```

`execute-test-cases` 为实际 Skill 名称。请在同一条消息中提供测试用例、测试环境 URL 和测试账号配置。

测试完成后，结果位于当前项目的 `.ai-auto-test/results/<run-id>/`；优先查看其中面向人的 `测试报告.md`，再按需查看 `summary.md`。

## 详细说明

- [安装说明](docs/installation.md)：安装验收、更新、二进制与平台支持。
- [使用说明](docs/usage.md)：输入格式、开发模式、结果检查和恢复复测。
- [两用例闭环验证提示词](prompts/cursor-closed-loop-validation.md)：可复制的完整验证 Prompt。

## 当前开发版本

```text
只读测试用例
    ↓
Playwright MCP 快速验证（全量）
    ↓
Cursor 原生 Browser 验证（强制用例 + 自动抽样）
    ↓
两阶段状态 + 追加式执行记录 + 报告 + 人工处理清单
```

- 默认 `stage=auto`：全量快速验证后，约 50%～70% 的用例进入 Browser 验证；高风险、复杂、写操作和快速阶段异常用例强制进入。
- `stage=all|fast|browser` 可覆盖默认流程；Playwright MCP 不可用时，自动模式可降级为全量 Browser 验证。
- 不读取代码，不调用内部 API，不访问数据库，不执行页面任意 JavaScript。
- AI 对每条用例给出 `passed`、`failed`、`blocked` 或 `inconclusive`，对外展示中文。
- 是否需要人工处理由 Browser 结果决定；人工可以在结果出来后追加测试范围。
- 测试用例只读；当前状态可更新；执行记录只追加、不覆盖。
- 支持 `mode=normal|development`；开发模式增加过程日志，不改变测试行为和结论。快速验证不保存截图，Browser 验证沿用既有截图策略。

两阶段设计与当前实现边界见[两阶段测试设计](docs/two-stage-design.md)。该开发版本仍需在 Cursor 中安装 Playwright MCP 后进行真实回归。

## 当前进度

Cursor 原生 Browser 的核心可行性已经验证，`execute-test-cases` 也已能作为独立入口调用 Browser。当前开发版本为 `0.3.0-dev.1`：增加 Playwright MCP 快速验证、自动 Browser 选择、两阶段状态与 Schema 5；Windows/Linux x64 的编译型结果写入器继续防止 JSONL 被 Agent 拼坏。两阶段能力尚待在 Cursor 中实际回归。

更多设计与背景：

- [当前 PoC 状态](docs/poc-status.md)
- [MVP 方案](docs/方案.md)
- [架构设计](docs/architecture.md)
- [构建、分发与安装设计](docs/distribution.md)
- [开源方案调研](docs/research/open-source-landscape.md)
- [版本变更记录](CHANGELOG.md)

## 当前交付方向

首版优先交付 Cursor Skill、Windows/Linux x64 静态结果写入器、输入模板和结果格式，不自研浏览器驱动。GitHub Actions 负责测试、双平台交叉编译和 tag Release；最终使用者不安装语言运行时。跑通完整闭环后，再评估 macOS/ARM64、Cursor/Agent Plugin、Excel/飞书适配、其他 Agent 平台、代码与数据库只读上下文以及独立服务化。
