# xiaohongshu-mcp 鉴权接入文档

## 1. 保护范围

当前版本默认要求鉴权的入口：

- `/api/v1/*`
- `/mcp`
- `/mcp/*`

默认放开的入口：

- `/health`

## 2. 鉴权方式

服务端使用共享 Bearer Token。

服务端环境变量：

```bash
XHS_AUTH_TOKEN=your-long-random-token
```

客户端请求头：

```http
Authorization: Bearer your-long-random-token
```

校验规则：

- 服务端读取自己容器内的 `XHS_AUTH_TOKEN`
- 客户端必须传 `Authorization: Bearer <token>`
- 两边完全一致才放行
- 不带或带错返回 `401`
- 服务端未配置 `XHS_AUTH_TOKEN` 返回 `503`

## 3. HTTP API 调用示例

### 3.1 GET 获取笔记互动数据

```bash
curl "http://127.0.0.1:18060/api/v1/feeds/metrics?url=https://www.xiaohongshu.com/explore/69bcad55000000001a0209ae?source=webshare&xhsshare=pc_web&xsec_token=your_xsec_token&xsec_source=pc_share" \
  -H "Authorization: Bearer your-long-random-token"
```

### 3.2 POST 获取笔记互动数据

```bash
curl -X POST "http://127.0.0.1:18060/api/v1/feeds/metrics" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-long-random-token" \
  -d "{\"url\":\"https://www.xiaohongshu.com/explore/69cb1ee2000000001f001509?source=webshare&xhsshare=pc_web&xsec_token=your_xsec_token&xsec_source=pc_share\"}"
```

成功响应示例：

```json
{
  "success": true,
  "data": {
    "liked_count": "3",
    "comment_count": "2",
    "shared_count": "2",
    "collected_count": "1"
  },
  "message": "获取Feed互动数据成功"
}
```

## 4. MCP 调用示例

如果 MCP 客户端支持自定义 Header，在连接配置中加入：

```http
Authorization: Bearer your-long-random-token
```

直接用 HTTP 调 MCP 示例：

```bash
curl -X POST "http://127.0.0.1:18060/mcp" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "Authorization: Bearer your-long-random-token" \
  -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2025-03-26\",\"capabilities\":{},\"clientInfo\":{\"name\":\"demo\",\"version\":\"1.0.0\"}}}"
```

## 5. Docker 运行示例

### 5.1 docker run

```bash
docker run -d \
  --name xiaohongshu-mcp \
  --restart unless-stopped \
  -p 18060:18060 \
  -e ROD_BROWSER_BIN=/usr/bin/google-chrome \
  -e COOKIES_PATH=/app/data/cookies.json \
  -e XHS_AUTH_TOKEN=your-long-random-token \
  -v /your/data:/app/data \
  -v /your/images:/app/images \
  xiaohongshu-mcp:local
```

### 5.2 docker compose

`.env`：

```bash
XHS_AUTH_TOKEN=your-long-random-token
```

`docker-compose.yml`：

```yaml
environment:
  - XHS_AUTH_TOKEN=${XHS_AUTH_TOKEN}
```

## 6. Linux 服务器替换现网容器

你当前线上容器是老版本镜像：

```text
swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/xpzouying/xiaohongshu-mcp:v1.1.8
```

这个版本没有你现在二开后的鉴权和新接口能力。最稳的替换方式是：

- 用你自己的 GitHub 仓库拉代码
- 在服务器本机 build 新镜像
- 备份旧容器数据后替换容器

你的仓库：

- `https://github.com/yangyangwen/xiaohongshu-mcp`

### 第 1 步：登录服务器

```bash
ssh your-user@your-server
```

### 第 2 步：查看旧容器是否挂载了数据目录

先确认旧容器的数据是不是已经落到宿主机了：

```bash
docker inspect xiaohongshu-mcp --format '{{json .Mounts}}'
```

如果输出里已经有宿主机路径挂载到：

- `/app/data`
- `/app/images`

那后面直接复用这些宿主机目录即可。

如果没有挂载，先备份容器内数据。

### 第 3 步：备份旧容器数据

先建一个备份目录：

```bash
mkdir -p /opt/xiaohongshu-mcp-backup
```

如果旧容器没有挂载卷，执行：

```bash
docker cp xiaohongshu-mcp:/app/data /opt/xiaohongshu-mcp-backup/data
docker cp xiaohongshu-mcp:/app/images /opt/xiaohongshu-mcp-backup/images
```

这一步主要是保住：

- `cookies.json`
- 已有图片目录
- 其他运行时数据

### 第 4 步：拉取你的二开仓库

如果服务器还没有代码：

```bash
cd /opt
git clone https://github.com/yangyangwen/xiaohongshu-mcp.git
cd /opt/xiaohongshu-mcp
```

