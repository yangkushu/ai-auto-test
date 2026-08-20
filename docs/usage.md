# 使用说明

## 1. 准备输入

复制示例配置并填写本地测试账号：

```text
examples/closed-loop/run-config.example.md
  → examples/closed-loop/run-config.local.md
```

`*.local.md` 已被 Git 忽略。用例可以参考 `examples/closed-loop/test-cases.md`，每条至少包含 ID、账号、前置条件、有序步骤和逐步预期。

## 2. 启动测试

在 Cursor Agent Window 中：

1. 确认 `Customize → Skills` 可见 `execute-test-cases`；
2. 在输入框选择 `/use-browser`；
3. 提交：

```text
使用 execute-test-cases Skill 执行以下测试：

- 运行配置：examples/closed-loop/run-config.local.md
- 测试用例：examples/closed-loop/test-cases.md

先执行 Browser Preflight。只通过可见浏览器 UI 测试，不读取代码、API 或数据库，不在结果中记录凭据。每条完成后先追加执行记录，再更新状态，最后运行结果校验器。
```

Skill 默认输出：

```text
.ai-auto-test/results/<run-id>/
├── case-status.csv
├── case-executions.jsonl
├── summary.md
└── *.png
```

Cursor 也支持在 `/` 菜单中直接搜索 `/execute-test-cases`。本 MVP 需要同时进入原生 Browser，因此当前推荐保留已经验证过的 `/use-browser` 入口，并在任务正文中明确写出 `execute-test-cases`；Skill 会根据描述自动匹配。

## 3. 校验结果

```bash
node scripts/validate-run.mjs .ai-auto-test/results/<run-id> --strict
```

也可以使用 npm script：

```bash
npm run validate-run -- .ai-auto-test/results/<run-id> --strict
```

输出 `VALIDATION PASSED` 且退出码为 0 才表示文件结构与跨表关系通过。`--strict` 会把 warning 也视为失败；机器消费时增加 `--json`。

校验器检查：

- 必需文件和 CSV 表头；
- 状态值、时间和 `manual_required`；
- JSONL 可解析性、唯一 execution ID 和连续 attempt；
- 状态表是否指向最新执行；
- summary 计数是否与状态表一致；
- 截图路径是否越界、缺失或为空；
- 常见密码/验证码字段是否误写入结果。

## 4. 恢复与复测

- 恢复时沿用原 run-id，先读取状态和执行历史；
- 只执行 `pending` 或 `retest_pending`；
- 复测分配新的 attempt 和 execution ID；
- `case-executions.jsonl` 只追加，禁止修改旧行；
- 如果状态与执行历史冲突，以追加式历史重建状态，并记录修复原因。

## 5. 权限和编码

- 只允许 Cursor 写入当前结果目录，不授权修改待测项目代码；
- JSONL、Markdown 使用 UTF-8，CSV 使用 UTF-8 BOM；
- Windows PowerShell 读取时显式增加 `-Encoding UTF8`；
- 截图复制到结果目录后记录相对路径，建议在 summary 中记录 SHA-256。
