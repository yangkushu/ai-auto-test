---
name: execute-test-cases
description: 使用 Cursor Agent Window 原生 Browser 执行已有的 Web 正流程测试用例，保存可审计的当前状态、追加式执行历史、截图证据和人工处理清单，并支持不中断历史的恢复与复测。当用户提供测试环境 URL、只读测试用例和专用测试账号，并要求通过 Cursor 进行浏览器自动化测试时使用。不要用于修改应用代码或测试生产环境。
---

# 执行浏览器测试用例

将本 Skill 作为唯一入口，从 `/execute-test-cases` 或明确要求使用 `execute-test-cases` 的 Prompt 启动。由 Skill 内部调用 Cursor Agent Window 原生 Browser；不要要求用户手动选择 `/use-browser`，也不要安装或静默替换为其他浏览器工具。

## 限制范围

- 只通过可见浏览器 UI 测试。
- 将源测试用例视为只读输入。
- 不读取或修改应用代码，不调用内部 API，不访问数据库，不修复发现的问题。
- 只使用测试环境和专用测试账号。
- 顺序执行所有输入用例，并将每条用例视为相互独立。
- 不在结果中写入密码或验证码。

## 表达规则

- Agent 新生成的 `summary.md`、`message`、`reason`、`observed`、诊断和人工处理说明统一使用中文；技术术语首次可使用“中文（English）”。
- `case-status.csv` 是人类可读状态表：表头、状态、布尔值和清理状态必须使用中文；ID、时间和执行记录 ID 保持原值。
- `case-executions.jsonl`、`run-events.jsonl` 是机器审计记录：字段名、状态值和事件名使用英文；其中 `message`、`reason`、`observed` 等人类可读字段使用中文。
- 代码、命令、URL、页面文案、测试用例原文、截图和浏览器原始错误保持原样；不要翻译或改写原始证据。

## 检查输入

确认已提供环境 URL、测试用例文件和具名测试账号。每条用例应包含 ID、使用账号、前置条件、有序步骤和逐步预期。

在创建 run 或启动 Browser 前完成严格校验：

- URL 必须是完整的 `http://` 或 `https://` 地址；配置文件和用例文件必须存在且可读；
- 拒绝空值以及 `<...>`、`TODO`、`待填写`、`填写路径` 等占位内容；
- 每条选中用例引用的账号必须在配置中定义；账号密码登录必须有用户名和密码，除非配置明确允许复用已有会话；
- 配置了 Preflight 成功标志时必须读取为非占位文本。

输入无效时列出缺失或无效字段并停止：不创建 run、不启动 Browser，也不从用例内容猜测账号配置。

短信 mock 登录标记为“浏览器自动填入”时，输入手机号并获取验证码，然后等待页面自动填入。超时仍未填入时将用例标记为 `blocked`。

## 解析运行参数

从本次 Prompt 和运行配置中读取 `mode`，只接受：

- `normal`：正常模式；
- `development`：开发模式。

新运行按“本次 Prompt 参数 > 运行配置 > `normal`”确定模式。参数形式为 `mode=development` 或 `mode=normal`。值非法时在启动 Browser 前停止，并列出合法值。恢复已有 run 时沿用 `summary.md` 中的原模式，不在中途切换。

读取本 Skill 同目录的 `VERSION` 作为 `skill_version`，使用 `schema_version=3`。开始前回显并在 `summary.md` 中记录最终生效的运行模式、Skill 版本和结果 Schema 版本。

仅支持 Windows x64 和 Linux x64。根据 Cursor 所在平台选择 `bin/ai-auto-test-store-windows-amd64.exe` 或 `bin/ai-auto-test-store-linux-amd64`，运行其 `version` 命令并要求输出版本与 `skill_version` 一致；其他 OS/Architecture、文件缺失、无法执行或版本不一致时在启动 Browser 前停止。使用者不需要安装 Go、Python、Node.js 或 PowerShell。

## 创建运行记录

