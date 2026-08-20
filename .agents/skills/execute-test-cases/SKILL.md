---
name: execute-test-cases
description: 使用 Cursor Agent Window 原生 Browser 执行已有的 Web 正流程测试用例，保存可审计的当前状态、追加式执行历史、截图证据和人工处理清单，并支持不中断历史的恢复与复测。当用户提供测试环境 URL、只读测试用例和专用测试账号，并要求通过 Cursor 进行浏览器自动化测试时使用。不要用于修改应用代码或测试生产环境。
---

# 执行浏览器测试用例

使用 Cursor Agent Window 原生 Browser，优先使用已验证的 `/use-browser` 入口。不要安装或静默替换为其他浏览器工具。

## 限制范围

- 只通过可见浏览器 UI 测试。
- 将源测试用例视为只读输入。
- 不读取或修改应用代码，不调用内部 API，不访问数据库，不修复发现的问题。
- 只使用测试环境和专用测试账号。
- 顺序执行所有输入用例，并将每条用例视为相互独立。
- 不在结果中写入密码或验证码。

## 检查输入

确认已提供环境 URL、测试用例文件和具名测试账号。每条用例应包含 ID、使用账号、前置条件、有序步骤和逐步预期。

短信 mock 登录标记为“浏览器自动填入”时，输入手机号并获取验证码，然后等待页面自动填入。超时仍未填入时将用例标记为 `blocked`。

## 解析运行参数

从本次 Prompt 和运行配置中读取 `mode`，只接受：

- `normal`：正常模式；
- `development`：开发模式。

新运行按“本次 Prompt 参数 > 运行配置 > `normal`”确定模式。参数形式为 `mode=development` 或 `mode=normal`。值非法时在启动 Browser 前停止，并列出合法值。恢复已有 run 时沿用 `summary.md` 中的原模式，不在中途切换。

读取本 Skill 同目录的 `VERSION` 作为 `skill_version`，使用 `schema_version=1`。开始前回显并在 `summary.md` 中记录最终生效的运行模式、Skill 版本和结果 Schema 版本。

## 创建运行记录

创建 `.ai-auto-test/results/<run-id>/`，包含：

- `case-status.csv`：当前状态投影；
- `case-executions.jsonl`：只追加的 attempt 历史；
- `run-events.jsonl`：只追加的运行过程事件；
- `summary.md`：运行元数据、状态统计和人工处理清单；
- 截图证据文件。

JSONL 和 Markdown 使用无 BOM 的 UTF-8，CSV 使用带 BOM 的 UTF-8。将每条用例初始化为 `pending`（待测试）。额外文件权限只能授予当前结果目录。

每条事件至少包含 `time`、`skill_version`、`schema_version`、`run_id`、`mode` 和 `event`，可按需增加 `case_id`、`attempt`、`step_index`、`execution_id`、`status`、`message`。每行必须是独立有效的 JSON，写入后禁止修改。事件内容必须脱敏，不记录凭据、Cookie、Token、完整请求头或 Agent 内部推理。

两种模式都记录 `run_started`、Preflight 结论、`case_started`、`execution_appended`、`status_updated`、`self_check_finished`、`run_finished` 以及 warning/error。

`development` 额外记录 `preflight_started`、`step_started`、`browser_action_started`、`browser_action_finished`、`step_observed`、`ui_adaptation`、`screenshot_saved`、关键文件的 `write_started`/`write_completed`、`self_check_started`/`self_check_detail` 和恢复依据。开发模式只能增加可观测信息，不得改变操作、结论或人工处理规则。

每条实际开始执行的来源用例步骤必须原样保留编号，并且恰好对应一组 `step_started` 和 `step_observed`；不得把多条来源步骤压缩成一条汇总步骤。`passed` 必须覆盖全部来源步骤；因前置条件提前 `blocked` 时可以没有业务步骤事件，但必须记录阻塞原因。一个步骤可以包含多个 Browser action。每次真实调用 Browser 前后分别追加：

- `browser_action_started`：包含 `case_id`、`attempt`、`step_index`、递增的 `action_seq`、`action` 和脱敏后的 `target`；
- `browser_action_finished`：包含相同定位字段和 `result=success|failed`，失败时记录脱敏错误；
- 输入操作只记录字段用途，不记录输入值；账号密码、验证码、Cookie、Token 和完整请求内容不得进入事件。

这些事件用于解释执行过程，不是独立于 Agent 的可信执行证明；最终仍需结合 Cursor Browser 工具轨迹、页面观察和截图判断。

先追加 `run_started`，再启动 Browser。

## 执行 Browser Preflight

打开目标 URL，确认能够读取可见页面内容。Browser 不可用时记录原始错误，追加 `preflight_failed` 和 `run_finished`，在 summary 中将整次运行标记为 `blocked`；不要把业务用例标记为失败。成功时追加 `preflight_passed`，再执行业务用例。

## 逐条执行用例

对每条 `pending` 或 `retest_pending` 用例：

