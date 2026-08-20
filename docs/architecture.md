# MVP 架构设计

## 总体

MVP 是 Cursor-first 的测试编排 Skill，不自研浏览器驱动。Cursor 主 Agent 负责读取用例、控制执行顺序和写入结果；Browser 子 Agent 通过 `/use-browser` 负责可见页面的观察与操作。

```text
run config ───────────────┐
                         ▼
read-only test cases → Cursor orchestration Skill
                         │
                         ├─ Browser Preflight
                         │
                         ├─ /use-browser executor
                         │       └─ step observations + verdict + evidence
                         │
                         ├─ case_status（可更新）
                         ├─ case_execution（只追加）
                         └─ summary + 人工处理清单
```

## 组件职责

### 主 Agent / Skill

- 校验输入是否齐全；
- 执行 Browser Preflight；
- 读取尚未完成的用例；
- 每次只向 Browser 提交一条用例；
- 校验 Browser 返回结果是否包含步骤观察和唯一结论；
- 先追加执行记录，再更新当前状态；
- 生成批次汇总和人工处理清单。

### Browser 子 Agent

- 只通过可见 UI 建立前置条件；
- 按顺序执行用例步骤；
- 每一步都检查对应预期；
- 只适应 UI 机械细节，不改变业务步骤和期望；
- 返回结构化观察、结论和证据引用；
- 不读代码、不调用内部 API、不访问数据库、不修复缺陷。

### 文件仓储

- `test_case`：只读输入；
- `case_status`：当前结果投影，可更新；
- `case_execution`：追加式历史，禁止覆盖；
- `summary/manifest`：保存批次级环境、版本、时间和统计。

文件编码约定：

- `case-executions.jsonl`：UTF-8，无 BOM；
- `summary.md`：UTF-8；
- `case-status.csv`：UTF-8 BOM，确保 Windows Excel 可直接识别中文；
- Windows PowerShell 读取文本时显式使用 `-Encoding UTF8`。

### 结果校验器

`scripts/validate-run.mjs` 是独立于 Browser Agent 的确定性检查层，使用 Node.js 内置模块，不依赖模型判断。它在每批完成后检查：

- 三个结果文件是否存在且格式可解析；
- 状态、执行记录、最新 attempt 与 summary 统计是否一致；
- 异常结论是否进入人工处理；
- 截图是否存在、非空并位于当前结果目录；
- 输出是否可能泄露密码或验证码。

校验器退出码为 `0` 才表示通过；CI 或正式使用建议增加 `--strict`，将 warning 也视为失败。它校验的是结果闭环完整性，不重新判断页面业务结论是否正确。

## Browser Preflight

正式执行前必须验证：

1. `/use-browser` 能启动；
2. 目标 URL 能打开；
3. Agent 能读取页面标题或主要可见内容。

若出现 `cursor-ide-browser not found` 等工具注册错误，整次运行标记为 `blocked`，不得给任何业务用例写入 `failed`。

## 单条用例事务顺序

```text
读取 pending/retest_pending 用例
  → 创建 attempt
  → Browser 执行
  → 校验返回结果
  → 追加 case_execution
  → 更新 case_status
  → 继续下一条
```

先写执行记录再更新状态，避免状态存在但历史证据缺失。若进程在两次写入之间中断，可根据最后记录重建状态。

恢复时必须沿用原 run-id，读取状态和历史记录，只执行 `pending`/`retest_pending`。复测生成新的 attempt 和 execution ID，禁止修改旧 JSONL 行。受控跨会话恢复已经验证；两次写入之间的突然崩溃恢复仍需验证。

若 Cursor 需要额外文件写权限，只允许当前结果目录。截图从 Cursor 临时目录复制后应记录相对路径和 SHA-256；相同页面允许产生相同哈希，但不得覆盖旧证据文件。

## Browser 返回契约

```ts
type BrowserCaseResult = {
  caseId: string;
  status: 'passed' | 'failed' | 'blocked' | 'inconclusive';
  startedAt: string;
  finishedAt: string;
  steps: Array<{
    stepIndex: number;
    action: string;
    expected: string;
    observed: string;
    result: 'passed' | 'failed' | 'blocked' | 'inconclusive';
  }>;
  reason?: string;
  finalUrl?: string;
  pageTitle?: string;
  evidence: Array<{
    kind: 'screenshot' | 'console_log' | 'network_log';
    uri: string;
  }>;
};
```

如果缺少可见观察，主 Agent 不得接受 `passed`。如果 Browser 无法提供足够证据，应返回 `inconclusive`。

## 人工处理规则

MVP 不在执行前为用例分类人工阶段。Browser 执行完成后：

- `failed`：进入人工复核/缺陷确认；
- `blocked`：进入环境、账号或能力处理；
- `inconclusive`：进入人工判定；
- `passed`：默认不要求人工，但测试人员可以追加抽查或主链路复测。

## MVP 验收

1. 已验证：读取外部只读正流程用例；
2. 已验证：Preflight 能确认 Browser 可用，并识别工具注册故障；
3. 已验证：逐条调用 Browser 并获得结构化结论；
4. 已验证：每条完成后形成执行记录和当前状态；
5. 已验证：复测追加新 attempt，不覆盖旧记录；
6. 已验证：受控中断后从 pending 用例继续；
7. 已实现：确定性校验器检查格式、跨表一致性、证据和凭据泄露；
8. 待验证：真实页面的异常结论能正确进入人工处理清单；
9. 待验证：突然崩溃后能自动重建状态。
