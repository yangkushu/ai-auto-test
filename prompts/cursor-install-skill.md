# Cursor 安装 Skill 提示词

请安装或更新 GitHub 仓库 `https://github.com/yangkushu/ai-auto-test` 中的 `execute-test-cases` Skill。不要只读取网页或只导入 `SKILL.md`；必须把完整 Skill 和二进制安装到本机用户级 Skill 目录。

要求：

1. 下载或克隆仓库到临时目录，定位 `.agents/skills/execute-test-cases/`。不要读取或修改待测项目代码。
2. 读取 `VERSION`，检测本机 OS 与 Architecture。当前仅支持 Windows x64 和 Linux x64；其他平台停止并报告不受支持。
3. 检查源目录包含 `SKILL.md`、`VERSION`、`agents/openai.yaml`、`bin/ai-auto-test-store-windows-amd64.exe` 和 `bin/ai-auto-test-store-linux-amd64`。
4. 为本次安装创建临时目标目录，完整复制 `execute-test-cases/` 目录到其中，包括 `bin/`；不要只复制 Markdown 或一个二进制。
5. 在临时目标目录执行当前平台的二进制：Windows x64 使用 `bin/ai-auto-test-store-windows-amd64.exe`，Linux x64 使用 `bin/ai-auto-test-store-linux-amd64`。运行 `version`，确认 JSON 输出中的 `ok=true` 且 `version` 与 `VERSION` 完全一致。
6. 若用户级 `~/.agents/skills/execute-test-cases/` 已存在，先将其保留为临时备份；仅当第 3～5 步成功后，才以校验后的新目录替换它。新目录验证失败时恢复旧目录。
7. 重载 Cursor 或提示我新建 Agent 会话，再确认 `/execute-test-cases` 可见。
8. 只有“完整目录已安装 + 当前平台 CLI 版本一致 + Cursor 已发现 Skill”同时满足时，报告安装成功；否则报告具体受阻项，不要宣称成功。

不安装 Node.js、npm、Python 或 Go。不要在安装 Skill 时改动 MCP 配置；运行默认的两阶段测试前，如需快速验证，由我在 Cursor 中另行安装 Playwright MCP。安装期间不要读取或输出测试账号、密码、Cookie、Token 或验证码。
