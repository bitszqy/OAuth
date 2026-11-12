# 字段级 OAuth 日历原型架构

## 一、系统目标

系统保留现有 OAuth 2.0 粗粒度授权，在资源数据上增加字段级加密和任务级密钥释放。应用即使拿到合法的 read:calendar token，也只能解密当前任务批准的字段。

## 二、组件分工

| 组件 | 职责 |
|---|---|
| lab00 | 原生 OAuth 对照基线，保留粗粒度 read:profile |
| lab01/server | OAuth 授权、日历资源、任务 grant、字段加密和密钥释放 |
| lab01/client | 请求日历数据、申请任务密钥并只解密允许字段 |
| MariaDB | 保存用户、客户端、consent、任务 grant、token 和撤销状态 |
| Valkey | 保存会话和授权码 |
| Caddy | 为本地实验域名提供 HTTPS 反向代理 |

OAuth 授权服务器和资源服务器当前部署在同一个可信 server 容器中；第三方 client 是独立进程。这个边界足以验证“应用拿到 token 后不能读取多余明文”。

## 三、任务目录

任务目录由服务端控制，当前定义在：

~~~text
lab01/server/internal/taskpolicy/catalog.go
~~~

| 任务 | 允许字段 |
|---|---|
| calendar.freebusy | start、end |
| calendar.reminder | start、end、title |

客户端不能请求额外字段，也不能自行把字段加入任务。

## 四、OAuth 与数据访问流程

1. 客户端发起 OAuth 授权，并附带 task_id=calendar.freebusy。
2. 授权服务器展示客户端、任务名称和允许字段。
3. 用户批准客户端和当前任务。
4. 授权码换取 access token，token 包含粗粒度 read:calendar scope。
5. 客户端请求 /api/calendar/events。
6. 脆弱版本返回完整日历明文。
7. 固定版本返回逐字段 AES-GCM 密文。
8. 客户端把 task_id 和 resource_id 提交到 /api/task-keys。
9. 密钥服务校验 token、scope、任务 grant 和资源归属。
10. 密钥服务只释放当前任务允许字段的密钥。
11. 客户端解密允许字段，其他字段保持 locked。

## 五、当前已经实现的机制

- Authorization Code Flow。
- state 和 S256 PKCE。
- redirect URI、scope、issuer、audience、client_id 和 token 撤销校验。
- 客户端级 consent。
- 服务端控制的任务级 consent。
- user_task_grants 任务授权表。
- 按 user、resource、attribute 和版本派生 HKDF 字段密钥。
- AES-256-GCM 字段加密。
- AAD 绑定用户、资源、字段和策略版本。
- 密钥接口的资源和任务授权检查。
- vulnerable/fixed 双变体。
- 单元测试和自动复现脚本。

## 六、数据流和信任边界

OAuth AS 与资源服务器属于可信区。第三方应用属于不可信区。第三方应用可以：

- 获得合法 access token。
- 获取完整密文。
- 修改 task_id 和 resource_id。
- 请求任务密钥。

第三方应用不能直接获得可信区中的 master key 或未批准字段的字段密钥。

当前尚未解决的边界是：access token 与客户端实例之间还没有强密码学绑定。后续需要 DPoP、客户端签名或 mTLS。

## 七、vulnerable 与 fixed 的差异

### vulnerable

- 日历接口直接返回完整明文。
- 任务密钥接口可以返回全部字段密钥。
- 用于展示原生 OAuth scope 过粗造成的过度访问。

### fixed

- 日历接口返回逐字段密文。
- 密钥接口要求 active task grant。
- calendar.freebusy 只释放 start 和 end。
- 标题、地点、参会人和描述保持锁定。

## 八、当前原型的明确限制

- 尚未使用客户端公钥封装字段密钥。
- 尚未实现短期 capability 和 jti 防重放。
- 尚未实现 key epoch、轮换和单任务撤销。
- 尚未实现结构化审计日志。
- 尚未完成 Docker 端到端实测。
- 尚未采集性能和安全效果数据。

## 九、下一阶段实现顺序

1. 在 Docker 环境完成端到端 OAuth 和两变体实验。
2. 加入 DPoP 或客户端签名请求。
3. 将字段密钥封装到客户端公钥。
4. 增加短期 capability、jti 和重放测试。
5. 增加 key epoch、轮换、单任务撤销和审计。
6. 完成攻击测试矩阵。
7. 完成性能和功能基准测试。
8. 整理最终报告、演示 PPT 和现场 Demo。
