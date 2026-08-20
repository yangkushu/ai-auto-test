# 安装说明

## 环境要求

- Cursor Desktop，使用 Agent Window；PoC 已在 Windows x64、Cursor 3.15.19 验证；
- Git；
- 可访问的测试环境和专用测试账号；
- Cursor 内置 `/use-browser` 可正常打开目标页面。

仅执行 Browser 测试不需要 Node.js。以下功能需要 Node.js 20 或更高版本：

- 运行确定性结果校验器；
- 运行项目自身的自动化测试；
- 参与校验器开发。

## 克隆项目

```bash
git clone git@github.com:yangkushu/ai-auto-test.git
cd ai-auto-test
```

项目不依赖第三方 npm 包，无需运行 `npm install`。Node.js 也不参与 Browser 的页面读取与操作。

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

## 可选：校验器与项目自测

如果已安装 Node.js：

```bash
node --version
npm test
```

测试通过说明结果校验器能够运行。没有 Node.js 时可以跳过这一步，但运行结果必须注明“确定性结果校验未执行”。Cursor Browser 必须另外通过 Preflight 验证，因为它属于 Cursor 运行时能力。
