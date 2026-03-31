param(
    [string]$ContainerName = "xiaohongshu-mcp",
    [string]$ImageTag = "xiaohongshu-mcp:local",
    [int]$HostPort = 18060,
    [string]$DataDir = "C:\Users\Administrator\xiaohongshu-mcp\data",
    [string]$ImagesDir = "C:\Users\Administrator\xiaohongshu-mcp\images"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $DataDir)) {
    New-Item -ItemType Directory -Path $DataDir | Out-Null
}

if (-not (Test-Path $ImagesDir)) {
    New-Item -ItemType Directory -Path $ImagesDir | Out-Null
}

$existing = docker ps -a --filter "name=^${ContainerName}$" --format "{{.Names}}"
if ($existing) {
    Write-Host "停止并删除旧容器: $ContainerName"
    docker rm -f $ContainerName | Out-Null
}

Write-Host "启动新容器: $ContainerName"
docker run -d `
  --name $ContainerName `
  -p "${HostPort}:18060" `
  -e "COOKIES_PATH=/app/data/cookies.json" `
  -v "${DataDir}:/app/data" `
  -v "${ImagesDir}:/app/images" `
  $ImageTag | Out-Null

Write-Host "容器已启动: http://localhost:$HostPort/mcp"