创建 `.ai-auto-test/results/<run-id>/`，包含：

- `case-status.csv`：中文表头与中文展示值的当前状态投影；
- `case-executions.jsonl`：只追加的 attempt 历史；
- `run-events.jsonl`：只追加的运行过程事件；
- `summary.md`：运行元数据、状态统计和人工处理清单；
- 截图证据文件。

JSONL 和 Markdown 使用无 BOM 的 UTF-8，CSV 使用带 BOM 的 UTF-8。`case-status.csv` 表头必须严格为 `运行ID,用例ID,浏览器状态,是否需要人工处理,是否需要清理,清理状态,最近执行ID,更新时间`。状态映射为 `pending=待测试`、`passed=已通过`、`failed=不通过`、`blocked=测试受阻`、`inconclusive=无法判断`、`retest_pending=待复测`；布尔值映射为 `true=是`、`false=否`；清理状态映射为 `not_applicable=不适用`、`completed=已完成`、`pending_manual=待人工清理`。将每条用例初始化为“待测试”。额外文件权限只能授予当前结果目录。

所有 `run-events.jsonl` 和 `case-executions.jsonl` 写入必须调用随 Skill 交付的 CLI，禁止使用文件编辑工具、shell 重定向或自行拼接文本直接写这两个文件：

```text
ai-auto-test-store init-jsonl --file <path> --kind events|executions
ai-auto-test-store append-jsonl --file <path> --kind events|executions --json <单个 JSON 对象>
ai-auto-test-store append-jsonl --file <path> --kind events|executions --json-file <单对象临时文件>
ai-auto-test-store validate-jsonl --file <path> --kind events|executions
```

创建结果目录后，先分别调用 `init-jsonl --kind events|executions` 创建两个空 JSONL，再追加 `run_started`。一次 append 命令只能提交一个 JSON 对象。较大的 execution 可以先写入当前结果目录中的单对象临时 JSON，再通过 `--json-file` 提交；成功后删除该明确的临时文件。CLI 会先校验已有文件与候选内容，再以 UTF-8 无 BOM 的单行 JSON 追加并立即复验；`executions` 还会在写入前拒绝重复 execution ID 和重复的“用例 ID + attempt”。任何 `ok=false` 或非零退出码都必须停止后续 JSONL 写入并进入失败自检，不得退回直接写文件。

每条事件至少包含 `time`、`skill_version`、`schema_version`、`run_id`、`mode` 和 `event`，可按需增加 `case_id`、`attempt`、`step_index`、`execution_id`、`status`、`message`。每行必须是独立有效的 JSON，写入后禁止修改。事件内容必须脱敏，不记录凭据、Cookie、Token、完整请求头或 Agent 内部推理。`event` 使用英文机器值；`message` 使用中文。禁止使用已废弃的 `self_check`，只能使用 `self_check_finished`。

两种模式都记录 `run_started`、Preflight 结论、`case_started`、`execution_appended`、`status_updated`、`self_check_finished`、`run_finished` 以及 warning/error。`self_check_finished` 是唯一合法的自检结束事件，必须在 `run_finished` 前追加。

`development` 额外记录 `preflight_started`、`step_started`、`browser_action_started`、`browser_action_finished`、`step_observed`、`ui_adaptation`、`screenshot_saved`、关键业务结果文件的 `write_started`/`write_completed`、`self_check_started`/`self_check_detail` 和恢复依据。不要为向 `run-events.jsonl` 追加事件本身再生成 write 事件，避免递归日志。开发模式对每条实际执行的用例保存最终页面或异常现场截图；正常模式只对 `failed`、`blocked`、`inconclusive` 保存异常现场截图，`passed` 不保存截图。两种模式在 Preflight 失败时都保存一张失败页面截图。模式只能改变可观测信息，不得改变操作、结论或人工处理规则。

