# ai-auto-test

`ai-auto-test` 是一个供测试人员使用的 AI 浏览器测试编排项目。它读取已有测试用例，调用 Browser Agent 在真实页面上执行正流程，记录每条用例的结论与证据，并生成需要人工处理的清单。

当前目标不是替代人工测试，也不是建立完整测试平台，而是先用 Cursor Agent 跑通一条可重复、可审计的最小流水线。

## 当前 MVP

```text
只读测试用例
    ↓
Cursor 主 Agent 调度
    ↓
/use-browser 执行页面观察和操作
    ↓
用例状态 + 追加式执行记录 + 证据
    ↓
人工处理清单
```

- 所有输入用例默认都进入 Browser 测试，不做前置分类。
- MVP 只通过可见浏览器 UI 测试，不读取代码，不调用内部 API，不访问数据库。
- 不设置独立的 Playwright 测试阶段；Cursor 原生 Browser 是当前执行入口。
- AI 对每条用例给出 `passed`、`failed`、`blocked` 或 `inconclusive`，对外展示中文。
- 是否需要人工处理由 Browser 结果决定；人工可以在结果出来后追加测试范围。
- 测试用例只读；当前状态可更新；执行记录只追加、不覆盖。

## 当前进度

Cursor 原生 Browser 的核心可行性已经验证：能够打开公网网页、访问本地测试页面、在未登录状态下完成登录，并依据可见页面判断结果。外部文件输入、跨会话恢复 pending 用例、复测追加 attempt、状态/执行记录/汇总落盘和截图持久化也已完成验证。无第三方依赖的结果校验器已经实现并通过自动化测试。尚未验证的是异常结论的真实页面校准、突然崩溃恢复、多账号和批量运行。

详见：

- [安装说明](docs/installation.md)
- [使用说明](docs/usage.md)
- [当前 PoC 状态](docs/poc-status.md)
- [MVP 方案](docs/方案.md)
- [架构设计](docs/architecture.md)
- [开源方案调研](docs/research/open-source-landscape.md)
- [两用例闭环验证提示词](prompts/cursor-closed-loop-validation.md)

## 当前交付方向

首版优先交付 Cursor Skill、输入模板和结果格式，不先开发 Plugin，也不自研浏览器驱动。跑通完整闭环后，再评估 Excel/飞书适配、其他 Agent 平台、代码与数据库只读上下文以及独立服务化。

## 快速开始

```bash
git clone git@github.com:yangkushu/ai-auto-test.git
cd ai-auto-test
```

Browser 测试核心不依赖 Node.js，也不需要安装 npm 包。使用 Cursor 打开仓库后，项目级 `execute-test-cases` Skill 会从 `.agents/skills/` 自动发现。

如果需要运行确定性结果校验器或项目自测，则需要 Node.js 20 或更高版本：

```bash
node scripts/validate-run.mjs .ai-auto-test/results/<run-id> --strict
npm test
```
