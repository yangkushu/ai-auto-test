# Cursor 安装 Skill 提示词

请安装 GitHub 仓库 `https://github.com/yangkushu/ai-auto-test` 中的 `execute-test-cases` Skill。

要求：

1. 优先使用 Cursor 官方的 GitHub Skill 导入能力；如果当前会话无法操作 Customize，则下载仓库并把完整的 `.agents/skills/execute-test-cases/` 安装到用户级 `~/.agents/skills/execute-test-cases/`。
2. 必须保留 `SKILL.md`、`VERSION`、`agents/` 和 `bin/`，不能只复制 `SKILL.md`。
3. 不安装 Node.js、npm、Python、Go、Playwright 或额外 Browser MCP，不执行待测项目代码。
4. 当前仅支持 Windows x64 和 Linux x64：选择 `bin/ai-auto-test-store-windows-amd64.exe` 或 `bin/ai-auto-test-store-linux-amd64`，执行其 `version` 命令，并确认输出版本与 `VERSION` 完全一致；其他平台报告不受支持。
5. 安装后重载 Cursor 或提示我新建 Agent 会话，再确认 Skills 中能看到 `/execute-test-cases`。
6. 只有目录完整、CLI 版本一致且 Skill 被 Cursor 发现时才报告安装成功；否则报告具体受阻项，不要宣称成功。

安装期间不要读取或输出测试账号、密码、Cookie、Token 或验证码。
