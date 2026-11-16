# winget 清单（`jjjjjjjjnnjnn.AI-Terminal-Pet`）

这里的目录结构就是 `microsoft/winget-pkgs` 仓库里的结构
（`manifests/<首字母>/<Publisher>/<PackageName>/<Version>/`），
提交时**原样复制这一个目录**，不要改布局。

```
manifests/j/jjjjjjjjnnjnn/AI-Terminal-Pet/1.0.15/
  jjjjjjjjnnjnn.AI-Terminal-Pet.yaml                    # version
  jjjjjjjjnnjnn.AI-Terminal-Pet.locale.en-US.yaml       # defaultLocale
  jjjjjjjjnnjnn.AI-Terminal-Pet.installer.yaml          # installer
```

## 为什么这个包能装也能卸

winget 对 `InstallerType: exe` 的产品，卸载走的是 **Windows 自己的「设置 › 应用」那一行**。
1.0.15 起 `scripts/install.ps1` 会写 `HKCU\...\Uninstall\AI-Terminal-Pet`
（`DisplayName` / `DisplayVersion` / `InstallLocation` / `QuietUninstallString`），
并把 `install.ps1` 自己复制进安装目录，所以：

- 静默安装 = `pet-setup.exe --install`（默认行为就是这样，`InstallerSwitches.Silent` 只是把它写明）；
- 静默卸载 = 那一行里的 `QuietUninstallString`，即
  `powershell -NoProfile -ExecutionPolicy Bypass -File "<安装目录>\install.ps1" -Uninstall -Force -Destination "<安装目录>\AI-Terminal-Pet.exe"`；
- 用户侧的 `pet-setup.exe --uninstall --force` 是同一件事的另一条入口。

`tests/fresh-machine.mjs` 的 S5/S8 就是按这条契约验的：**执行注册表里记的那条命令**，
然后断言目录与那一行都不在了。

## 每次发版要改什么

```powershell
# 1. 让发布资产先生出来（tag 之后才有稳定 URL）
gh release view pet-v1.0.15 --json assets --jq '.assets[].name'

# 2. 复制上一版目录，改三个值：
#    - 三份清单里的 PackageVersion
#    - installer.yaml 里的 InstallerUrl（新 tag）与 InstallerSha256（新 exe 的真实摘要）
#    - locale 里的 ReleaseNotesUrl（如果换了指向）
$sha = (Get-FileHash .\dist\pet-setup.exe -Algorithm SHA256).Hash
# 3. 本地校验：必须打印“清单验证成功”
winget validate --manifest .\packaging\winget\manifests\j\jjjjjjjjnnjnn\AI-Terminal-Pet\1.0.15

# 4. 提交到 microsoft/winget-pkgs：一个分支只加这一个版本目录
#    （他们的 pipeline 会自己下载 exe 校验 sha256、跑静默安装与静默卸载）
```

## 两个必须提前说清的限制

1. **`pet-setup.exe` 没有 Authenticode 签名。** winget 的自动校验会真的下载并静默执行，
   而 SmartScreen 可能拦住未签名的新下载文件。若校验因此失败，**不要**改成"解压版"清单
   （`InstallerType: zip` + `portable` 只会把 exe 解到 PATH 里，没有 `protocol.js`、没有 bundle、
   没有自启动，等于 `docs/ROADMAP.md` §5 明令禁止的降级安装）。正确的处理是记录失败原因、
   保留 1.0.15 的安装器改进（它对**所有** Windows 用户都有效），等有证书再说。
2. **清单不随 npm/tag 自动前进。** 每个版本要一个目录、一次 PR；仓库里保留历史目录，
   因为它们是"当时发了什么"的记录。