1. 重新读取执行历史，令 `attempt` 等于该用例已有最大 attempt 加一；新用例从 1 开始。
2. 生成包含 run ID、用例 ID 和 attempt 的 execution ID；执行前确认执行历史中不存在相同 execution ID 或相同用例与 attempt。
3. 选择指定账号；仅在用例允许时复用登录会话。
4. 只通过浏览器建立前置条件；无法建立时返回 `blocked`。
5. 按来源步骤顺序执行，不改变、合并或跳过业务意图。开发模式在每个来源步骤前后写入对应事件，并记录其间的 Browser action。
6. 只适应关闭无害弹窗、滚动、等待或切换标签页等 UI 机械细节，并记录偏差。
7. 根据渲染后的可见内容验证每项预期；点击成功不能单独作为通过证据。
8. 截取最终页面或异常页面，并将截图保存到结果目录。
9. 追加执行记录前再次读取执行历史，确认 execution ID 和“用例 ID + attempt”仍然唯一。若冲突，不得追加，记录 `integrity_error`，停止新的业务执行并进入失败自检。
10. 先追加完整执行记录，再追加 `execution_appended` 事件。
11. 更新 `last_execution_id`、状态、`manual_required` 和时间，再追加 `status_updated` 事件。
12. 继续下一条用例。

仅当步骤明确允许自由取值时生成简单唯一值。不要改变固定金额、角色、商品、状态或预期结果。

## 给出唯一结论

- `passed`（已通过）：直接观察到所有预期。
- `failed`（不通过）：可见行为与预期矛盾。
- `blocked`（测试受阻）：登录、权限、环境、数据、Browser 或其他前置条件阻止执行。
- `inconclusive`（无法判断）：已经执行，但可见证据不足以判断。

仅对 `passed` 设置 `manual_required=false`；其他结论全部设置为 `true`。不要把 `blocked` 或 `inconclusive` 改判为通过。

按以下边界判定，不得用 `blocked` 掩盖页面缺陷：

- 已进入正确页面，但用例预期存在的按钮、字段、提示或状态不存在：`failed`；
- 因前置数据、账号、权限、环境或 Browser 能力而无法到达需要断言的页面或状态：`blocked`；
- 已完成操作，但仅凭可见 UI 无法确认预期是否成立：`inconclusive`；
- 当前 Browser-only 边界无法执行内部 API、数据库或伪造内部标识：`blocked`，并明确记录 `reasonCode=capability_boundary`。

## 恢复与复测

沿用原 run ID。先读取状态和历史，只执行 `pending`/`retest_pending`，禁止删除或修改已有 JSONL 行。分配下一个 attempt 和唯一 execution ID。中断后状态与历史不一致时，以追加式执行历史重建状态，并记录修复原因。若历史已经存在重复 execution ID 或重复的“用例 ID + attempt”，该 run 无法在只追加约束下安全修复：将其标为无效，保留原始文件并使用新 run ID 重跑。

## 汇总并自检

在 `summary.md` 中记录 `skill_version`、`schema_version`、`run_id`、`mode`、`run_state`、`self_check`、`run_valid`、Cursor 版本、用例来源、当前状态统计、每条用例的最新结果、证据引用、可用时的截图哈希，以及所有异常用例组成的人工处理清单。`run_state` 表示执行进度，使用 `completed|blocked|interrupted`；`run_valid` 独立表示结果记录是否通过完整性自检。业务执行可以 `completed` 但因记录冲突而 `run_valid=false`。

完成前重新读取四个结果文件和证据目录，逐项检查：

- 每条输入用例在状态表中恰好出现一次；
- 每个完成状态都指向该用例最新的 execution ID；
- execution ID 唯一，“用例 ID + attempt”唯一，attempt 从 1 连续递增；
- 状态表与最新执行结论一致；
- `failed`、`blocked`、`inconclusive` 均设置 `manual_required=true`；
- summary 状态计数与状态表一致；
- `run-events.jsonl` 每行是有效 JSON，且版本、run ID 和模式一致；
- 每个完成用例都能找到 `execution_appended` 和其后的 `status_updated`；
- 开发模式下，已执行的来源步骤与执行记录、`step_started`、`step_observed` 一一对应；每个 `browser_action_started` 都有相同定位字段的 `browser_action_finished`；
- 引用的截图位于当前结果目录并实际存在；
- 结果文件不包含密码或验证码。

自检全部通过时写入 `self_check=passed`、`run_valid=true`。任何 JSONL 无法解析、execution ID 重复、“用例 ID + attempt”重复、状态指针错误、状态与最新执行不一致或开发步骤事件缺失都属于硬失败，必须写入 `self_check=failed`、`run_valid=false`，追加带 `result=failed` 的 `self_check_finished`，并在 `run_finished` 中携带 `run_valid=false`。不得把硬失败降级为 warning 或宣告闭环成功。

存在不一致时，只修复能从唯一追加历史确定重建的状态或汇总。无法安全修复时保留原始记录，在 summary 增加运行级人工处理项；不要删除、修改或再次追加冲突的执行记录。该检查由 Agent 完成，不要声称它是独立程序或确定性校验。报告所有 warning 和未能持久化的证据。
