# Cursor Browser PoC 当前状态

## 结论

截至 2026-08-19，Cursor Agent 原生 Browser 已完成最小可行性验证。Agent 能够打开网页、执行搜索、访问本地测试页面、填写登录表单，并根据页面可见内容确认登录成功。

这证明“自然语言指令 → Browser Agent → 页面操作与观察 → 测试结论”的核心技术路线可行。它尚未证明批量测试流水线已经完成。

## 验证环境

| 项目 | 值 |
|---|---|
| Cursor 版本 | 3.15.6（User Setup） |
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

## 尚未验证

- 主 Agent 从文件读取多条用例并逐条调度 Browser；
- Browser 结果回传后自动更新用例状态；
- 每条用例完成后追加执行记录，且中断后可以恢复；
- 截图等证据的稳定命名、保存和引用；
- 多账号用例及账号切换；
- 短信 mock 的“浏览器自动填入”；
- Excel 输入与输出；
- 飞书文档下载、回写与同步；
- 大批量用例的稳定性、耗时和成本；
- Cursor 之外的平台迁移。

## 下一里程碑

使用两条相互独立的正流程用例验证最小完整闭环：

```text
读取用例文件
  → Browser Preflight
  → 逐条执行
  → 每条立即写入状态
  → 追加执行记录
  → 汇总结果与人工处理清单
```

验收标准：

1. 两条用例都获得唯一结论；
2. 每个结论都有实际步骤、可见观察和时间；
3. 状态表能查询当前结论；
4. 执行记录保留每次尝试，后续复测不会覆盖旧记录；
5. 任意一步中断后，能够识别已完成和仍待测试的用例。