每条实际开始执行的来源用例步骤必须原样保留编号，并且恰好对应一组 `step_started` 和 `step_observed`；不得把多条来源步骤压缩成一条汇总步骤。`passed` 必须覆盖全部来源步骤；因前置条件提前 `blocked` 时可以没有业务步骤事件，但必须记录阻塞原因。一个步骤可以包含多个 Browser action。每次真实调用 Browser 前后分别追加：

- `browser_action_started`：包含 `case_id`、`attempt`、`step_index`、递增的 `action_seq`、`action` 和脱敏后的 `target`；
- `browser_action_finished`：包含相同定位字段和 `result=success|failed`，失败时记录脱敏错误；
- 输入操作只记录字段用途，不记录输入值；账号密码、验证码、Cookie、Token 和完整请求内容不得进入事件。

这些事件用于解释执行过程，不是独立于 Agent 的可信执行证明；最终仍需结合 Cursor Browser 工具轨迹、页面观察和截图判断。

先追加 `run_started`，再启动 Browser。

## 执行分阶段前置检查（Preflight）

业务用例前依次完成以下阶段；任一阶段失败都不得进入下一阶段：

1. `input_validation`：运行配置、用例、账号、平台二进制和版本检查；此阶段在创建 run 前完成；
2. `browser_capability`：Cursor 原生 Browser 能启动并读取页面；
3. `target_navigation`：目标 URL 可导航，记录请求 URL、最终 URL、页面标题、可见错误内容和浏览器报告的重定向；
4. `application_identity`：页面不是浏览器错误页、连接失败页、4xx/5xx、`Service Unavailable`、`Bad Gateway`、`Gateway Timeout`、`Internal Server Error` 或同类服务错误页；同时匹配配置成功标志，或识别登录表单、应用导航、主标题等明确应用界面。

`input_validation` 在创建 run 前完成：输入无效时不创建结果目录，因此不产生运行事件；输入通过时在 `run_started` 中说明该阶段已通过。创建 run 后，开发模式在其余每个阶段开始时追加 `preflight_stage_started`，通过时追加 `preflight_stage_passed`。所有阶段通过后追加 `preflight_passed`。

阶段失败时追加 `preflight_failed`，字段必须包含：

- `preflight_stage`：失败阶段；
- `reason_code`：如 `browser_unavailable`、`target_navigation_failed`、`target_error_page`、`application_identity_missing`；
- `observations`：可直接观察到的事实，包括请求 URL、最终 URL、标题、可见错误文案或缺失的成功标志；
- `hypotheses`：可选推断，每项包含 `cause`、`confidence=confirmed|likely|possible`、`evidence`；
- `recommended_actions`：中文、可执行的人工处理建议。

不要把观察到 `503` 直接写成“等待环境恢复”。只可确认“目标 URL 未获得应用页面”；域名迁移、网关故障、服务未部署、路径或协议错误都只能列为带置信度的可能原因。只可跟随 Browser 明确报告的重定向；不得猜测、扫描或尝试其他域名。

Preflight 失败时，不执行业务用例。所有用例保持 `pending`，不得创建 `case_started`、case execution 或 `status_updated`。完成 JSONL 自检后写入 `run_state=blocked`；记录完整时使用 `self_check=passed`、`run_valid=true`，并在 summary 中按“事实、推断、建议”写入运行级人工处理项。仅当 URL、运行配置均未改变且属于暂时性可用性问题时，才可沿用该 run ID 从 pending 继续；目标 URL 或其他运行配置改变时必须新建 run。

## 逐条执行用例

对每条 `pending` 或 `retest_pending` 用例：

