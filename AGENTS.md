# S-UI Custom 发布规则

发布仓库为 `beacon8/s-ui-custom`，默认分支为 `main`。这里的 `frontend/` 是随主仓库提交的普通目录。首次发布为 `v1.6.9`。

## 发行范围

只发布适用于 Debian 等 Linux 系统的两个静态 ELF 安装包。32 位 x86 使用 `linux/386`，64 位 x86 使用 `linux/amd64`。每个正式 Release 的自定义附件必须恰好为以下四项（GitHub 自动生成的源码下载除外）：

- `s-ui-linux-386.tar.gz`
- `s-ui-linux-amd64.tar.gz`
- `install.sh`
- `SHA256SUMS`

## 提交和推送

1. 先核对 `git status -sb`、`git remote -v` 和 `gh repo view --json nameWithOwner`。推送目标应为上述仓库。若从开发仓库同步，请同时核对其前端子模块提交，保留发布仓库自己的安装器、升级地址、工作流和本规则。提交前检查暂存区，排除数据库、密钥、环境文件、本机日志和构建缓存。
2. 从最新正式版递增版本，使用 `v主.次.修订` 标签。更新 `config/version`、前端 package.json / package-lock.json 的项目版本，以及 README 的固定版本安装命令。确认远端尚无同名标签。
3. 在 frontend 目录执行 `npm ci && npm test && npm run build`，将新 dist 同步到主仓库的 web/html 后执行 `go test ./...`。对安装器和菜单分别执行 `bash -n install.sh`、`bash -n s-ui.sh`。安装启动测试在 Linux 测试机进行，macOS 只做静态检查和构建。
4. 审核差异后将发布提交推到 main，再推带注释标签。已有标签保持原样，在已有版本上修复代码应发新版本。

```bash
tag="v$(cat config/version)"
git diff --cached --stat
git commit -m "release: S-UI Custom $tag"
git push origin main
git tag -a "$tag" -m "S-UI Custom $tag"
git push origin "refs/tags/$tag"
```

## 构建和验收

标签触发 `.github/workflows/release.yml`。用 `gh run list --workflow release.yml --commit "$(git rev-parse HEAD)"` 定位，再用 `gh run watch RUN_ID --exit-status` 跟到结束。若标签推送没有触发，先确认没有在跑的发布任务，再执行 `gh workflow run release.yml --ref "$tag"`。重试失败任务可用 `gh run rerun RUN_ID --failed`。

工作流构建前端并嵌入两个 Linux 二进制，验证 ELF 架构、归档和校验清单，在草稿中检查四个附件及 SHA256 后正式发布。若任一步失败，先处理失败原因再发布。发布后用 `gh release view "$tag" --json url,isDraft,assets` 核对正式状态和附件，并匿名验证公开的 SHA256SUMS 与两个包的下载地址。草稿附件对匿名用户返回 404，公开下载通过后再给用户安装命令。安装命令中的脚本和安装包应固定到同一标签：

```bash
SUI_LANG=zhcn bash <(curl -fsSL "https://raw.githubusercontent.com/beacon8/s-ui-custom/${tag}/install.sh") "$tag"
```
