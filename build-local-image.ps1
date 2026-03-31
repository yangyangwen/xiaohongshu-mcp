param(
    [string]$ImageTag = "xiaohongshu-mcp:local",
    [string]$BaseImage = "xpzouying/xiaohongshu-mcp:latest",
    [string]$GoVersion = "1.24.0"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$artifact = Join-Path $repoRoot ".codex-local-build-app"
$goTar = "go$GoVersion.linux-amd64.tar.gz"
$goUrl = "https://golang.google.cn/dl/$goTar"

Write-Host "检查本地基础镜像: $BaseImage"
docker image inspect $BaseImage | Out-Null

Write-Host "在临时容器内编译 Linux 二进制..."
docker run --rm `
  -v "${repoRoot}:/src" `
  -w /src `
  $BaseImage `
  bash -lc "set -e;
    apt-get update -qq >/dev/null;
    apt-get install -y -qq wget ca-certificates tar >/dev/null;
    wget -q -O /tmp/$goTar $goUrl;
    rm -rf /usr/local/go;
    tar -C /usr/local -xzf /tmp/$goTar;
    export PATH=/usr/local/go/bin:`$PATH;
    export GOPROXY=https://goproxy.cn,direct;
    export GOSUMDB=sum.golang.google.cn;
    go version;
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /src/.codex-local-build-app ."

if (-not (Test-Path $artifact)) {
    throw "编译产物不存在: $artifact"
}

Write-Host "构建本地镜像: $ImageTag"
docker build -f (Join-Path $repoRoot "Dockerfile.local") -t $ImageTag $repoRoot

Write-Host "完成: $ImageTag"
