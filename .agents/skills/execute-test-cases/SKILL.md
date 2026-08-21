---
name: execute-test-cases
description: 使用 Playwright MCP 快速验证与 Cursor Agent Window 原生 Browser 验证执行已有 Web 测试用例，保存两阶段状态、追加式执行历史、证据、测试报告和人工处理清单。用户提供测试环境 URL、只读用例和专用测试账号，并要求通过 Cursor 自动测试、快速验证或 Browser 验证时使用。不要用于修改应用代码、调用内部 API/数据库或测试生产环境。
---

# 执行两阶段 Web 测试

将本 Skill 作为唯一入口，从 `/execute-test-cases` 或明确要求使用 `execute-test-cases` 的 Prompt 启动。Skill 在内部调用 Playwright MCP 和 Cursor 原生 Browser；不要要求用户手动选择 `/use-browser`，不要静默替换为其他浏览器工具，也不要安装、修改或启用 MCP。

## 边界

- 将来源用例视为只读输入；不读取或修改待测项目代码，不调用内部 API，不访问数据库，不修复缺陷。
- 只使用测试环境、专用测试账号和真实 UI；不在结果中记录密码、验证码、Cookie、Token 或输入值。
- 只可适应滚动、等待、关闭无害提示等机械 UI 细节。不得用根路径、其他页面、其他账号或其他业务步骤替代来源步骤和预期；无法按原用例验证时使用 `blocked` 或 `inconclusive`，不得判 `passed`。
- 人类可读内容使用中文；代码、机器字段、URL、页面/用例原文和 Browser 原始错误保持原样。

## 输入与参数

开始前确认环境 URL、只读测试用例和具名测试账号均存在且可读。每条用例应有 ID、标题、账号、前置条件、有序步骤和逐步预期。拒绝空值、`<...>`、`TODO`、`待填写` 等占位值；输入无效时停止，不创建 run、不启动浏览器、不猜测账号或配置。

解析本次 Prompt 的参数，Prompt 优先于运行配置：

| 参数 | 合法值 | 默认值 |
| --- | --- | --- |
| `stage` | `auto`、`all`、`fast`、`browser` | `auto` |
| `browser_coverage` | `50%`～`70%`，仅 `stage=auto` | `60%` |
| `mode` | `normal`、`development` | `normal` |

- `auto`：全部用例先快速验证，再执行强制 Browser 用例和自动补足覆盖率的用例。
- `all`：全部用例先快速验证，再对全部输入用例执行 Browser 验证。
- `fast`：只执行快速验证；需要 Browser 验证的用例保持 Browser“待测试”，自动化结论为“待完成”。
- `browser`：快速验证标“跳过（指定仅 Browser 验证）”，全部输入用例直接 Browser 验证。

读取同目录 `VERSION`，使用 `schema_version=5`。Schema 1～4 只能校验，禁止恢复或追加；两阶段运行必须新建 run。验证随包 Windows x64 或 Linux x64 `ai-auto-test-store` 的 `version` 与 Skill 版本一致后才创建 run。

## 结果文件

在 `.ai-auto-test/results/<run-id>/` 创建：

- `case-status.csv`：UTF-8 BOM，严格表头为 `运行ID,用例ID,快速验证状态,Browser验证状态,自动化结论,是否需要人工处理,是否需要清理,清理状态,Browser验证原因,最近执行ID,更新时间`；
- `case-executions.jsonl`、`run-events.jsonl`：UTF-8 无 BOM，只能由 `ai-auto-test-store` 初始化、追加和校验；
- `summary.md`、`测试报告.md`：中文；
- Browser 阶段按模式生成的截图证据。

阶段状态机器值为 `pending|passed|failed|blocked|inconclusive|skipped|retest_pending`，中文展示依次为“待测试、已通过、不通过、测试受阻、无法判断、跳过、待复测”。布尔值展示为“是/否”，清理状态展示为“不适用、已完成、待人工清理”。JSONL 继续使用英文机器字段和值。

每条 execution 必须有 `stage=fast|browser`。attempt 按“用例 ID + stage”独立递增，execution ID 必须包含 run ID、用例 ID、stage 和 attempt。所有 JSONL 写入仅可使用：

```text
ai-auto-test-store init-jsonl --file <path> --kind events|executions
ai-auto-test-store append-jsonl --file <path> --kind events|executions --json-file <single-object-file>
ai-auto-test-store validate-jsonl --file <path> --kind events|executions
```

## 前置检查与降级

先完成 `input_validation`，通过后创建 run、初始化 JSONL、追加 `run_started`。后续每个阶段必须记录事实、可选的带置信度推断和可执行建议；503 或其他错误页只能证明目标未获得应用页面，不能猜测或扫描其他域名。

1. **快速验证需要时**：检查 Playwright MCP 已连接并实际可调用导航、无障碍快照、点击和输入。不可用时：
   - `fast` 或 `all`：停止，提示在 Cursor 安装/启用 Playwright MCP 后重试；
   - `auto`：询问用户“安装后重试”或“跳过第一阶段”。只有用户明确选择跳过时，记录 `requested_stage=auto`、`effective_stage=browser`、`fallback_reason=playwright_mcp_unavailable`，将全部用例改为 Browser 验证；
   - `browser`：不检查 Playwright MCP。
