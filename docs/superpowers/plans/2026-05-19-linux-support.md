# Linux 发行支持 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 GitHub Release 中增加 Linux CLI 和 Desktop 的预编译产物（amd64 + arm64），同时修复 Desktop 应用在 Linux 下的 OS 名称显示。

**Architecture:** 采用保守追加方案，不动现有 macOS/Windows 构建逻辑，仅在 `release.yml` 中追加 3 个 Linux build job 并更新 publish 依赖。Desktop 的 Linux 兼容通过 Wails 官方支持的 GTK/WebKit 依赖实现。

**Tech Stack:** GitHub Actions, Go (cross-compile), Wails v2, Ubuntu 24.04/24.04-arm

---

## Task 1: 修复 desktop/app.go titleOS 函数

**Files:**
- Modify: `desktop/app.go:1643-1652`

- [ ] **Step 1: 修改 titleOS 函数，增加 Linux 分支**

找到 `titleOS` 函数，在 `"windows"` 分支后增加 `"linux"` 分支：

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

- [ ] **Step 2: 验证编译通过**

Run: `cd /Users/cutiechi/sources/Lutiancheng1/lingma-proxy/desktop && go build ./...`
Expected: 无错误输出

- [ ] **Step 3: Commit**

```bash
git add desktop/app.go
git commit -m "fix(desktop): correct OS title label for Linux"
```

---

## Task 2: 在 release.yml 中添加 build-cli-linux job

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: 在 build-cli-windows job 结束后插入 build-cli-linux job**

找到 `build-cli-windows` job 的结束位置（第 102 行 `path: lingma-proxy_${{ env.RELEASE_TAG }}_windows_amd64.zip` 之后），在其后插入以下内容：

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

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci(release): add Linux CLI build job (amd64 + arm64)"
```

---

## Task 3: 在 release.yml 中添加 build-desktop-linux-amd64 job

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: 在 build-desktop-windows job 结束后插入 build-desktop-linux-amd64 job**

找到 `build-desktop-windows` job 的结束位置（第 207 行 `path: lingma-proxy-desktop_${{ env.RELEASE_TAG }}_windows_amd64.zip` 之后），在其后插入以下内容：

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

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci(release): add Linux Desktop amd64 build job"
```

---

## Task 4: 在 release.yml 中添加 build-desktop-linux-arm64 job

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: 在 build-desktop-linux-amd64 job 结束后插入 build-desktop-linux-arm64 job**

在 Task 3 插入的内容之后，继续插入以下内容：

```yaml
  build-desktop-linux-arm64:
    name: Build Desktop Linux arm64
    runs-on: ubuntu-24.04-arm
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
          wails build -platform linux/arm64 -clean

      - name: Package app
        run: |
          binary_path="desktop/build/bin/LingmaProxy"
          test -f "$binary_path"
          tar -C desktop/build/bin -czf "lingma-proxy-desktop_${RELEASE_TAG}_linux_arm64.tar.gz" LingmaProxy

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: desktop-linux-arm64
          path: lingma-proxy-desktop_${{ env.RELEASE_TAG }}_linux_arm64.tar.gz
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci(release): add Linux Desktop arm64 build job"
```

---

## Task 5: 更新 publish job 的 needs

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: 在 publish needs 中追加 3 个 Linux job**

找到 `publish` job 的 `needs` 部分（约第 212-216 行），将其替换为：

```yaml
    needs:
      - build-cli-macos
      - build-cli-windows
      - build-cli-linux
      - build-desktop-macos
      - build-desktop-windows
      - build-desktop-linux-amd64
      - build-desktop-linux-arm64
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci(release): add Linux build jobs to publish dependencies"
```

---

## Task 6: 验证 release.yml 语法

**Files:**
- Validate: `.github/workflows/release.yml`

- [ ] **Step 1: 使用 yamllint 或在线工具检查 YAML 语法**

Run:
```bash
python3 -c "import yaml; yaml.safe_load(open('/Users/cutiechi/sources/Lutiancheng1/lingma-proxy/.github/workflows/release.yml'))" && echo "YAML valid"
```

Expected: `YAML valid`

如果系统没有 PyYAML，可用以下命令安装：
```bash
pip3 install pyyaml
```

- [ ] **Step 2: 确认无语法错误后结束验证**

如果 YAML 解析失败，回到对应 Task 修复缩进或格式问题。

---

## Task 7: 本地编译验证 Linux CLI

**Files:**
- Validate: `cmd/lingma-ipc-proxy/main.go`

- [ ] **Step 1: 交叉编译 linux/amd64**

Run:
```bash
cd /Users/cutiechi/sources/Lutiancheng1/lingma-proxy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/lingma-proxy-linux-amd64 ./cmd/lingma-ipc-proxy
```

Expected: 无输出（编译成功）

- [ ] **Step 2: 交叉编译 linux/arm64**

Run:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /tmp/lingma-proxy-linux-arm64 ./cmd/lingma-ipc-proxy
```

Expected: 无输出（编译成功）

- [ ] **Step 3: 清理临时产物**

Run:
```bash
rm -f /tmp/lingma-proxy-linux-amd64 /tmp/lingma-proxy-linux-arm64
```

---

## Task 8: 最终检查与推送

- [ ] **Step 1: 查看完整提交历史**

Run:
```bash
cd /Users/cutiechi/sources/Lutiancheng1/lingma-proxy && git log --oneline feature/linux-support
```

Expected 输出示例（顺序可能不同）：
```
xxxxxxx ci(release): add Linux build jobs to publish dependencies
xxxxxxx ci(release): add Linux Desktop arm64 build job
xxxxxxx ci(release): add Linux Desktop amd64 build job
xxxxxxx ci(release): add Linux CLI build job (amd64 + arm64)
xxxxxxx fix(desktop): correct OS title label for Linux
xxxxxxx docs: add Linux support release spec
```

- [ ] **Step 2: 确认所有改动已提交**

Run:
```bash
git status
```

Expected: `nothing to commit, working tree clean`

- [ ] **Step 3: 推送分支到远程**

Run:
```bash
git push -u origin feature/linux-support
```

Expected: 分支成功推送到 origin
