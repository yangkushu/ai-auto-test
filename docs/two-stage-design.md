# 两阶段测试设计

> `0.3.0-dev.1` 已按本文实现两阶段运行契约和结果 Schema，但尚未完成 Cursor 中 Playwright MCP 的真实回归。旧 Schema 1～4 运行仍是单阶段历史，只可校验，不可追加。

## 目标

在不读取待测项目代码、不调用内部 API、不访问数据库的前提下，降低 Browser 读图带来的 Agent token 消耗，同时保留对复杂页面和高风险流程的可见 UI 验证。

默认路径为：全部输入用例先进行第一阶段测试（快速验证），再由 Agent 选择约 50%～70% 的用例进行第二阶段测试（Browser 验证）。AI 用于提升人工测试效率，不替代人工验收。

```text
只读测试用例
  ↓
共享前置检查
  ↓
第一阶段：Playwright MCP 快速验证（全量）
  ↓
第二阶段选择：强制用例 + 自动补足覆盖率
  ↓
第二阶段：Cursor 原生 Browser 验证（部分或全量）
  ↓
结果汇总、测试报告与人工处理清单
```

## 术语与运行模式

- **第一阶段测试 / 快速验证**：使用 Playwright MCP 的结构化无障碍快照执行简单、确定性的 UI 验证；不使用截图或视觉定位。
- **第二阶段测试 / Browser 验证 / 浏览器验证**：复用现有 Cursor 原生 Browser，通过可见 UI 和按需截图验证复杂、高风险或需要复核的用例。

| 参数 | 含义 |
| --- | --- |
| `stage=auto` | 默认。全部快速验证，再自动选择 Browser 验证用例。 |
| `stage=all` | 全部快速验证，再对全部输入用例执行 Browser 验证。 |
| `stage=fast` | 仅执行全部快速验证。 |
| `stage=browser` | 跳过快速验证，全部输入用例直接执行 Browser 验证。 |
| `browser_coverage=50%` | 仅在 `stage=auto` 下覆盖率目标为 50%。 |
| `browser_coverage=70%` | 仅在 `stage=auto` 下覆盖率目标为 70%。 |

`stage=auto` 的默认 Browser 覆盖率目标为约 60%，允许落在全部输入用例的 50%～70% 区间。强制进入 Browser 验证的用例超过 70% 时，不为了控制比例而跳过高风险用例；实际覆盖率可以超过 70%。

## 共享前置检查与降级

开始业务用例前，始终检查输入、目标 URL、测试账号、测试环境和应用页面识别。前置检查继续使用“事实、推断、建议”表达错误，不把单个 HTTP 错误直接归因为某一种环境故障。

第一阶段额外检查 Playwright MCP 已在 Cursor 中连接，且导航、无障碍快照、点击和输入等核心工具可调用；第二阶段额外检查 Cursor 原生 Browser 可调用。Skill 只检测 MCP 是否可用，不要求在待测项目中安装 Node.js 或 Playwright，也不静默修改用户的 MCP 配置。

| 请求模式 | Playwright MCP 不可用时的处理 |
| --- | --- |
| `stage=fast` | 停止，提示在 Cursor 中安装/启用 Playwright MCP 后重试。 |
| `stage=all` | 停止，提示在 Cursor 中安装/启用 Playwright MCP 后重试。 |
| `stage=browser` | 不检查 Playwright MCP，直接执行第二阶段。 |
| `stage=auto` | 提示用户选择：安装后重试，或跳过第一阶段。 |

在 `stage=auto` 下选择跳过第一阶段时，记录：

```text
requested_stage=auto
effective_stage=browser
fallback_reason=playwright_mcp_unavailable
```

并将 Browser 覆盖率提升为 100%，避免任何输入用例未被测试。

## 第一阶段：快速验证

第一阶段 V1 使用 Playwright MCP 的结构化无障碍快照，不使用截图、坐标点击、视频、Trace、DevTools、Network、Cookie/Storage 读写、接口 Mock 或任意 JavaScript 执行（包括 `browser_run_code`、`page.evaluate`）。

仅支持以下操作与断言：

