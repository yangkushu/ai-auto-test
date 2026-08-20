# 安装说明

## 环境要求

- Cursor Desktop，使用 Agent Window；PoC 已在 Windows x64、Cursor 3.15.19 验证；
- 可访问的测试环境和专用测试账号；
- Cursor 原生 Browser 能力可用。

首版随 Skill 交付跨平台编译后的结果写入器，不要求安装 Go、Python、Node.js、npm、Playwright 或额外 Browser MCP。Git 仅用于克隆和更新项目，也可以直接下载仓库 ZIP。

## 推荐：从 GitHub 导入

Cursor 官方支持从 GitHub 仓库导入 Skill：

1. 打开侧边栏 `Customize`；
2. 进入 `Rules`，点击 `Add Rule`；
3. 选择 `Remote Rule (Github)`；
4. 输入 `https://github.com/yangkushu/ai-auto-test`；
5. 重载 Cursor 或新建 Agent 会话；
6. 在 `Customize → Skills` 中确认 `execute-test-cases` 可见。

官方支持的目录和 GitHub 导入入口见 [Cursor Agent Skills](https://cursor.com/docs/skills)。如果当前 Cursor 版本的界面文字不同，以 `Customize` 中的 GitHub Remote Rule/Skill 导入入口为准。

如果希望让 Cursor Agent 代办，可把[标准安装提示词](../prompts/cursor-install-skill.md)发给它。自然语言不是非交互安装 API；网络、文件权限或 Cursor UI 能力不足时，Agent 应报告受阻而不是宣称安装成功。

## 备选：克隆或下载

已安装 Git 时：

```bash
git clone git@github.com:yangkushu/ai-auto-test.git
cd ai-auto-test
```

未安装 Git 时，可以从 GitHub 下载仓库 ZIP，解压后直接使用 Cursor 打开目录。

项目没有最终用户依赖安装步骤。下载完成后可以直接使用 Cursor 打开项目。

仓库已经包含 Windows、macOS 和 Linux 的 x64/ARM64 `ai-auto-test-store`。Skill 会按当前平台选择对应文件并校验版本，不需要单独安装。

## 让 Cursor 发现本地 Skill

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

## 验证安装完成

只有以下检查全部通过，才算安装成功：

1. `Customize → Skills` 能看到 `execute-test-cases`；
2. Skill 目录包含 `SKILL.md`、`VERSION`、`agents/` 和 `bin/`；
3. 当前平台对应的 `ai-auto-test-store` 存在且能执行；
4. CLI 输出版本与 `VERSION` 一致；
5. 新 Agent 会话中能选择 `/execute-test-cases`。

可以在 Windows x64 上手动确认写入器版本：

```powershell
& '.agents\skills\execute-test-cases\bin\ai-auto-test-store-windows-amd64.exe' version
```

输出中的 `version` 应与同目录 `VERSION` 一致。

macOS/Linux 应选择对应的 `darwin-*` 或 `linux-*` 文件。Skill 每次运行都会自动重复这项版本检查；缺文件、不可执行或版本不一致时，会在启动 Browser 前停止。

## Browser Preflight

直接调用 `/execute-test-cases`；Skill 会在内部启动 Browser Preflight。不要手动选择 `/use-browser`。若出现：

```text
MCP server "cursor-ide-browser" not found
```

说明 Cursor Browser 组件没有注册；停止测试，不要把业务用例标记为失败。PoC 中重新安装最新版 Cursor 后恢复。

只有可见健康应用界面且匹配配置的成功标志才算通过。503、其他 4xx/5xx 或浏览器连接错误页必须得到运行级 `blocked`，不能进入业务用例。

## 维护者构建

只有维护二进制文件时才需要 Go，版本以仓库 `.go-version` 为准。修改 CLI 或更新 Skill 版本后运行：

```bash
./scripts/build-result-store.sh
```

该脚本使用 Go 标准库交叉编译六个平台文件；最终使用者不执行此步骤。

GitHub Actions 在 pull request 和 `main` push 时自动执行测试、静态检查、六平台构建与交付校验。推送与 `VERSION` 一致的 `v<version>` tag 后，会自动发布完整 Skill `tar.gz`、六个平台二进制和 `SHA256SUMS`。详见[构建、分发与安装设计](distribution.md)。
