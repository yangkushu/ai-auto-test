# Cursor 两用例闭环验证提示词

## 使用方式

1. 将 `examples/closed-loop/run-config.example.md` 复制为同目录的 `run-config.local.md`，填写测试账号；不要提交该文件。
2. 在 Cursor 3.15.x Agent Window 输入框中选择 `/use-browser`。
3. 将下面的提示词提交给 Cursor Agent。

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

1. 读取运行配置和测试用例。
2. 按 Prompt 参数覆盖运行配置，确认本次使用 `development`，读取 Skill 同目录的 `VERSION`，并回显运行模式、Skill 版本和 `schema_version=1`。
3. 创建时间戳格式的 `<run-id>`，并在 `.ai-auto-test/results/<run-id>/` 创建：
   - `case-status.csv`：参考 `examples/closed-loop/case-status.example.csv`；
   - `case-executions.jsonl`：每次尝试追加一行 JSON，参考 `examples/closed-loop/case-execution.example.jsonl`；
   - `run-events.jsonl`：开发过程事件，参考 `examples/closed-loop/run-events.example.jsonl`；
   - `summary.md`：批次信息、统计和人工处理清单。
   创建后先向 `run-events.jsonl` 写入 `run_started`。
4. 做 Browser Preflight：使用 Cursor 原生 Browser 打开目标 URL，并确认能读取页面标题或主要可见内容。
5. 如果 Preflight 失败：
   - 创建 `.ai-auto-test/results/<run-id>/summary.md`；
   - 向 `run-events.jsonl` 追加 `preflight_failed` 和 `run_finished`；
   - 将运行结果标记为 `blocked`，原样记录 Browser 错误；
   - 在 summary 中记录 skill_version、schema_version、run_id、mode 和失败原因；
   - 不执行业务用例，也不得将业务用例标记为 `failed`。
6. 如果 Preflight 成功，向 `run-events.jsonl` 追加 `preflight_passed`，然后继续。
7. 将两条用例初始化为 `pending（待测试）`。
8. 每次只执行一条用例。每一步都记录实际操作、预期和页面上的可见观察，不能因为点击成功就判定通过。按开发模式记录步骤、UI 适应、截图和关键写入事件。
9. 每条用例结束时只允许一个结论：
   - `passed（已通过）`
   - `failed（不通过）`
   - `blocked（测试受阻）`
   - `inconclusive（无法判断）`
10. 每条用例结束后按此顺序落盘：
   - 截取最终页面或异常页面；
   - 向 `case-executions.jsonl` 追加完整执行记录；
   - 向 `run-events.jsonl` 追加 `execution_appended`；
   - 更新 `case-status.csv` 中该用例的当前状态、最后执行记录 ID 和时间；
   - 向 `run-events.jsonl` 追加 `status_updated`；
   - 再开始下一条用例。
11. `manual_required` 规则：`passed` 为 `false`；其余三个结果为 `true`。
12. 如果截图只能存在于 Cursor 会话，无法保存为本地文件，必须在执行记录和汇总中明确写出限制，不得伪造文件路径。
13. 全部完成后更新 `summary.md`，至少包含：
    - skill_version、schema_version、run_id、mode、Cursor 版本、目标环境（不含凭据）、起止时间；
    - passed/failed/blocked/inconclusive 数量；
    - 每条用例的最终结论和简要可见证据；
    - 所有 `manual_required=true` 的人工处理清单；
    - 截图证据是否成功持久化。
14. 最后重新读取 `case-status.csv`、`case-executions.jsonl`、`run-events.jsonl`、`summary.md` 和截图目录，检查用例数量、最新 execution ID、状态、attempt、事件顺序、统计、人工处理标记和证据引用是否一致，并将自检结论写入 `summary.md`。该步骤由 Agent 完成，不调用外部校验脚本。

输入齐全后直接执行。不要搜索 MCP、扫描待测项目、设计新框架或改写测试用例。
