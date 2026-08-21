# ai-auto-test

`ai-auto-test` 是一个供测试人员使用的 AI 浏览器测试编排项目。它读取已有测试用例，调用 Browser Agent 在真实页面上执行正流程，记录每条用例的结论与证据，并生成需要人工处理的清单。

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

## 详细说明

- [安装说明](docs/installation.md)：安装验收、更新、二进制与平台支持。
- [使用说明](docs/usage.md)：输入格式、开发模式、结果检查和恢复复测。
- [两用例闭环验证提示词](prompts/cursor-closed-loop-validation.md)：可复制的完整验证 Prompt。

## 当前 MVP

```text
只读测试用例
    ↓
Cursor 主 Agent 调度
    ↓
Skill 内部调用 Cursor 原生 Browser
    ↓
用例状态 + 追加式执行记录 + 过程事件 + 证据
    ↓
人工处理清单
```

- 所有输入用例默认都进入 Browser 测试，不做前置分类。
- MVP 只通过可见浏览器 UI 测试，不读取代码，不调用内部 API，不访问数据库。
- 不设置独立的 Playwright 测试阶段；Cursor 原生 Browser 是当前执行入口。
- AI 对每条用例给出 `passed`、`failed`、`blocked` 或 `inconclusive`，对外展示中文。
- 是否需要人工处理由 Browser 结果决定；人工可以在结果出来后追加测试范围。
- 测试用例只读；当前状态可更新；执行记录只追加、不覆盖。
- 支持 `mode=normal|development`；开发模式增加过程日志，不改变测试行为和结论。

## 当前进度

Cursor 原生 Browser 的核心可行性已经验证，`execute-test-cases` 也已能作为独立入口调用 Browser。当前开发版本为 `0.2.0-dev.5`：增加分阶段前置检查（Preflight）诊断、测试数据副作用记录和 `self_check_finished` 事件约束；Windows/Linux x64 的编译型结果写入器继续防止 JSONL 被 Agent 拼坏。

更多设计与背景：

- [当前 PoC 状态](docs/poc-status.md)
- [MVP 方案](docs/方案.md)
- [架构设计](docs/architecture.md)
- [构建、分发与安装设计](docs/distribution.md)
- [开源方案调研](docs/research/open-source-landscape.md)
- [版本变更记录](CHANGELOG.md)

## 当前交付方向

首版优先交付 Cursor Skill、Windows/Linux x64 静态结果写入器、输入模板和结果格式，不自研浏览器驱动。GitHub Actions 负责测试、双平台交叉编译和 tag Release；最终使用者不安装语言运行时。跑通完整闭环后，再评估 macOS/ARM64、Cursor/Agent Plugin、Excel/飞书适配、其他 Agent 平台、代码与数据库只读上下文以及独立服务化。
