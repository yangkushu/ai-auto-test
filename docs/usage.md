# 使用说明

## 1. 准备输入

复制 `examples/closed-loop/run-config.example.md` 为 `run-config.local.md`，填写本地测试账号。用例至少包含 ID、标题、账号、前置条件、有序步骤和逐步预期；用例始终只读。

运行参数：

| 参数 | 说明 |
| --- | --- |
| `stage=auto` | 默认。全量快速验证，再选择约 50%～70% 用例进行 Browser 验证。 |
| `stage=all` | 两阶段均全量执行。 |
| `stage=fast` | 仅快速验证；需要 Browser 验证的用例保持待完成。 |
| `stage=browser` | 仅 Browser 验证，覆盖全部输入用例。 |
| `browser_coverage=50%`～`70%` | `auto` 下的 Browser 目标覆盖率；默认 `60%`。 |
| `mode=normal` | 正常模式。Browser 阶段只保存异常截图。 |
| `mode=development` | 开发模式。增加过程事件，Browser 阶段每条 execution 保存截图；快速验证仍不截图。 |

第一阶段需要在 Cursor 中安装并启用 Playwright MCP；Skill 只检测 MCP 是否可用，不修改 MCP 配置。`stage=auto` 下不可用时，Skill 会要求选择“安装后重试”或“跳过第一阶段”。选择跳过会自动转为全量 Browser 验证。

## 2. 启动测试

在 Cursor Agent Window 中确认 `execute-test-cases` 可见后，提交：

```text
使用 execute-test-cases Skill 执行以下测试：

- 运行参数：stage=auto browser_coverage=60% mode=development
- 运行配置：examples/closed-loop/run-config.local.md
- 测试用例：examples/closed-loop/test-cases.md
```

Browser 由 Skill 内部调用。不要手动选择 `/use-browser`；不要在 Prompt 中要求搜索 MCP、读取代码、安装框架、调用 API 或改写测试用例。

## 3. 两阶段执行规则

快速验证使用 Playwright MCP 的无障碍快照执行简单、非写操作。它只允许导航、简单表单操作和 URL/文本/可见性等确定性断言，不使用截图、坐标点击、视频、Trace、Network、Cookie/Storage 或任意 JavaScript。

写操作、登录/会话/权限、多账号、验证码、上传、拖拽、多标签页、视觉断言、快速阶段异常和无法可靠表达的用例，会进入 Browser 验证。其余快速通过用例由 Agent 补足到目标覆盖率；每条选择均记录中文原因。

来源用例必须严格执行。特别是直链、账号、目标页面和业务预期不得被根路径或相似页面替代；无法按原步骤验证时只能判为“测试受阻”或“无法判断”。

## 4. 结果与检查

结果目录：

```text
.ai-auto-test/results/<run-id>/
├── case-status.csv
├── case-executions.jsonl
├── run-events.jsonl
├── summary.md
├── 测试报告.md
└── Browser 阶段截图（按 mode 与结论生成）
```

优先阅读 `测试报告.md`。`case-status.csv` 使用中文表头，包含“快速验证状态”“Browser验证状态”“自动化结论”“Browser验证原因”。JSONL 保持英文机器字段；每条新 execution 必有 `stage=fast|browser`。

结束前 Skill 先调用随包 `ai-auto-test-store validate-jsonl`，再检查：

- 每个输入用例均在状态表恰好一次，并至少经过快速或 Browser 阶段；
- execution ID 与“用例 ID + stage + attempt”唯一；
- Browser 覆盖率、强制选择与降级结果可解释；
- 状态表、执行记录、summary、测试报告和人工处理一致；
- Browser 截图存在且路径正确；快速阶段没有截图；
- 结果不含密码或验证码。

Schema 5 两阶段运行必须新建 run。Schema 1～4 的单阶段历史可以由 CLI 校验，但不得恢复或追加。

## 5. 恢复与复测

恢复时保持目标 URL、账号配置、用例选择、运行模式、请求/生效阶段和 Schema 一致。每个阶段独立生成下一 attempt；禁止修改旧 JSONL 行。任何历史重复、状态指针错误、覆盖率遗漏或报告不一致都属于硬失败，必须标记 `self_check=failed`、`run_valid=false` 并新建 run。

## 6. 回归建议

先以同一批 10 条用例分别运行：

1. `stage=browser mode=development`，作为当前 Browser 基线；
2. `stage=auto browser_coverage=60% mode=development`，验证两阶段。

固定环境、账号初始状态、用例、Cursor 和 MCP 版本；对比结论、人工处理、Browser 覆盖率、token 消耗、执行时间和结果完整性。
