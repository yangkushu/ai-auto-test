# Cursor Browser PoC 当前状态

## 结论

截至 2026-08-20，Cursor Agent 原生 Browser 已完成最小可行性验证。Agent 能够打开网页、执行搜索、访问本地测试页面、填写登录表单，并根据页面可见内容确认登录成功。随后完成了基础闭环以及恢复与复测运行，状态、追加式执行记录、汇总和截图均成功落盘；确定性结果校验器也已实现并通过自动化测试。

这证明“外部自然语言用例 → Browser Agent → 页面操作与观察 → 测试结论 → 文件记录 → 恢复与复测”的核心技术路线可行。它尚未证明突然崩溃恢复、异常结论校准和批量测试流水线已经完成。

## 验证环境

| 项目 | 值 |
|---|---|
| 初始 Cursor 版本 | 3.15.6（User Setup） |
| 重装后验证版本 | 3.15.19（x64） |
| Cursor 布局 | Agent Window |
| 操作系统 | Windows x64 |
| Browser 入口 | `/use-browser` |
| 测试对象 | 公网页面与本地测试环境 |
| 测试权限 | 仅可见浏览器 UI |

文档不记录测试账号、密码、验证码或个人本机路径。

## 已验证

1. Cursor Agent 能通过原生 Browser 打开公网网页并完成一次搜索。
2. Cursor Agent 能访问本地运行的 Web 系统。
3. Browser Agent 能识别登录页面、填写测试账号并提交。
4. Browser Agent 能根据登录后的可见页面判断登录成功。
5. 验证过程不需要读取项目代码、调用内部 API、访问数据库或生成 Playwright 测试脚本。
6. 两条内嵌正流程用例能按顺序执行，并分别获得唯一结论。
7. `case-status.csv`、`case-executions.jsonl` 和 `summary.md` 能成功创建。
8. Browser 截图能从 Cursor 临时目录复制到批次结果目录。
9. Agent 能从外部 Markdown 文件读取运行配置和测试用例。
10. 跨会话恢复时能识别并继续执行 `pending` 用例。
11. 复测能将状态置为 `retest_pending`，追加 `attempt=2`，并保留旧记录。
12. 未登录用例能先退出已有会话，再重新提交登录表单并验证成功。
13. 本地校验器能检查 CSV、JSONL、summary、attempt、证据文件和凭据泄露风险。

## 两用例闭环运行

运行 ID：`20260819-123448`。

| 用例 | 结果 | 关键可见证据 |
|---|---|---|
| TC-01 管理员登录 | `passed` | 打开目标 URL 时已处于后台登录状态，可见管理员已登录信息 |
| TC-02 查看门店管理页 | `passed` | 可见门店管理标题、导航、查询区域和门店列表统计 |

运行汇总：`passed=2`，其他状态为 0，没有自动生成的人工处理项。两张截图均已保存到结果目录。

CSV、JSONL 和 summary 已通过用户提供的 UTF-8 解析输出完成结构审计；JSONL 可解析，跨文件 ID、状态、时间、证据引用和统计一致。

### 本轮边界

- TC-01 复用了已有登录会话，本轮验证的是“识别已登录状态”，不是重新提交账号密码；重新登录能力由此前单独 PoC 支持。
- 用例内容直接包含在 Prompt 中，没有验证从 Markdown、Excel 或飞书产物读取用例。
- 默认 sandbox 无法执行终端命令；时间戳生成和截图复制使用了非受限 shell。结果文件本身由文件写入工具创建。

## 恢复与复测运行

运行 ID：`20260819-130012-recovery`，最终 `run_state=completed`。

| 执行 | 结果 | 说明 |
|---|---|---|
| TC-LOGIN-01 attempt=1 | `passed` | 第一阶段执行并落盘后受控停止 |
| TC-STORES-01 attempt=1 | `passed` | 新会话从 `pending` 恢复执行 |
| TC-LOGIN-01 attempt=2 | `passed` | 先置为 `retest_pending`，退出登录后重新登录 |

审计结果：

- `case-executions.jsonl` 包含 3 行有效 JSON 和 3 个唯一 execution ID；
- TC-LOGIN-01 的 attempt 1、attempt 2 同时保留，当前状态指向 attempt 2；
- TC-STORES-01 从 pending 恢复，当前状态指向 attempt 1；
- 三张截图均存在并记录 SHA-256；两张相同落地页截图内容及哈希相同，旧截图未被覆盖；
- 结果文件没有记录登录凭据；
- 本轮是跨会话的受控中断，不等同于进程在两次写入之间突然崩溃。

## 已发现的平台风险

Cursor 原生 Browser 曾出现工具未注册的问题：

```text
MCP server "cursor-ide-browser" not found
Available servers: cursor-app-control
```

此时普通提示词和 `/use-browser` 都无法启动 Browser。重新安装 Cursor 后能力恢复。因此 MVP 必须在正式执行前增加 Browser Preflight，并把这类问题标记为运行级 `blocked`，不能误判为业务用例失败。

建议的 Preflight：

1. 启动 `/use-browser`；
2. 打开一个明确的目标 URL；
3. 确认 Agent 能读取页面标题或主要可见内容；
4. 成功后才创建并执行测试批次；
5. 失败时记录 Cursor 版本、原始错误和时间，不执行用例。

Windows Agent Window 的 workspace sandbox 可能无法提供文件系统隔离。闭环运行中，截图复制和结果追加需要申请结果目录写权限。执行器必须把授权限制在当前结果目录，禁止扩大到待测项目代码或其他目录。

## 尚未验证

- 进程在追加执行记录和更新状态之间突然崩溃后的自动修复；
- `failed`、`blocked`、`inconclusive` 三种异常结论的实际校准；
- 多账号用例及账号切换；
- 短信 mock 的“浏览器自动填入”；
- Excel 输入与输出；
- 飞书文档下载、回写与同步；
- 大批量用例的稳定性、耗时和成本；
- Cursor 之外的平台迁移。

## 校验器实现状态

项目已提供 `scripts/validate-run.mjs`。当前自动化样例覆盖：正常批次、重复 execution/attempt、异常状态未进入人工处理、状态投影过期、summary 计数不一致、证据越界、凭据泄露以及 CLI 严格模式。

这些测试证明校验规则本身可重复运行；尚需在 Cursor 所在的 Windows 环境对真实结果目录执行一次校验，以确认端到端路径和编码兼容性。

## 下一里程碑

验证异常结论和确定性结果校验器：

```text
构造可控的 failed / blocked / inconclusive 场景
  → 检查 Browser 判定与人工队列
  → 使用本地校验器检查 Cursor 真实输出的 CSV / JSONL / summary
  → 模拟两次落盘之间的中断并重建状态
```

验收标准：

1. 三种异常结论不会被误标为通过；
2. 所有异常结论进入人工处理清单；
3. 已有校验规则在真实 Cursor 输出上通过；
4. 在状态更新前中断时，能根据追加记录重建状态；
5. 输出编码在 Windows PowerShell 和 Excel 中可正确读取。
