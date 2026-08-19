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

1. 能读取两条只读正流程用例；
2. Preflight 能区分 Browser 平台故障和业务失败；
3. 能逐条调用 Browser 并获得结构化结论；
4. 每条完成后立即形成执行记录和当前状态；
5. 失败重试不会覆盖旧记录；
6. 中断后能从尚未完成的用例继续；
7. 汇总能列出所有需要人工处理的用例。