2. 对实际执行的首个阶段完成 `target_navigation` 和 `application_identity`：记录请求 URL、最终 URL、标题、可见错误和成功标志。失败时不进入该阶段业务用例。
3. Browser 阶段开始前，检查 Cursor 原生 Browser 可用，再重新完成目标导航和应用识别。Browser 不可用时，只保留已经完成的快速验证结果；所有已选择但未执行的 Browser 用例保持“待测试”，自动化结论为“待完成”。

Preflight 失败时不创建对应阶段的 case execution。`stage=auto` 的快速验证不可用且用户选择降级时，Browser 覆盖率必须为 100%，不得遗漏输入用例。

## 第一阶段：Playwright MCP 快速验证

只使用结构化无障碍快照和以下操作：导航、刷新、返回、等待、点击、填写、清空、下拉选择、勾选/取消勾选，以及 URL、标题、元素可见性、文本、输入值、按钮启用/禁用和列表数量断言。

- 定位顺序：`data-testid` → 角色和可访问名称 → 标签 → placeholder → 精确文本。不得默认生成长 CSS selector 或 XPath。
- 不使用截图、坐标/视觉工具、视频、Trace、DevTools、Network、Cookie/Storage、接口 Mock、`browser_run_code`、`page.evaluate` 或其他任意 JavaScript。
- 仅在初始定位、断言点或引用失效后读取必要深度的快照；不要在每次点击或填写后读取完整树。
- 多账号协作、写操作、验证码、上传、拖拽、多标签页、视觉断言或无法可靠表达的用例，快速验证标记 `skipped` 并记录原因；写操作不得在快速验证中执行。
- 每条实际快速执行用例写入一个 `stage=fast` execution。快速阶段不保存截图；步骤中的结构化可见观察是其证据。

## 第二阶段选择与 Browser 验证

无条件选择 Browser 验证：快速验证不是 `passed`、含写操作、登录/退出/会话保持/权限、多账号、验证码、上传、拖拽、多标签页、明确视觉断言，或用户明确指定的用例。

在 `auto` 下，从其余快速验证已通过用例中，优先按主流程、模块、账号角色和交互复杂度补选，直到 Browser 选择数量达到 `browser_coverage` 的目标。强制选择已经超过目标时允许超出。每条用例必须在状态表、summary 和报告记录简短中文选择原因。V1 不实现复杂风险评分、模块配额或权重算法。

Browser 阶段复用现有 Cursor 原生 Browser 规则：按原用例步骤通过可见 UI 执行和验证；写操作由 Browser 主执行，并遵守测试数据冲突、唯一数据和清理规则。`normal` 只为 `failed|blocked|inconclusive` 保存截图，`development` 为每条 Browser execution 保存最终或异常截图。快速验证不得因开发模式而截图。

## 状态、结论与审计

每次 execution 追加后才更新状态表和 `status_updated`。当前状态不覆盖历史；复测仅追加同一 `stage` 的下一 attempt。每条 execution 仍必须记录来源步骤、可见观察、最终 URL/标题、状态、原因、`manual_required`、证据与 `sideEffects`。无写操作使用 `cleanupStatus=not_applicable`。

按以下规则合成“自动化结论”：

- 快速验证通过，且 Browser 通过或未选中：自动化通过；未选中时在报告中写“快速验证通过，未进行 Browser 验证”。
- 两阶段都执行且结论冲突：自动化无法判断。
- 必需阶段受阻或无法判断：自动化受阻或自动化无法判断。
- 其他任一已执行阶段不通过：自动化不通过。
- `stage=fast` 下 Browser 必需但未执行，或 Browser 阶段不可用而仍有已选择用例：自动化待完成。

仅“自动化通过”设置 `manual_required=false`；其他结论均为 `true`。状态表、summary 和测试报告必须区分快速验证、Browser 验证和最终自动化结论。

## 汇总、自检与报告

生成 `summary.md` 和 `测试报告.md`，记录请求/生效阶段、覆盖率、MCP 可用性、降级原因、两阶段统计、每条用例状态、Browser 选择原因、人工处理、测试数据清理、证据与限制。报告必须使用中文，且说明快速阶段不保存截图是既定策略。

先用 CLI 校验两个 JSONL，再检查：

- 每个输入用例在状态表恰好一次，阶段状态与最新对应 stage execution 一致；
- execution ID 与“用例 ID + stage + attempt”唯一，attempt 在各自阶段连续；
- 已选 Browser 用例均有选择原因；`auto` 覆盖率或超出原因可解释；
- `stage=auto` 降级为 Browser 时覆盖率为 100%；
- 状态、自动化结论、人工处理、清理状态、summary 和报告一致；
- 开发模式下 Browser 步骤/Browser action 事件配对；快速阶段只要求步骤观察，不要求截图；
- 所有 Browser 截图引用存在；结果没有密码或验证码。

报告生成失败、JSONL 校验失败、重复 execution、状态指针错误、覆盖率遗漏、阶段/结论不一致或证据引用错误均为硬失败：`self_check=failed`、`run_valid=false`。成功时只追加 `self_check_finished`，再写入 `run_finished`；不得使用废弃的 `self_check` 事件。
