# 使用说明

## 1. 准备输入

复制示例配置并填写本地测试账号：

```text
examples/closed-loop/run-config.example.md
  → examples/closed-loop/run-config.local.md
```

`*.local.md` 已被 Git 忽略。用例可以参考 `examples/closed-loop/test-cases.md`，每条至少包含 ID、账号、前置条件、有序步骤和逐步预期。

运行配置中的 `mode` 支持：

- `normal`：正常模式，记录恢复与审计所需的关键事件；`passed` 不保存截图，异常用例和 Preflight 失败保存截图；
- `development`：开发模式，额外记录步骤观察、UI 适应、截图和关键文件写入事件；每条实际执行用例均保存截图。

未填写时默认 `normal`。本次 Prompt 中的 `mode=...` 优先于运行配置；恢复已有 run 时沿用原模式。

## 2. 启动测试

在 Cursor Agent Window 中：

1. 确认 `Customize → Skills` 可见 `execute-test-cases`；
2. 在输入框选择 `/execute-test-cases`；如果菜单未显示，则在 Prompt 中明确要求使用项目的 `execute-test-cases` Skill；
3. 提交：

```text
使用 execute-test-cases Skill 执行以下测试：

- 运行参数：mode=development
- 运行配置：examples/closed-loop/run-config.local.md
- 测试用例：examples/closed-loop/test-cases.md

先按输入校验、Browser 能力、目标导航和应用识别执行分阶段前置检查（Preflight）。只通过可见浏览器 UI 测试，不读取代码、API 或数据库，不在结果中记录凭据。Preflight 失败时以“事实、推断、建议”说明问题，业务用例保持 `pending`。每条完成后先追加执行记录，再更新状态，最后重新读取结果文件完成一致性自检。
```

Skill 默认输出：

```text
.ai-auto-test/results/<run-id>/
├── case-status.csv
├── case-executions.jsonl
├── run-events.jsonl
├── summary.md
├── 测试报告.md
└── *.png
```

Browser 由 Skill 内部调用。不要手动选择 `/use-browser`；否则只能验证 Cursor Browser，不能证明本 Skill 的编排、状态和审计逻辑生效。

`测试报告.md` 是面向测试、研发和产品的中文交付物，优先查看。它包含结论、范围、中文用例明细、人工处理、测试数据影响和证据/限制。`case-status.csv` 使用中文表头和中文展示值，例如“已通过”“测试受阻”“是”“待人工清理”。`case-executions.jsonl` 和 `run-events.jsonl` 保持英文机器字段与状态值，供 CLI 校验与恢复使用。

## 3. 检查结果

Skill 使用随包交付的 `ai-auto-test-store` 初始化、写入和校验两个 JSONL，再执行跨文件检查：

- 必需文件和 CSV 表头；
- 中文状态表表头、中文状态/布尔值/清理状态与最新执行记录的映射；
- JSONL 可解析性、唯一 execution ID 和连续 attempt；
- “用例 ID + attempt”组合唯一；
- 事件中的 Skill 版本、Schema 版本、run ID、模式和关键写入顺序；
- 开发模式下来源步骤事件一一对应、Browser action 前后配对；
- 状态表是否指向最新执行；
- summary 计数是否与状态表一致；
- `测试报告.md` 是否存在、中文表达完整，且统计与状态表一致；
- 截图路径是否越界、缺失或为空；开发模式每条执行都必须有截图，正常模式只要求异常用例和 Preflight 失败有截图；
- 常见密码/验证码字段是否误写入结果。

自检结论写入 `summary.md`。正常完成必须满足 `self_check=passed`、`run_valid=true`。CLI 校验失败、execution ID 重复、“用例 ID + attempt”重复、状态指针错误或开发步骤事件缺失属于硬失败，必须标记 `self_check=failed`、`run_valid=false`，不能只写 warning。Agent 只能根据唯一追加历史重建状态和汇总，禁止修改旧执行记录；不能安全修复时保留原始 run，并进入运行级人工处理。

当前 Skill 版本读取自 `.agents/skills/execute-test-cases/VERSION`。`summary.md` 和每条运行事件都必须记录 `skill_version`、`schema_version` 和 `mode`。Schema 2 及之后的事件必须使用 `self_check_finished`，不能使用已废弃的 `self_check`；Schema 1、2、3 的历史结果可校验但不能在 Schema 4 下恢复追加。`summary.md`、`测试报告.md`、`case-status.csv`、事件消息、观察和建议使用中文；代码、URL、JSONL 机器字段与原始页面证据保持原样。

## 4. 恢复与复测

- 恢复时沿用原 run-id，先读取状态和执行历史；
- 只执行 `pending` 或 `retest_pending`；
- 复测分配新的 attempt 和 execution ID；
- 在执行前和追加前分别检查 execution ID 与“用例 ID + attempt”不存在；
- `case-executions.jsonl` 只追加，禁止修改旧行；
- 如果状态与唯一执行历史冲突，以追加式历史重建状态，并记录修复原因；历史本身存在重复时将 run 标记为无效并新建 run 重跑。

## 5. 权限和编码

- 只允许 Cursor 写入当前结果目录，不授权修改待测项目代码；
- JSONL、Markdown 使用 UTF-8，CSV 使用 UTF-8 BOM；
- JSONL 禁止直接编辑，只能由随 Skill 交付的 CLI 追加和校验；
- Windows PowerShell 读取时显式增加 `-Encoding UTF8`；
- 截图复制到结果目录后记录相对路径，建议在 summary 中记录 SHA-256。

## 6. 对比两个 Skill 版本

分别使用旧版本和候选版本创建独立 run，禁止覆盖结果目录。对比时固定测试环境、用例文件、账号初始状态和 Cursor 版本，并均使用 `development` 模式。

至少比较：

- 最终结论和人工复核后的正确性；
- `blocked`/`inconclusive` 是否合理；
- 用例完成率、执行时间和中断位置；
- execution、status、summary 是否一致；
- 过程事件能否解释关键操作和写入顺序。

每次对比在结论中明确列出 baseline 与 candidate 的 `skill_version` 和 run ID。
