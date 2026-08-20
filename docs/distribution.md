# 构建、分发与安装设计

## 总体结论

首版采用“GitHub Skill 仓库 + 随包静态二进制 + GitHub Release”的分发方式。最终使用者只依赖 Cursor；Go、shell 和 GitHub Actions 都是维护者侧构建工具，不是用户运行时依赖。

Cursor 官方支持从 GitHub 仓库导入 Skill，并自动发现 `.agents/skills/`。因此仓库中的 `execute-test-cases` 必须始终是一个完整交付单元：`SKILL.md`、`VERSION`、Agent 元数据和六个平台的 `ai-auto-test-store` 缺一不可。

依据：[Cursor Agent Skills](https://cursor.com/docs/skills)、[GitHub Actions 构建与测试 Go](https://docs.github.com/en/actions/tutorials/build-and-test-code/go)、[GitHub Release 完整性校验](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity)。

## 交付结构

```text
Git 源码
  ├─ Skill 说明与版本
  ├─ Go CLI 源码和测试
  └─ 六个平台的预编译 CLI
          ↓ push / pull request
GitHub Actions
  ├─ go test / go vet
  ├─ 交叉编译 Windows、Linux、macOS 的 x64/ARM64
  └─ 校验 Skill 交付文件
          ↓ version tag
GitHub Release
  ├─ 完整 Skill tar.gz
  ├─ 六个平台独立二进制
  └─ SHA256SUMS
```

支持目标：

| OS | Architecture | 文件 |
|---|---|---|
| Windows | x64 | `ai-auto-test-store-windows-amd64.exe` |
| Windows | ARM64 | `ai-auto-test-store-windows-arm64.exe` |
| Linux | x64 | `ai-auto-test-store-linux-amd64` |
| Linux | ARM64 | `ai-auto-test-store-linux-arm64` |
| macOS | Intel | `ai-auto-test-store-darwin-amd64` |
| macOS | Apple Silicon | `ai-auto-test-store-darwin-arm64` |

CLI 只使用 Go 标准库并设置 `CGO_ENABLED=0`，不依赖 Node.js、Python、PowerShell、动态 C 库或用户机器上的 Go 工具链。

为使随仓二进制可复现，维护者编译器固定为 `.go-version` 中的版本，构建同时启用 `-trimpath`、关闭 VCS 元数据并清空 Go build ID。CI 会重新构建并与仓库内二进制逐字节比较；源码、版本或编译器变化而未更新随仓产物时，检查直接失败。

## CI 与 Release

`.github/workflows/ci-release.yml` 在 pull request 和 `main` push 时执行测试、静态检查、六目标构建、随仓二进制一致性检查和交付检查。推送与 `VERSION` 一致的 `v<version>` tag 后，工作流额外创建 GitHub Release。

Release 规则：

1. `VERSION` 使用 Semantic Versioning；候选版本允许 `-dev.N`、`-rc.N`；
2. tag 必须严格等于 `v` 加 `VERSION`，否则发布失败；
3. Release 包由 tag 对应源码重新构建，不复用开发者机器产物；
4. Release 同时发布完整 Skill `tar.gz`、六个平台二进制和 `SHA256SUMS`；
5. 正式发布前必须在至少 Windows x64 和当前开发平台做 Cursor 回归。

仓库内预编译二进制服务于 Cursor 的 GitHub Skill 直接导入；Release 资产服务于不可变版本归档、校验和手动恢复。二者版本都必须与 `VERSION` 一致。

## Cursor 安装合同

### 官方入口

优先使用 Cursor 的官方 GitHub Skill 导入：

1. 打开 `Customize`；
2. 进入 `Rules` 并选择 `Add Rule`；
3. 选择 `Remote Rule (Github)`；
4. 输入 `https://github.com/yangkushu/ai-auto-test`；
5. 重载 Cursor 或新建 Agent 会话；
6. 在 Skills 中确认 `execute-test-cases` 可见。

Cursor 也会从项目级或用户级 `.agents/skills/` 自动发现 Skill。无法使用 UI 导入时，可以完整复制 `.agents/skills/execute-test-cases/`，但不能只复制 `SKILL.md`。

### 安装成功判定

“目录已复制”不等于安装完成。必须同时满足：

- Cursor Skills 列表中出现 `execute-test-cases`；
- `SKILL.md`、`VERSION`、`agents/` 和 `bin/` 都存在；
- 当前 OS/Architecture 对应二进制存在且可执行；
- 二进制 `version` 输出与 `VERSION` 一致；
- 新 Agent 会话能选择 `/execute-test-cases`。

Skill 自身在 Browser 启动前重复执行版本检查。检查失败时必须停止并报告安装受阻，不能绕过 CLI 直接写 JSONL。

## 自然语言安装的保证边界

用户可以告诉 Cursor：“安装这个 GitHub 仓库中的 Skill”。Agent 通常能够按仓库说明完成导入或复制，但自由文本 Prompt 不是稳定的安装 API，会受到 Cursor 版本、网络权限、终端权限和 UI 能力影响。因此项目不能承诺任何一句自然语言在所有环境中都必然成功。

项目能保证的是：

- 提供官方支持的 GitHub Skill 结构；
- 提供不依赖用户语言运行时的全平台产物；
- 提供明确且可自动检查的安装完成条件；
- 失败时安全停止，不产生“Skill 已安装但运行时缺工具”的假成功。

用于 Cursor Agent 的标准安装 Prompt 见 [cursor-install-skill.md](../prompts/cursor-install-skill.md)。

## Plugin 演进

Cursor Plugin/Agent Plugin 是更完整的分发形态，可进入 Marketplace，并由 Cursor 管理安装范围和更新。首版暂不把 Plugin 作为必需项，原因是当前只交付一个 Skill，GitHub Skill 导入已经能验证核心流程。

完成 `0.2.x` 稳定回归后再增加 Agent Plugin manifest 并提交 Marketplace。Marketplace 安装应作为未来的一键安装主通道；GitHub 仓库和 Release 继续作为可审计来源与离线恢复通道。
