# Linux 发行支持设计文档

**日期**: 2026-05-19  
**分支**: `feature/linux-support`  
**作者**: Agent  

---

## 背景

当前 `release.yml` 仅构建并发行 macOS（darwin/arm64）和 Windows（windows/amd64）的产物。Linux 用户只能本地手动编译 CLI，无法从 Release 页面直接下载预编译二进制。

代码层面已具备 Linux 兼容性：
- `internal/lingmaipc/pipe_other.go` 已处理非 Windows 平台的 pipe 逻辑
- `desktop/app.go` 的 `OpenPathInFileManager` 已使用 `xdg-open`
- CLI 入口 `cmd/lingma-ipc-proxy/main.go` 是纯 Go，可交叉编译
- Desktop 基于 Wails v2，官方支持 Linux（需 GTK/WebKit 依赖）

## 目标

为 GitHub Release 增加 Linux 预编译产物，覆盖：
- **CLI**: amd64 + arm64
- **Desktop**: amd64 + arm64

## 方案选择

采用**保守追加方案**（方案 1）。理由：
1. 不改动现有 macOS/Windows 构建逻辑，风险最低
2. 直接追加独立 job，review 简单
3. 后续如需矩阵化重构，可单独进行

## 具体改动

### 1. `.github/workflows/release.yml`

#### 新增 `build-cli-linux` job

```yaml
  build-cli-linux:
    name: Build CLI Linux
    runs-on: ubuntu-latest
    needs: test
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Build CLI amd64
        run: |
          mkdir -p dist-amd64
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o dist-amd64/lingma-proxy ./cmd/lingma-ipc-proxy
          tar -C dist-amd64 -czf "lingma-proxy_${RELEASE_TAG}_linux_amd64.tar.gz" lingma-proxy
      - name: Build CLI arm64
        run: |
          mkdir -p dist-arm64
          CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o dist-arm64/lingma-proxy ./cmd/lingma-ipc-proxy
          tar -C dist-arm64 -czf "lingma-proxy_${RELEASE_TAG}_linux_arm64.tar.gz" lingma-proxy
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: cli-linux
          path: lingma-proxy_${{ env.RELEASE_TAG }}_linux_*.tar.gz
```

#### 新增 `build-desktop-linux-amd64` job

```yaml
  build-desktop-linux-amd64:
    name: Build Desktop Linux amd64
    runs-on: ubuntu-latest
    needs: test
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: npm
          cache-dependency-path: desktop/frontend/package-lock.json
      - name: Install Wails dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
      - name: Install frontend dependencies
        run: npm ci --prefix desktop/frontend
      - name: Build app
        run: |
          cd desktop
          wails build -platform linux/amd64 -clean
      - name: Package app
        run: |
          binary_path="desktop/build/bin/LingmaProxy"
          test -f "$binary_path"
          tar -C desktop/build/bin -czf "lingma-proxy-desktop_${RELEASE_TAG}_linux_amd64.tar.gz" LingmaProxy
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: desktop-linux-amd64
          path: lingma-proxy-desktop_${{ env.RELEASE_TAG }}_linux_amd64.tar.gz
```

#### 新增 `build-desktop-linux-arm64` job

与 `amd64` 完全一致，仅差异：
- `runs-on: ubuntu-24.04-arm`
- `wails build -platform linux/arm64 -clean`
- 产物名：`lingma-proxy-desktop_${RELEASE_TAG}_linux_arm64.tar.gz`
- artifact name：`desktop-linux-arm64`

#### 更新 `publish` job

在 `needs` 中追加：
- `build-cli-linux`
- `build-desktop-linux-amd64`
- `build-desktop-linux-arm64`

### 2. `desktop/app.go`

修复 `titleOS` 函数，使 Linux 显示为 `"Linux"` 而非 `"linux"`：

```go
func titleOS(osName string) string {
	switch osName {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	default:
		return osName
	}
}
```

## 构建产物

打 tag 后 Release 页面新增：

| 产物文件名 | 类型 | 平台/架构 |
|---|---|---|
| `lingma-proxy_vX.Y.Z_linux_amd64.tar.gz` | CLI | linux/amd64 |
| `lingma-proxy_vX.Y.Z_linux_arm64.tar.gz` | CLI | linux/arm64 |
| `lingma-proxy-desktop_vX.Y.Z_linux_amd64.tar.gz` | Desktop | linux/amd64 |
| `lingma-proxy-desktop_vX.Y.Z_linux_arm64.tar.gz` | Desktop | linux/arm64 |

## 风险与缓解

| 风险 | 缓解措施 |
|---|---|
| `ubuntu-24.04-arm` runner 在某些 GitHub 组织中未启用 | Desktop arm64 构建失败不会阻塞其他平台；如不可用可降级为仅发行 amd64 Desktop |
| Wails Linux 依赖导致构建失败 | 针对 `ubuntu-latest` (即 24.04) 及 `ubuntu-24.04-arm` 需使用 `libwebkit2gtk-4.1-dev`；构建失败会在 PR/test 阶段暴露 |
| 现有 macOS/Windows 发行被破坏 | 保守追加方案不动任何现有 job 逻辑，仅追加新 job |

## 测试计划

1. **本地验证**: 在本地 Linux 环境（或 Docker）执行 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/lingma-ipc-proxy`，确认 CLI 可编译
2. **CI 验证**: 推送分支后，手动触发 `workflow_dispatch` 或在 PR 中观察 Actions 运行结果
3. **产物验证**: 下载构建产物，在目标 Linux 环境中解压并运行，确认二进制可执行
