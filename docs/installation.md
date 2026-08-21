# 安装说明

## 环境要求

- Cursor Desktop，使用 Agent Window；PoC 已在 Windows x64、Cursor 3.15.19 验证；
- 可访问的测试环境和专用测试账号；
- Cursor 原生 Browser 能力可用。

首版随 Skill 交付 Windows x64 与 Linux x64 编译后的结果写入器，不要求安装 Go、Python、Node.js、npm、Playwright 或额外 Browser MCP。Git 仅用于克隆和更新项目，也可以直接下载仓库 ZIP。

## 推荐：让 Cursor Agent 安装或更新

本项目的主安装入口是向 Cursor Agent 提供仓库 URL：

```text
安装或更新这个 Skill：https://github.com/yangkushu/ai-auto-test
```

仅给出 URL 不足以证明安装完成。Agent 必须遵守以下安装合同：

1. 下载或克隆该仓库到临时位置；不能只读取 GitHub 网页中的 `SKILL.md`；
2. 定位 `.agents/skills/execute-test-cases/`，读取其 `VERSION`；
3. 检测运行 Cursor 的主机平台；首版只允许 Windows x64 与 Linux x64；
4. 校验源目录包含 `SKILL.md`、`VERSION`、`agents/`、`bin/` 及对应平台二进制；
5. 将**完整目录**安装到用户级 `~/.agents/skills/execute-test-cases/`，包括 `bin/`；不得只复制 `SKILL.md`、提示词或单个二进制；
6. Windows x64 必须安装并运行 `bin/ai-auto-test-store-windows-amd64.exe`；Linux x64 必须安装并运行 `bin/ai-auto-test-store-linux-amd64`；
7. 执行该二进制的 `version` 命令，输出版本必须与 `VERSION` 文件完全相同；
8. 重载 Cursor 或新建 Agent 会话，并确认 `/execute-test-cases` 可见；
9. 仅当上述检查全部通过时，报告“安装成功”。

完整可复制 Prompt 见[安装提示词](../prompts/cursor-install-skill.md)。网络、文件权限或 Cursor 能力不足时，Agent 必须报告具体受阻项，不能把“已读取 URL”或“已复制 Markdown”报告为安装成功。

### 二进制如何随 Skill 安装

二进制不是通过用户机器安装 Go 后生成，也不是由测试时的 Cursor 自动下载。它们已经随 Git 仓库中的 Skill 目录交付：

```text
.agents/skills/execute-test-cases/
├── SKILL.md
├── VERSION
├── agents/openai.yaml
└── bin/
    ├── ai-auto-test-store-windows-amd64.exe
    └── ai-auto-test-store-linux-amd64
```

Agent 从 GitHub 获得完整仓库后，将上述目录整体复制到 `~/.agents/skills/execute-test-cases/`。因此当前平台二进制自然随 Skill 安装；运行时不下载二进制，不要求 Node.js、Python 或 Go。

### 更新规则

用户再次发送同一 URL 即表示请求更新。Agent 必须：

1. 先下载新版本到临时目录并完成“目录完整 + 当前平台 `version`”校验；
2. 校验通过后才替换用户级现有目录；
3. 替换前保留旧目录作为临时备份；
4. 新目录重载后无法被 Cursor 发现或二进制不可执行时，恢复旧目录并报告更新失败；
5. 不删除旧版本，也不覆盖现有安装，除非新版本已经通过校验。

此过程不读取待测项目、测试账号、密码、Cookie、Token 或验证码。

## 备选：Cursor GitHub 导入

Cursor 官方支持从 GitHub 仓库导入 Skill：在 `Customize → Rules → Add Rule → Remote Rule (Github)` 中输入仓库 URL。官方支持的目录和 GitHub 导入入口见 [Cursor Agent Skills](https://cursor.com/docs/skills)。

该方式可以帮助 Cursor 发现 Skill，但仍必须按上节检查二进制目录是否完整、当前平台 CLI 是否可执行且版本一致；如果导入只得到规则或 Markdown，应改用“让 Agent 安装或更新”。

## 备选：克隆或下载

已安装 Git 时：

```bash
git clone git@github.com:yangkushu/ai-auto-test.git
cd ai-auto-test
```

未安装 Git 时，可以从 GitHub 下载仓库 ZIP，解压后直接使用 Cursor 打开目录。

项目没有最终用户依赖安装步骤。下载完成后可使用 Cursor 打开项目，或者完整复制 Skill 目录到用户级位置。

仓库已经包含 Windows x64 与 Linux x64 的 `ai-auto-test-store`。Skill 会按当前平台选择对应文件并校验版本，不需要单独安装。macOS 和 ARM64 当前不受支持。

## 让 Cursor 发现本地 Skill

项目 Skill 位于：

```text
.agents/skills/execute-test-cases/SKILL.md
```

Cursor 会自动发现项目中的 `.agents/skills/`。克隆后使用 Cursor 打开仓库，重新创建 Agent 会话；在侧边栏 `Customize → Skills` 中应能看到 `execute-test-cases`。这是 Cursor 官方支持的项目级 Skill 目录，详见 [Cursor Agent Skills](https://cursor.com/docs/skills)。

如果只在本仓库中维护配置和用例，项目级安装已经足够。如果希望打开任意待测项目时都能使用，必须使用用户级目录；这也是 Agent 安装合同使用的目标位置。

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

Linux x64 应使用 `ai-auto-test-store-linux-amd64`。Skill 每次运行都会自动重复这项版本检查；macOS、ARM64、缺文件、不可执行或版本不一致时，会在启动 Browser 前停止。

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

该脚本使用 Go 标准库交叉编译 Windows x64 和 Linux x64 文件；最终使用者不执行此步骤。

GitHub Actions 在 pull request 和 `master` push 时自动执行测试、静态检查、双平台构建与交付校验。推送与 `VERSION` 一致的 `v<version>` tag 后，会自动发布完整 Skill `tar.gz`、两个平台二进制和 `SHA256SUMS`。详见[构建、分发与安装设计](distribution.md)。