如果服务器已经有代码：

```bash
cd /opt/xiaohongshu-mcp
git pull origin main
```

### 第 5 步：在服务器本机构建新镜像

```bash
cd /opt/xiaohongshu-mcp
docker build -t xiaohongshu-mcp:auth .
```

这样 build 出来的镜像就是你自己 fork 后二开代码对应的版本。

### 第 6 步：准备宿主机数据目录

建议统一放在：

```bash
mkdir -p /opt/xiaohongshu-mcp/data
mkdir -p /opt/xiaohongshu-mcp/images
```

如果你第 3 步备份了旧容器内的数据，再把它恢复回来：

```bash
cp -r /opt/xiaohongshu-mcp-backup/data/* /opt/xiaohongshu-mcp/data/ 2>/dev/null || true
cp -r /opt/xiaohongshu-mcp-backup/images/* /opt/xiaohongshu-mcp/images/ 2>/dev/null || true
```

### 第 7 步：停掉旧容器

```bash
docker rm -f xiaohongshu-mcp
```

建议先 build 完成再删旧容器，这样停机时间最短。

### 第 8 步：启动新容器

```bash
docker run -d \
  --name xiaohongshu-mcp \
  --restart unless-stopped \
  -p 18060:18060 \
  -e ROD_BROWSER_BIN=/usr/bin/google-chrome \
  -e COOKIES_PATH=/app/data/cookies.json \
  -e XHS_AUTH_TOKEN=your-long-random-token \
  -v /opt/xiaohongshu-mcp/data:/app/data \
  -v /opt/xiaohongshu-mcp/images:/app/images \
  xiaohongshu-mcp:auth
```

### 第 9 步：验证容器是否启动成功

查看容器：

```bash
docker ps | grep xiaohongshu-mcp
```

查看日志：

```bash
docker logs -f xiaohongshu-mcp
```

### 第 10 步：验证鉴权是否生效

健康检查，不需要 token：

```bash
curl http://127.0.0.1:18060/health
```

不带 token，应该返回 `401`：

```bash
curl http://127.0.0.1:18060/api/v1/login/status
```

带 token，应该返回 `200`：

```bash
curl http://127.0.0.1:18060/api/v1/login/status \
  -H "Authorization: Bearer your-long-random-token"
```

验证互动数据接口：

```bash
curl "http://127.0.0.1:18060/api/v1/feeds/metrics?url=https://www.xiaohongshu.com/explore/69bcad55000000001a0209ae?source=webshare&xhsshare=pc_web&xsec_token=ABB159aKu1mQvoDwJahKbOtRSRZ760gMglwt9hdSiKnG4%3D&xsec_source=pc_share" \
  -H "Authorization: Bearer your-long-random-token"
```

## 7. 后续更新方式

以后你自己的二开代码有更新，服务器继续这样升级：

```bash
cd /opt/xiaohongshu-mcp
git pull origin main
docker build -t xiaohongshu-mcp:auth .
docker rm -f xiaohongshu-mcp
docker run -d \
  --name xiaohongshu-mcp \
  --restart unless-stopped \
  -p 18060:18060 \
  -e ROD_BROWSER_BIN=/usr/bin/google-chrome \
  -e COOKIES_PATH=/app/data/cookies.json \
  -e XHS_AUTH_TOKEN=your-long-random-token \
  -v /opt/xiaohongshu-mcp/data:/app/data \
  -v /opt/xiaohongshu-mcp/images:/app/images \
  xiaohongshu-mcp:auth
```

## 8. 要不要 git 到服务器

不是必须，但对你当前这个二开项目来说，最简单的方案就是：

- 直接把你的 fork 仓库 `git clone` 到服务器
- 在服务器本机 `docker build`
- 然后替换容器

这是最适合你当前阶段的方案，原因是：

- 你已经有自己的 GitHub 仓库
- 你改的是源码，不是只改配置
- 服务器本机构建能直接适配服务器架构
- 后续更新也只要 `git pull + docker build + docker run`

如果你不想在服务器上放 git 仓库，也可以用另一套方式：

- 在本地或 CI 构建镜像
- 推到镜像仓库
- 服务器只执行 `docker pull` 和 `docker run`

那样也可以，但比“服务器直接 git + build”多了一层镜像发布流程。

## 9. 注意事项

- 如果旧容器没有挂载卷，删容器前一定先备份 `/app/data`
- 如果 `cookies.json` 没保住，替换后需要重新扫码登录
- 现在端口还是 `0.0.0.0:18060` 对外监听，鉴权虽然已加，但仍建议后续再收口 IP 或加防火墙
- 文档中写入的是当前线上 token，转发文档时等于同时转发了访问密钥
