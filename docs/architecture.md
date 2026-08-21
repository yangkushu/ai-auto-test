# MVP 架构设计

## 总体

MVP 是 Cursor-first 的测试编排 Skill，不自研浏览器驱动。用户直接调用 `execute-test-cases`；Skill 负责读取用例和控制顺序，在内部调用 Cursor 原生 Browser，并使用随包交付的 CLI 写入关键结果。

```text
run config ───────────────┐
                         ▼
read-only test cases → Cursor orchestration Skill
                         │
                         ├─ Browser Preflight
                         │
                         ├─ Cursor native Browser
                         │       └─ step observations + verdict + evidence
                         │
                         ├─ case_status（可更新）
                         ├─ ai-auto-test-store
                         │       ├─ case_execution（只追加）
                         │       └─ run_event（只追加）
                         └─ summary + 人工处理清单
```

## 组件职责

### 主 Agent / Skill

- 校验输入是否齐全；
- 执行 Browser Preflight；
- 读取尚未完成的用例；
- 每次只向 Browser 提交一条用例；
- 校验 Browser 返回结果是否包含步骤观察和唯一结论；
- 通过结果写入器先追加执行记录，再更新当前状态；
- 按运行模式追加关键或详细过程事件；
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
- `case_status`：当前结果投影，可更新；以中文表头、中文状态和中文清理状态供测试人员直接查看；
- `case_execution`：追加式历史，禁止覆盖，包含写操作副作用和清理状态；
- `run_event`：追加式过程事件，用于观察、中断定位和恢复审计；
- `summary/manifest`：保存批次级环境、版本、时间和统计。

`ai-auto-test-store` 使用 Go 标准库实现并随 Skill 提供编译后的平台可执行文件。它负责初始化空 JSONL、单对象压缩、UTF-8 无 BOM 追加、即时复验，以及 execution ID/attempt 唯一性；不读取测试账号，不操作 Browser，也不改变业务结论。最终使用者无需安装 Go 或其他语言运行时。

## 构建与分发架构

运行架构与构建架构分离：Cursor 和静态 CLI 属于用户运行时；Go、shell 构建脚本和 GitHub Actions 只属于维护者侧。CI 在 Linux runner 上使用 Go 交叉编译 Windows x64 和 Linux x64 两个目标，因此 CI 使用 shell 不会给最终用户引入 shell 或 PowerShell 依赖。

仓库中的 `.agents/skills/execute-test-cases/` 是 Cursor GitHub 导入所需的完整交付单元，包含 Windows x64 和 Linux x64 二进制。每次 push 都重新构建和检查；版本 tag 还会生成不可变 GitHub Release、完整 Skill `tar.gz`、独立平台文件和 SHA-256 清单。

安装采用两级策略：MVP 使用 Cursor 官方 GitHub Skill 导入或本地 `.agents/skills/` 发现；稳定后增加 Agent Plugin/Cursor Marketplace 作为一键安装与更新通道。自由文本要求 Agent 安装 URL 只能触发这个流程，不能取代安装后的目录、版本和 Skill 发现校验。完整设计见[构建、分发与安装设计](distribution.md)。

文件编码约定：

- `case-executions.jsonl`：UTF-8，无 BOM；
- `run-events.jsonl`：UTF-8，无 BOM；
- `summary.md`：UTF-8；
- `case-status.csv`：UTF-8 BOM，确保 Windows Excel 可直接识别中文；
- Windows PowerShell 读取文本时显式使用 `-Encoding UTF8`。

### 结果自检

主 Agent 在每批完成后先调用 `ai-auto-test-store validate-jsonl`，再重新读取结果文件并检查：

- 四个结果文件是否存在且格式可解析；
- 状态、执行记录、最新 attempt 与 summary 统计是否一致；
- 异常结论是否进入人工处理；
- 截图是否存在、非空并位于当前结果目录；
- 输出是否可能泄露密码或验证码。

自检结果写入 `summary.md`。JSONL 语法、编码和 execution 唯一性由 CLI 确定性检查；状态投影、步骤语义、证据和 summary 的跨文件判断仍由 Agent 完成，不能把整套自检描述成完全确定性校验。

### 运行模式与事件

`mode` 只接受 `normal` 和 `development`。新运行按“Prompt 参数 > 运行配置 > normal”解析；恢复运行沿用原模式。模式不得改变浏览器操作、结论或人工处理规则。

正常模式保留批次、Preflight、用例开始、execution 追加、状态更新、自检和结束等关键事件；只为异常用例和 Preflight 失败保存截图。开发模式额外保留与来源步骤一一对应的步骤观察、脱敏后的 Browser action 前后事件、UI 适应、截图、关键写入前后和自检细节；每条实际执行用例均保存截图。事件逐行追加到 `run-events.jsonl`，不得记录凭据、Cookie、Token、输入值或内部推理。过程事件用于审计解释，不等同于独立工具证明。

### 版本体系

- Skill 版本：读取 `.agents/skills/execute-test-cases/VERSION`，使用 Semantic Versioning；
- Schema 版本：当前为 `3`，将状态表升级为中文展示；
- run ID：唯一标识一次测试运行。

每次运行必须将三者和生效模式写入 summary；每条过程事件必须包含 Skill 版本、Schema 版本、run ID 和模式。Schema 1、2 的历史运行仍可校验；从 Schema 2 起，自检结束事件固定为 `self_check_finished`，Schema 3 使用中文状态表。候选版本使用 `dev`/`rc` 后缀，正式版本使用 Git Tag。