1. 重新读取执行历史，令 `attempt` 等于该用例已有最大 attempt 加一；新用例从 1 开始。
2. 生成包含 run ID、用例 ID 和 attempt 的 execution ID；执行前确认执行历史中不存在相同 execution ID 或相同用例与 attempt。
3. 选择指定账号；仅在用例允许时复用登录会话。
4. 只通过浏览器建立前置条件；无法建立时返回 `blocked`。
5. 识别是否存在写操作：创建、新增、保存、提交、发布、修改、删除、支付或任何可改变业务数据的操作均属于写操作。
6. 对写操作先通过可见 UI 检查测试数据冲突。仅当来源步骤明确允许自由取值或给出唯一值占位符时，生成 `AUTO-<run-id>-<case-id>` 形式的唯一值；固定金额、名称、角色、商品、状态和预期不得擅自改写。固定值已存在、无法确认目标对象唯一，或无法通过 UI 建立可靠创建前状态时，不执行写操作，标记 `blocked`，使用 `reasonCode=test_data_conflict`。
7. 按来源步骤顺序执行，不改变、合并或跳过业务意图。开发模式在每个来源步骤前后写入对应事件，并记录其间的 Browser action。
8. 只适应关闭无害弹窗、滚动、等待或切换标签页等 UI 机械细节，并记录偏差。
9. 根据渲染后的可见内容验证每项预期；点击成功不能单独作为通过证据。
10. 按模式保存截图：`development` 对每条实际执行用例保存最终页面或异常现场；`normal` 只在 `failed`、`blocked`、`inconclusive` 时立即保存异常现场。两种模式的 Preflight 失败都保存失败页面。每次 attempt 使用独立文件名，不得覆盖历史证据。
11. 对每条 execution 写入 `sideEffects`：`hasWriteOperation`、`resources`（`resourceType`、`visibleIdentifier`、`created`）、`cleanupRequired`、`cleanupStatus=not_applicable|completed|pending_manual`、`cleanupReason`。无写操作时使用 `hasWriteOperation=false`、`cleanupRequired=false`、`cleanupStatus=not_applicable`。写操作没有来源用例中的已验证清理步骤时，使用 `cleanupRequired=true`、`cleanupStatus=pending_manual`。
12. 除非来源用例明确包含清理步骤，否则不得自动删除、撤销或修改已创建数据。`manual_required` 只表示人工业务复核需求；测试数据待清理时保持原业务结论，并单独进入“测试数据清理清单”。
13. 追加执行记录前再次读取执行历史，确认 execution ID 和“用例 ID + attempt”仍然唯一。若冲突，不得追加，记录 `integrity_error`，停止新的业务执行并进入失败自检。
14. 通过 `ai-auto-test-store append-jsonl --kind executions` 追加完整执行记录，再通过 `append-jsonl --kind events` 追加 `execution_appended` 事件。
15. 使用中文表头和中文展示值更新状态表中的最近执行记录、浏览器状态、人工处理、清理状态和时间，再追加 `status_updated` 事件。
16. 继续下一条用例。

## 给出唯一结论

- `passed`（已通过）：直接观察到所有预期。
- `failed`（不通过）：可见行为与预期矛盾。
- `blocked`（测试受阻）：登录、权限、环境、数据、Browser 或其他前置条件阻止执行。
- `inconclusive`（无法判断）：已经执行，但可见证据不足以判断。

仅对 `passed` 设置 `manual_required=false`；其他结论全部设置为 `true`。不要把 `blocked` 或 `inconclusive` 改判为通过。

按以下边界判定，不得用 `blocked` 掩盖页面缺陷：

- 已进入正确页面，但用例预期存在的按钮、字段、提示或状态不存在：`failed`；
- 因前置数据、账号、权限、环境或 Browser 能力而无法到达需要断言的页面或状态：`blocked`；
- 固定写入值已存在、目标对象无法唯一识别或创建前状态无法可靠建立：`blocked`，并使用 `reasonCode=test_data_conflict`；
- 已完成操作，但仅凭可见 UI 无法确认预期是否成立：`inconclusive`；
- 当前 Browser-only 边界无法执行内部 API、数据库或伪造内部标识：`blocked`，并明确记录 `reasonCode=capability_boundary`。

## 恢复与复测

