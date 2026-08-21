# Cursor 两阶段闭环验证提示词

## 使用方式

1. 复制 `examples/closed-loop/run-config.example.md` 为同目录的 `run-config.local.md`，填写测试账号；不要提交该文件。
2. 在 Cursor 3.15.x Agent Window 选择 `/execute-test-cases`。
3. 若要验证第一阶段，先在 Cursor 中安装并启用 Playwright MCP；不要手动选择 `/use-browser`。

## 提示词

```text
使用 execute-test-cases Skill 执行两阶段最小闭环验证。

输入：
- 运行参数：stage=auto browser_coverage=60% mode=development
- 运行配置：examples/closed-loop/run-config.local.md
- 只读用例：examples/closed-loop/test-cases.md

要求：
1. 不读取或修改待测项目代码，不调用内部 API，不访问数据库，不修复问题。
2. 只按来源用例的原步骤和预期测试；不得用其他 URL 或业务页面替代来源指定的直链与断言。
3. 不新增、修改或删除业务数据；不在结果或回复中写入账号密码或验证码。
4. 先运行输入和两阶段前置检查。Playwright MCP 不可用时，询问我“安装后重试”或“跳过第一阶段”；我若选择跳过，则改为全部用例 Browser 验证。
5. 第一阶段仅使用 Playwright MCP 的结构化无障碍快照和允许的简单操作；不截图、不运行任意 JavaScript、不调用视觉或网络能力。
6. 第二阶段按 Skill 的强制规则和 60% 覆盖率选择；所有选择或未选择均写入中文原因。
7. 创建 Schema 5 的新 run，使用随 Skill 交付的 ai-auto-test-store 写入 JSONL。不要恢复或覆盖 Schema 1～4 的运行。
8. 完成后生成中文 summary.md 和 测试报告.md，明确两阶段状态、Browser 覆盖率、选择原因、人工处理和证据限制，并完成自检。
```

## 预期检查点

- `case-status.csv` 有“快速验证状态”“Browser验证状态”“自动化结论”“Browser验证原因”列；
- `case-executions.jsonl` 的每条新记录均有 `stage=fast|browser`；
- 快速验证没有截图；Browser 阶段在开发模式保存截图；
- `stage=auto` 下所有输入用例都至少经过一个阶段；
- Playwright MCP 不可用且选择跳过时，`effective_stage=browser` 且 Browser 覆盖率为 100%；
- Schema、状态、执行历史、summary 和 `测试报告.md` 自检一致。