- 打开页面、刷新、返回、等待页面稳定；
- 按 `data-testid`、角色和名称、标签、placeholder 或明确文本定位；
- 点击、填写、清空、选择下拉项、勾选或取消勾选；
- 断言 URL、标题、元素可见性、文本、输入值、按钮启用/禁用与列表数量。

定位顺序固定为：`data-testid` → 角色和可访问名称 → 标签 → placeholder → 精确文本。不得默认生成长 CSS selector 或 XPath；无法可靠表达的用例标为“快速验证跳过”，并进入第二阶段。

每次仅在页面初始定位、断言点或既有引用失效后读取必要深度的快照，不在每次填充和点击后读取完整页面树。第一阶段仅处理单账号、简单、非写操作用例；多账号、写操作、视觉与复杂交互直接交由第二阶段主执行。

Playwright MCP 是 Agent 驱动的 MCP 工具，而非批量 Playwright Test Runner。它能减少视觉输入，但结构化快照和工具调用仍会消耗 token。真正的脚本化批量执行器属于后续优化，不在本版本范围内。

## 第二阶段选择与执行

以下用例无条件进入 Browser 验证：

- 快速验证为不通过、测试受阻、无法判断或跳过；
- 含创建、提交、发布、修改、删除、支付、验收或结算等写操作；
- 含登录、退出、会话保持、权限或多账号隔离；
- 含验证码、文件上传、拖拽、多标签页或明确视觉断言；
- 用户在本次 Prompt 中明确指定。

其他“快速验证已通过”的用例构成抽样池。Agent 以业务主流程、不同模块、不同账号角色和交互复杂度为优先依据，补足 `browser_coverage` 目标；每条用例必须记录简短的 Browser 选择原因。V1 不引入复杂风险评分、模块配额或权重算法。

写操作在 V1 中由 Browser 验证主执行，快速验证标为“跳过（写操作由 Browser 主执行）”，避免两个阶段重复提交固定测试数据。

第二阶段继续沿用当前已验证的规则：只通过可见 UI 操作；通过用例在正常模式不保存截图，异常用例和开发模式按既有截图策略保存证据；不得读取代码、访问数据库、调用内部 API 或修复缺陷。

## 最小结果模型

首次加入第一阶段时，不引入人工阶段状态机。状态表最小增加以下人类可读列：

```text
快速验证状态
Browser验证状态
自动化结论
是否需要人工处理
Browser验证原因
```

阶段状态均可展示为：待测试、已通过、不通过、测试受阻、无法判断、跳过、待复测。执行历史继续只追加，但每条记录增加机器字段：

```json
"stage": "fast"
```

或：

```json
"stage": "browser"
```

自动化结论的 V1 合成规则：

- 快速验证通过且 Browser 验证通过或未选中：自动化通过；未选中时必须标记“快速验证通过，未进行 Browser 验证”。
- 任一已执行阶段不通过：自动化不通过。
- 必需阶段受阻或无法判断：自动化受阻或自动化无法判断，并进入人工处理。
- 两阶段结论冲突：自动化无法判断，并进入人工处理。

`测试报告.md` 新增“第一阶段结果统计”和“第二阶段覆盖率与选择原因摘要”；现有中文报告、执行记录、证据、自检和人工处理清单继续保留。

## 实施边界与验收

两阶段运行使用 Schema 5，必须新建 run，不得向 Schema 1～4 的单阶段历史追加。发布前应在 Cursor 中确认官方 Playwright MCP 可连接，并以相同的 10 条测试用例分别运行单阶段 Browser 与两阶段 `stage=auto`，对比：

- 全部输入用例是否均获得第一阶段或第二阶段结论；
- Browser 验证覆盖率、强制进入理由和抽样理由是否可解释；
- 结果表、执行记录、测试报告与人工处理清单是否一致；
- 不读取页面截图的第一阶段是否降低 token 消耗；
- 快速验证不可用时，降级到全量 Browser 验证是否不遗漏用例。

相关外部能力参考：[Playwright MCP](https://github.com/microsoft/playwright-mcp)、[Cursor MCP](https://cursor.com/docs/context/model-context-protocol)、[Cursor Browser](https://cursor.com/docs/agent/tools/browser)。
