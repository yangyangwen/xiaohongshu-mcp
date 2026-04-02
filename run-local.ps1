param(
    [string]$ContainerName = "xiaohongshu-mcp",
    [string]$ImageTag = "xiaohongshu-mcp:local",
    [int]$HostPort = 18060,
    [string]$DataDir = "C:\Users\Administrator\xiaohongshu-mcp\data",
    [string]$ImagesDir = "C:\Users\Administrator\xiaohongshu-mcp\images",
    [string]$AuthToken = $env:XHS_AUTH_TOKEN
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
    Write-Host "Stopping and removing existing container: $ContainerName"
    docker rm -f $ContainerName | Out-Null
}

Write-Host "Starting container: $ContainerName"
$dockerArgs = @(
    "-d",
    "--name", $ContainerName,
    "-p", "${HostPort}:18060",
    "-e", "COOKIES_PATH=/app/data/cookies.json",
    "-v", "${DataDir}:/app/data",
    "-v", "${ImagesDir}:/app/images"
)

if ($AuthToken) {
    $dockerArgs += @("-e", "XHS_AUTH_TOKEN=$AuthToken")
}

$dockerArgs += $ImageTag
docker run @dockerArgs | Out-Null

Write-Host "Container started: http://localhost:$HostPort/mcp"
