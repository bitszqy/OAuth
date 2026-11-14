# Windows 环境搭建与运行说明

本文说明如何从一台 Windows 机器开始运行 lab00 和 lab01。

## 一、系统要求

- Windows 10/11 64 位。
- 开启 CPU 虚拟化。
- 至少 8 GB 内存，建议 16 GB。
- 管理员权限，用于首次安装 WSL、Docker 和修改 hosts。
- 可访问 Docker Hub 的网络环境。

检查 CPU 虚拟化：

1. 打开任务管理器。
2. 进入“性能”中的“CPU”。
3. 确认“虚拟化”显示为已启用。

如果未启用，需要在 BIOS 中开启 Intel VT-x 或 AMD-V。

## 二、安装 WSL2

以管理员身份打开 PowerShell：

~~~powershell
wsl --install
wsl --update
~~~

重启电脑并检查：

~~~powershell
wsl --status
wsl --list --verbose
~~~

Windows Home 必须使用 WSL2 后端。

## 三、安装 Docker Desktop

从官方网站下载 Docker Desktop for Windows：

<https://www.docker.com/products/docker-desktop/>

安装时选择：

~~~text
Use WSL 2 instead of Hyper-V
~~~

启动 Docker Desktop，等待底部显示 Engine running。

验证：

~~~powershell
docker version
docker info
docker compose version
docker run --rm hello-world
~~~

如果出现 npipe 或 daemon 连接失败，说明 Docker Engine 还没有启动。

如果 Docker Hub 下载超时，需要在 Docker Desktop 的 Settings → Resources → Proxies 中配置可信代理，或使用组织提供的可信 registry mirror。

## 四、配置 Windows hosts

以管理员身份打开记事本：

~~~powershell
notepad C:\Windows\System32\drivers\etc\hosts
~~~

加入：

~~~text
127.0.0.1 server-00.oauth.labs
127.0.0.1 client-00.oauth.labs
127.0.0.1 server-01.oauth.labs
127.0.0.1 client-01.oauth.labs
~~~

刷新并验证：

~~~powershell
ipconfig /flushdns
Resolve-DnsName server-00.oauth.labs
Resolve-DnsName client-00.oauth.labs
Resolve-DnsName server-01.oauth.labs
Resolve-DnsName client-01.oauth.labs
~~~

四个域名都应解析为 127.0.0.1。

## 五、准备仓库

~~~powershell
cd "E:\homework\computer security\project\OAuth"
git config core.autocrlf true
~~~

仓库中的 .gitattributes 会把 shell 脚本保持为 LF。

检查入口脚本：

~~~powershell
git check-attr text eol -- docker/entrypoint.sh
git ls-files --eol docker/entrypoint.sh
~~~

工作区应显示 w/lf 或 w/mixed，不能使用 CRLF。

如果曾出现 CRLF，执行：

~~~powershell
git add --renormalize docker/entrypoint.sh .gitattributes
git checkout -- docker/entrypoint.sh
~~~

## 六、生成实验配置

首次运行或配置丢失时执行：

~~~powershell
make config
~~~

如果没有安装 GNU Make，可以构建配置生成器：

~~~powershell
cd configure
go build -o bin/configure.exe ./cmd/configure
.\bin\configure.exe
cd ..
~~~

该步骤会生成：

- docker/lab00 和 docker/lab01 下的本地配置。
- 数据库账号和密码。
- 本地 RSA 密钥。
- docker/db/init.prod.sql。

这些文件已被 Git 忽略，不能提交到仓库。

## 七、启动 lab01

~~~powershell
docker compose down
docker compose up -d --build server-01 client-01
docker compose ps
~~~

应当运行以下容器：

- oauth-db-1
- oauth-valkey-1
- oauth-caddy-1
- oauth-server-01-1
- oauth-client-01-1

server-01 和 client-01 不能显示为 Exited。

## 八、检查健康状态

~~~powershell
curl.exe -k -i https://server-01.oauth.labs/health
curl.exe -k -i https://client-01.oauth.labs/health
~~~

两个接口都应返回 200 和健康状态。

运行 smoke test：

~~~powershell
python scripts\smoke.py --lab 01
~~~

预期输出包含：

~~~text
lab01: healthy, state and S256 PKCE present, invalid state rejected
~~~

## 九、手动运行两个版本

### vulnerable

~~~powershell
docker compose down -v
$env:LAB_VARIANT="vulnerable"
docker compose up -d --build server-01 client-01
python lab01\reproduce.py --variant vulnerable
Remove-Item Env:LAB_VARIANT
~~~

预期：

~~~json
{"exploitable": true, "impact_verified": true}
~~~

### fixed

~~~powershell
docker compose down -v
$env:LAB_VARIANT="fixed"
docker compose up -d --build server-01 client-01
python lab01\reproduce.py --variant fixed
Remove-Item Env:LAB_VARIANT
~~~

预期：

~~~json
{"exploitable": false, "impact_verified": false}
~~~

## 十、运行完整自动实验

~~~powershell
python scripts\run_experiment.py --lab 01
~~~

预期：

~~~text
lab01: healthy
lab01 vulnerable: expected outcome confirmed
lab01: healthy, state and S256 PKCE present, invalid state rejected
lab01 fixed: expected outcome confirmed
~~~

该脚本会自动：

- 启动和删除实验容器。
- 分别运行 vulnerable 和 fixed。
- 删除数据库卷，保证实验环境干净。
- 验证两个版本的预期结果。

## 十一、停止和清理

只停止容器，保留数据库数据：

~~~powershell
docker compose stop
~~~

停止并删除容器与网络，保留数据卷：

~~~powershell
docker compose down
~~~

完全重置实验室，包括数据库卷：

~~~powershell
docker compose down -v
~~~

完全重置后需要重新运行：

~~~powershell
make config
docker compose up -d --build server-01 client-01
~~~

## 十二、常见错误

### Docker daemon 未运行

错误：

~~~text
failed to connect to the docker API at npipe:////./pipe/docker_engine
~~~

处理：

1. 启动 Docker Desktop。
2. 等待 Engine running。
3. 再执行 docker info。

### 域名无法解析

错误：

~~~text
getaddrinfo failed
~~~

处理：检查 Windows hosts 中四个 OAuth 域名。

### Caddy 返回 502

错误：

~~~text
HTTP/1.1 502 Bad Gateway
~~~

处理：

~~~powershell
docker compose ps
docker compose logs --no-color --tail=200 server-01 client-01
~~~

如果日志包含 exec /docker-entrypoint.sh: no such file or directory，说明入口脚本被转成了 CRLF。应确认 .gitattributes 存在，并重新规范化脚本后重建镜像。

### Docker Hub 拉取镜像超时

处理：

- 检查 Docker Desktop 代理。
- 执行 docker login 登录自己的 Docker 账号。
- 使用明确可信的 registry mirror。
- 不要随意使用来源不明的第三方镜像站。

## 十三、禁止提交的文件

以下内容属于本地实验环境，不应提交：

- docker/lab00/*.config.yaml
- docker/lab01/*.config.yaml
- docker/db/*.sql
- docker/db/root_password.prod.txt
- 私钥、数据库密码和会话密钥
- 运行时容器数据和日志