沿用原 run ID 前，确认目标 URL、账号配置、用例选择、运行模式和 Schema 版本均未改变。任一运行配置改变，尤其是目标 URL 或 Schema 版本改变时，必须新建 run，禁止把不同目标环境或结果契约混入旧运行历史。配置未变时，读取状态和历史，只执行 `pending`/`retest_pending`，禁止删除或修改已有 JSONL 行。分配下一个 attempt 和唯一 execution ID。中断后状态与历史不一致时，以追加式执行历史重建状态，并记录修复原因。若历史已经存在重复 execution ID 或重复的“用例 ID + attempt”，该 run 无法在只追加约束下安全修复：将其标为无效，保留原始文件并使用新 run ID 重跑。Schema 1、2 的历史记录可由 CLI 校验，但不得在 Schema 3 下继续追加。

## 汇总并自检

在 `summary.md` 中用中文记录 `skill_version`、`result_store_version`、`schema_version`、`run_id`、`mode`、`run_state`、`self_check`、`run_valid`、Cursor 版本、用例来源、当前状态统计、每条用例的最新结果、证据引用、可用时的截图哈希，以及所有异常用例组成的人工处理清单。Preflight 失败时，增加“前置检查诊断”，并分为“事实、推断、建议”。额外生成“测试数据清理清单”，列出所有 `cleanupStatus=pending_manual` 的最新 execution；它独立于人工业务处理清单。`run_state` 表示执行进度，使用 `completed|blocked|interrupted`；`run_valid` 独立表示结果记录是否通过完整性自检。业务执行可以 `completed` 但因记录冲突而 `run_valid=false`。

完成前重新读取四个结果文件和证据目录，逐项检查：

- 每条输入用例在状态表中恰好出现一次；
- 每个完成状态都指向该用例最新的 execution ID；
- execution ID 唯一，“用例 ID + attempt”唯一，attempt 从 1 连续递增；
- 状态表与最新执行结论一致；
- `failed`、`blocked`、`inconclusive` 均设置 `manual_required=true`；
- 每条完成 execution 均包含 `sideEffects`；状态表的中文状态、人工处理与清理状态能按既定映射对应最新 execution；
- summary 状态计数与状态表一致；
- `run-events.jsonl` 每行是有效 JSON，且版本、run ID 和模式一致；不存在已废弃的 `self_check`；
- 每个完成用例都能找到 `execution_appended` 和其后的 `status_updated`；
- 开发模式下，已执行的来源步骤与执行记录、`step_started`、`step_observed` 一一对应；每个 `browser_action_started` 都有相同定位字段的 `browser_action_finished`；
- `development` 下每条完成 execution 都有截图；`normal` 下只有非 `passed` execution 与 Preflight 失败必须有截图，所有引用截图都位于当前结果目录并实际存在；
- 结果文件不包含密码或验证码。

自检时必须先使用 `ai-auto-test-store validate-jsonl` 分别校验事件和执行记录，再做跨文件检查。全部通过时先追加 `self_check_finished`（`status=passed`、`run_valid=true`），再写入 `self_check=passed`、`run_valid=true` 和 `run_finished`。任何 CLI 校验失败、JSONL 无法解析、已废弃事件名、execution ID 重复、“用例 ID + attempt”重复、状态指针错误、状态与最新执行不一致或开发步骤事件缺失都属于硬失败，必须写入 `self_check=failed`、`run_valid=false`。如果事件文件仍可安全追加，则追加带 `status=failed` 的 `self_check_finished` 和 `run_finished`；如果 CLI 已拒绝该文件，禁止绕过 CLI，改在 summary 和最终回复中记录硬失败。不得把硬失败降级为 warning 或宣告闭环成功。

存在不一致时，只修复能从唯一追加历史确定重建的状态或汇总。无法安全修复时保留原始记录，在 summary 增加运行级人工处理项；不要删除、修改或再次追加冲突的执行记录。只把 CLI 覆盖的 JSONL 语法、编码和唯一性描述为确定性校验；其余仍是 Agent 自检。报告所有 warning 和未能持久化的证据。
