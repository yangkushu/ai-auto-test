# 开源方案调研

> 历史说明：本文记录 2026-08-18 的早期三阶段方案调研，用于保留选型依据。当前 MVP 已收敛为 Cursor 原生 Browser 单阶段执行，不再采用 Playwright/Browser/人工三阶段模型；当前决策以 [MVP 方案](../方案.md) 和 [PoC 状态](../poc-status.md) 为准。

## 结论

截至 2026-08-18，存在成熟的浏览器 Agent 与自然语言测试框架，但未发现完整覆盖本项目核心模型——“外部只读用例 + Playwright/浏览器/人工三阶段状态 + 不可覆盖执行流水 + 人工复核队列”的成熟开源产品。因此应复用浏览器执行层，自己实现轻量的编排与审计层，而不从零造浏览器 Agent。

## 候选对比

| 项目 | 适合复用的能力 | 成熟度信号 | 不作为本项目主体的原因 |
|---|---|---|---|
| [Playwright MCP](https://github.com/microsoft/playwright-mcp) | 标准化 MCP 浏览器工具、Cursor 可直接配置、控制台与页面操作 | Microsoft、Apache-2.0、约 36.2k stars、572 commits | 提供浏览器工具，不提供用例批次、三阶段状态或审计模型 |
| [Stagehand](https://github.com/browserbase/stagehand) | 自然语言 `act/observe/extract` 与确定性 locator 混用 | MIT、约 24.0k stars、TS/Python/Go | 是 SDK，编排和记录仍须自行实现；常见配置依赖 Browserbase/LLM |
| [Browser Use](https://github.com/browser-use/browser-use) | 自主浏览 Agent、CDP、会话和本地/云运行选择 | 大型社区、Python/Rust 核心 | 面向通用网页任务，执行成功语义不能直接充当测试结论 |
| [Skyvern](https://github.com/Skyvern-AI/skyvern) | 基于视觉与 Playwright 的 AI 工作流、结构化提取 | 约 21.8k stars、2026-05 仍发版 | AGPL-3.0，作为可分发工具的直接嵌入会带来许可证约束 |
| [Hercules](https://github.com/test-zeus-ai/testzeus-hercules) | Gherkin 输入、planner/executor/assertion、HTML/JUnit/证据 | 约 1.1k stars、仍有 PR；明确的测试定位 | AGPL-3.0 且运行时模型与输出模型较重，缺少本项目的人工阶段状态模型 |
| [OpenQA](https://github.com/openqa-labs/openqa) | BDD/YAML + Agent + Playwright MCP 的直接集成 | 117 commits，但约 21 stars | 最接近 Agent 测试执行器，但社区和长期稳定性仍需观察，且无完整批次/人工复核模型 |

上述 star/commit/release 数据来自 2026-08-18 的 GitHub 页面快照，仅作成熟度参考。

## 技术决策

1. 首选将 `@playwright/mcp` 作为 Agent 浏览器工具的默认适配器；其 Apache-2.0 许可证适合交付型项目。
2. 保留 `BrowserAgentAdapter` 接口，以便后续接入 Stagehand、Browser Use 或企业自建浏览器服务。
3. Playwright 的确定性用例执行与 Agent 浏览器旅程分别建适配器，即使底层最终都使用 Playwright，也必须使用独立上下文与独立证据。
4. 参考 Hercules 和 OpenQA 的 Gherkin/BDD 输入与执行证据设计，不复制其运行时或 AGPL 代码。