版本升级规则：

- MAJOR：状态语义、执行流程或结果契约发生不兼容变化；
- MINOR：兼容增加运行模式、输入来源或测试能力；
- PATCH：修复 Prompt、判定规则或文档，不改变兼容契约。

`0.x` 阶段使用 `0.<功能版本>.<修订>`，开发候选依次使用 `-dev.N`、`-rc.N`。任何可能影响操作、判定或结果结构的修改都必须更新 Skill 版本和 Changelog。

## 分阶段前置检查（Preflight）

正式执行前依次验证：

1. `input_validation`：输入、账号、平台二进制与版本；
2. `browser_capability`：Skill 能在内部启动 Cursor 原生 Browser；
3. `target_navigation`：目标 URL、最终 URL、标题、可见错误和重定向；
4. `application_identity`：页面不是连接失败、4xx/5xx 或反向代理错误页，并匹配成功标志或明确应用界面。

`input_validation` 在创建 run 前完成，输入无效时不创建运行结果；输入通过会写入 `run_started`。其余阶段在开发模式下各自记录开始和通过事件。失败时必须将“事实、推断、建议”写入运行级人工处理项。503 只能确认目标 URL 未获得应用页面，不能直接断言“等待环境恢复”；域名迁移、服务故障和路径错误都只是带置信度的可能原因。所有业务用例保持 `pending`，不得创建业务 execution。目标 URL 或运行配置发生变化时必须新建 run；仅同一配置下的暂时性故障才可恢复原 run。

## 单条用例事务顺序

```text
读取 pending/retest_pending 用例
  → 创建 attempt
  → Browser 执行
  → 校验返回结果
  → ai-auto-test-store 候选校验
  → CLI 追加 case_execution 并即时复验
  → 追加 execution_appended 事件
  → 更新 case_status
  → 追加 status_updated 事件
  → 继续下一条
```

先写执行记录再更新状态，避免状态存在但历史证据缺失。执行前和追加前都必须检查 execution ID 与“用例 ID + attempt”唯一。若发生冲突，不得追加；记录完整性错误、停止新的业务执行并进入失败自检。若进程在两次写入之间中断，可根据最后一条唯一记录重建状态。

恢复时必须沿用原 run-id，读取状态和历史记录，只执行 `pending`/`retest_pending`。恢复前目标 URL、账号配置、用例选择、运行模式和 Schema 版本必须一致；任一项改变都新建 run。复测生成新的 attempt 和 execution ID，禁止修改旧 JSONL 行。Schema 1、2 的历史记录仍可由 CLI 校验，但不得在 Schema 3 下继续追加。历史已经存在重复 ID 或重复 attempt 时无法在只追加约束下修复，应保留 run、标记 `run_valid=false` 并新建 run 重跑。受控跨会话恢复已经验证；两次写入之间的突然崩溃恢复仍需验证。

若 Cursor 需要额外文件写权限，只允许当前结果目录。截图从 Cursor 临时目录复制后应记录相对路径和 SHA-256；相同页面允许产生相同哈希，但不得覆盖旧证据文件。

## 测试数据副作用

创建、新增、保存、提交、发布、修改、删除、支付等会改变业务数据的步骤属于写操作。每条 execution 必须记录 `sideEffects`：是否写操作、可见资源标识、是否创建、是否需要清理及清理状态。

固定测试数据不得擅自加后缀。执行写操作前，若 UI 已显示同名/同标识对象，或无法可靠建立对象唯一性，标记 `blocked` 与 `reasonCode=test_data_conflict`，不要把历史数据当作本次创建证据。只有来源用例明确允许动态值时才使用 `AUTO-<run-id>-<case-id>`。没有来源清理步骤时不得自动删除，进入独立的测试数据清理清单；这不改变该用例的业务结论或 `manual_required`。

## Browser 返回契约示意

以下 TypeScript 仅用于表达字段结构，不是首版运行代码，也不要求安装 TypeScript 或 Node.js：

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
  reasonCode?: string;
  finalUrl?: string;
  pageTitle?: string;
  evidence: Array<{
    kind: 'screenshot' | 'console_log' | 'network_log';
    uri: string;
  }>;
  sideEffects: {
    hasWriteOperation: boolean;
    resources: Array<{
      resourceType: string;
      visibleIdentifier: string;
      created: boolean;
    }>;
    cleanupRequired: boolean;
    cleanupStatus: 'not_applicable' | 'completed' | 'pending_manual';
    cleanupReason?: string;
  };
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
7. 已实测：mode 参数、开发模式、过程事件和版本落盘；
8. 已发现、待回归：Agent 能发现重复 execution ID，但旧版未使无效 run 失败；
9. 已实测、待校准：真实页面异常能进入人工清单，但目标控件缺失存在误判为 `blocked` 的风险；
10. 待验证：突然崩溃后能自动重建状态；
11. 后续评估：是否需要把当前 JSONL CLI 扩展为完整的跨文件确定性校验器。
12. 待验证：从 GitHub Remote Rule 导入完整 Skill 后，Windows x64 能发现 Skill 并执行随包 CLI；
13. 已实现、待线上验收：GitHub Actions Windows/Linux x64 构建和 tag Release。
