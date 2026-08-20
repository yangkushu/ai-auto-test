# 安装说明

## 环境要求

- Cursor Desktop，使用 Agent Window；PoC 已在 Windows x64、Cursor 3.15.19 验证；
- 可访问的测试环境和专用测试账号；
- Cursor 内置 `/use-browser` 可正常打开目标页面。

首版不要求安装 Node.js、npm、Playwright 或额外 Browser MCP。Git 仅用于克隆和更新项目，也可以直接下载仓库 ZIP。

## 克隆项目

已安装 Git 时：

```bash
git clone git@github.com:yangkushu/ai-auto-test.git
cd ai-auto-test
```

未安装 Git 时，可以从 GitHub 下载仓库 ZIP，解压后直接使用 Cursor 打开目录。

项目没有安装脚本或依赖安装步骤。下载完成后直接使用 Cursor 打开项目。

## 让 Cursor 发现 Skill

项目 Skill 位于：

```text
.agents/skills/execute-test-cases/SKILL.md
```

Cursor 会自动发现项目中的 `.agents/skills/`。克隆后使用 Cursor 打开仓库，重新创建 Agent 会话；在侧边栏 `Customize → Skills` 中应能看到 `execute-test-cases`。这是 Cursor 官方支持的项目级 Skill 目录，详见 [Cursor Agent Skills](https://cursor.com/docs/skills)。

如果只在本仓库中维护配置和用例，项目级安装已经足够。如果希望打开任意待测项目时都能使用，推荐复制到用户级目录。

Windows PowerShell：

```powershell
$target = Join-Path $HOME '.agents\skills\execute-test-cases'
New-Item -ItemType Directory -Force $target | Out-Null
Copy-Item -Recurse -Force '.agents\skills\execute-test-cases\*' $target
```

macOS / Linux：

```bash
mkdir -p ~/.agents/skills
cp -R .agents/skills/execute-test-cases ~/.agents/skills/
```

最终用户级位置为：

```text
~/.agents/skills/execute-test-cases/
```

复制后重启 Cursor 或执行 `Developer: Reload Window`。不要把包含测试账号的本地配置提交到 Skill 目录。

## Browser Preflight

正式运行前，在 Agent 输入框选择 `/use-browser`，要求它打开一个测试 URL 并报告页面标题。若出现：

```text
MCP server "cursor-ide-browser" not found
```

说明 Cursor Browser 组件没有注册；停止测试，不要把业务用例标记为失败。PoC 中重新安装最新版 Cursor 后恢复。

完成 Preflight 即表示首版运行环境就绪。
