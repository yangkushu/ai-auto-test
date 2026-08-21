# Cursor 两用例闭环验证提示词

## 使用方式

1. 将 `examples/closed-loop/run-config.example.md` 复制为同目录的 `run-config.local.md`，填写测试账号；不要提交该文件。
2. 在 Cursor 3.15.x Agent Window 输入框中选择 `/execute-test-cases`；如果菜单未显示，则在 Prompt 中明确要求使用项目的 `execute-test-cases` Skill。
3. 将下面的提示词提交给 Cursor Agent。不要手动选择 `/use-browser`，Browser 应由 Skill 内部调用。

## 提示词

请执行 ai-auto-test 的两用例最小闭环验证。

输入：

- 运行参数：`mode=development`
- 运行配置：`examples/closed-loop/run-config.local.md`
- 只读用例：`examples/closed-loop/test-cases.md`

执行范围：

1. 只操作和观察浏览器，不读取或修改待测项目代码，不调用内部 API，不访问数据库。
2. 不新增、修改或删除业务数据，不修复发现的问题，不生成 Playwright 测试脚本。
3. 测试用例只读，按文件顺序执行。两条用例使用同一账号，可以复用登录会话。
4. 不得把用户名、密码或验证码写入任何结果文件或回复。

执行流程：

1. 读取运行配置和测试用例。拒绝不存在的文件、未替换的 `<...>`/TODO 占位符、无效 URL 和未定义账号；输入无效时在创建 run 或启动 Browser 前停止。
2. 按 Prompt 参数覆盖运行配置，确认本次使用 `development`，读取 Skill 同目录的 `VERSION`，并回显运行模式、Skill 版本和 `schema_version=4`。
3. 创建时间戳格式的 `<run-id>`，并在 `.ai-auto-test/results/<run-id>/` 创建：
   - `case-status.csv`：参考 `examples/closed-loop/case-status.example.csv`；
   - `case-executions.jsonl`：每次尝试追加一行 JSON，参考 `examples/closed-loop/case-execution.example.jsonl`；
   - `run-events.jsonl`：开发过程事件，参考 `examples/closed-loop/run-events.example.jsonl`；
   - `summary.md`：批次信息、统计和人工处理清单。
   - `测试报告.md`：面向测试、研发和产品的中文测试结论。
   通过随 Skill 交付的 `ai-auto-test-store init-jsonl` 创建两个空 JSONL，禁止用编辑工具直接创建；随后通过 CLI 向 `run-events.jsonl` 追加 `run_started`。
4. 做分阶段前置检查（Preflight）：依次检查 Browser 能力、目标导航和应用识别。确认能读取健康应用界面并匹配至少一个配置的成功标志。503、其他 4xx/5xx、Service Unavailable、Bad Gateway、Gateway Timeout、Internal Server Error、连接失败或浏览器错误页都必须判定为失败，并记录事实、推断和建议；不得直接要求等待环境恢复。
5. 如果 Preflight 失败：
   - 创建 `.ai-auto-test/results/<run-id>/summary.md` 和 `测试报告.md`；
   - 通过随 Skill 交付的 `ai-auto-test-store` 追加 `preflight_failed`、自检和 `run_finished`；
   - 将 `run_state` 标记为 `blocked`，原样记录脱敏后的 Browser 或环境错误；
   - 在 summary 中记录 skill_version、schema_version、run_id、mode，以及“事实、推断、建议”的失败诊断；
   - 所有用例保持 `pending`，不创建 case execution、`case_started` 或 `status_updated`。
6. 如果 Preflight 成功，通过 CLI 向 `run-events.jsonl` 追加 `preflight_passed`，然后继续。
7. 使用中文表头和中文展示值创建 `case-status.csv`：浏览器状态使用“待测试、已通过、不通过、测试受阻、无法判断、待复测”，人工处理使用“是/否”，清理状态使用“不适用、已完成、待人工清理”。JSONL 的机器字段与状态值保持英文。
8. 每次只执行一条用例。每条来源步骤必须保留原编号，分别记录实际操作、预期和页面上的可见观察，不得合并成用例级汇总步骤。不能因为点击成功就判定通过。按开发模式记录步骤、脱敏后的 `browser_action_started`/`browser_action_finished`、UI 适应、截图和关键写入事件；输入动作不得记录账号、密码或验证码值。
9. 每条用例结束时只允许一个结论：
   - `passed（已通过）`
   - `failed（不通过）`
   - `blocked（测试受阻）`
   - `inconclusive（无法判断）`
10. 每条用例结束后按此顺序落盘，所有 JSONL 必须通过随 Skill 交付的 `ai-auto-test-store` 写入，禁止直接编辑或 shell 重定向：
   - 因本次为 `development`，截取最终页面或异常页面；正常模式仅对异常用例截图；
   - 重新读取执行历史，确认 execution ID 和“用例 ID + attempt”均不存在；若冲突，禁止追加并进入失败自检；
   - 向 `case-executions.jsonl` 追加完整执行记录；
   - 向 `run-events.jsonl` 追加 `execution_appended`；
   - 更新 `case-status.csv` 中该用例的当前状态、最后执行记录 ID 和时间；
   - 向 `run-events.jsonl` 追加 `status_updated`；
   - 再开始下一条用例。
11. `manual_required` 规则：`passed` 为 `false`；其余三个结果为 `true`。写操作还必须记录 `sideEffects` 和清理状态；没有来源清理步骤时进入独立测试数据清理清单，不自动删除。
12. 如果截图只能存在于 Cursor 会话，无法保存为本地文件，必须在执行记录和汇总中明确写出限制，不得伪造文件路径。
13. 全部完成后更新 `summary.md`，至少包含：
    - skill_version、result_store_version、schema_version、run_id、mode、run_state、self_check、run_valid、Cursor 版本、目标环境（不含凭据）、起止时间；
    - passed/failed/blocked/inconclusive 数量；
    - 每条用例的最终结论和简要可见证据；
    - 所有 `manual_required=true` 的人工处理清单；
    - 所有 `cleanupStatus=pending_manual` 的测试数据清理清单；
    - 截图证据是否成功持久化。
14. 生成 `测试报告.md`，使用中文给出报告结论、范围、中文用例明细、问题与人工处理、测试数据影响、证据与限制。ID、URL、版本、机器原因码、原始页面/用例文案和浏览器原始错误可以保持原样。不得复制完整事件日志或写入凭据。
15. 最后先调用 `ai-auto-test-store validate-jsonl` 校验两个 JSONL，再重新读取 `case-status.csv`、`summary.md`、`测试报告.md` 和截图目录，检查用例数量、execution ID 与“用例 ID + attempt”的唯一性、最新状态指针、来源步骤与开发事件的一一对应、Browser action 配对、统计、人工处理标记、清理状态、报告内容和证据引用是否一致。只能追加 `self_check_finished`，不得使用已废弃的 `self_check`。任何 CLI 校验、唯一性、状态指针、报告或步骤事件问题都必须令 `self_check=failed`、`run_valid=false`，不得只记 warning 或宣告闭环成功。

除代码、机器字段、URL、页面/用例原文和浏览器原始错误外，所有新生成的人类可读内容必须使用中文。

输入齐全后直接执行。不要搜索 MCP、扫描待测项目、设计新框架或改写测试用例。
