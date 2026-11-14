# OAuth 数据过度访问项目全链条

## 一、项目最终目标

完成一个可运行、可验证、可复现的 OAuth 字段级访问控制原型，并回答：

> 第三方应用即使拥有合法的 OAuth access token，能否只读取用户批准任务真正需要的数据字段？

最终交付不能只有代码，还需要威胁模型、安全测试、性能数据、项目报告和演示材料。

## 二、当前已经完成的事情

### 1. 本地仓库和分支

- 仓库位置：E:\homework\computer security\project\OAuth
- 当前分支：feature/field-level-oauth
- lab00 保留为原生 OAuth 对照基线。
- lab01 用于实现字段级访问控制实验。

### 2. OAuth 基础平台

当前已经具备：

- OAuth 2.0 Authorization Code Flow。
- state 和 S256 PKCE。
- redirect URI、scope、client_id、issuer、audience 和 JWT 校验。
- access token 和 refresh token。
- token 撤销。
- 客户端 consent。
- Docker、MariaDB、Valkey 和 Caddy 配置文件。

这些能力保证调用者拿到合法 token，但不能单独解决授权后的字段过度访问。

### 3. 日历业务场景

lab01 已经加入：

- read:calendar 粗粒度 scope。
- /api/calendar/events 日历资源接口。
- start、end、title、location、attendees、description 字段。
- calendar.freebusy 任务，只允许 start 和 end。
- calendar.reminder 任务，允许 start、end 和 title。
- 服务端任务目录。

### 4. 任务级用户授权

已经实现：

- user_task_grants 数据表。
- consent 页面显示任务名称和具体字段。
- 授权记录绑定 user_id、client_id 和 task_id。
- 撤销 consent 时同步撤销任务 grant 和 token。

### 5. 字段级加密

已经实现：

- 使用 HKDF 按 user、resource、attribute 和策略版本派生字段密钥。
- 使用 AES-256-GCM 加密单个字段。
- 每个字段使用随机 nonce。
- AAD 绑定用户、资源、字段和策略版本。
- 交换字段或用户的密文会导致解密失败。

### 6. 密钥释放

已经实现：

- /api/task-keys 接口。
- token 和 read:calendar scope 校验。
- 当前任务的 active user grant 校验。
- 固定版本只释放任务允许字段的密钥。
- 脆弱版本释放全部字段密钥，用于安全对照。
- 资源 ID 必须属于 token 中的用户。

### 7. 客户端

已经实现：

- 客户端请求日历资源。
- 固定版本获取任务密钥并只解密允许字段。
- 未授权字段显示为 locked。
- 脆弱版本展示完整明文。
- 默认使用 calendar.freebusy 任务。

### 8. 自动验证

已经完成：

- lab01/reproduce.py 自动 OAuth 复现脚本。
- 字段加密和 AAD 绑定单元测试。
- 任务混淆测试。
- 跨用户密钥测试。
- Python 脚手架测试。
- 修复 new_lab.py 插入 Compose 服务位置错误的问题。
- Go 单元测试通过。
- Go vet 通过。
- 服务端和客户端二进制构建通过。
- Python 8 个测试通过。

### 9. Docker 端到端验证

Docker 端到端实验已经完成，自动输出：

- lab01: healthy
- lab01 vulnerable: expected outcome confirmed
- lab01: healthy, state and S256 PKCE present, invalid state rejected
- lab01 fixed: expected outcome confirmed

这说明 OAuth、consent、任务 grant、字段加密、密钥释放和 vulnerable/fixed 自动对照已经形成完整闭环。

## 三、仍未完成的安全与评价工作

### 1. 客户端证明还没有实现

目前 token 与客户端实例之间还没有强密码学绑定。下一步需要加入以下一种机制：

- DPoP。
- 客户端签名请求。
- 客户端公私钥绑定。
- mTLS。

### 2. 密钥还没有封装给客户端公钥

现在字段密钥通过 TLS 直接返回。最终方案应使用客户端公钥封装字段密钥，避免 bearer token 本身成为获取密钥的充分条件。

### 3. Capability 和密钥生命周期还没有完成

还需要：

- 短期 capability。
- capability 的 jti、audience 和过期时间。
- 重放保护。
- key epoch 和密钥轮换。
- 单任务撤销。
- 撤销后禁止解密未来数据。

### 4. 攻击测试还不够完整

还需要测试：

- 伪造 capability。
- 使用过期 capability。
- 重放 capability。
- 使用另一个客户端的 capability。
- 使用另一个用户的 token。
- 替换 resource_id。
- 撤销后再次申请密钥。
- 只有 token、没有客户端私钥。

### 5. 实验数据还没有测量

还需要测量：

- 可读非必要字段数。
- 数据暴露减少比例。
- 功能成功率。
- p50 和 p95 请求延迟。
- 加密、解密和密钥申请耗时。
- 响应字节数和密文膨胀比例。
- 不同事件数量下的扩展性。

### 6. 最终交付物还没有完成

还需要：

- 最终技术报告。
- 威胁模型和架构图。
- 安全测试结果表。
- 性能结果图。
- 8 至 12 分钟演示 PPT。
- 现场 Demo 和失败备用录屏。
- 完整复现实验说明。

## 四、整个项目链条

### 阶段 1：固定研究问题

需要产出：

- 一句话研究问题。
- 保护资产。
- 攻击者类型。
- 攻击能力。
- 信任边界。
- 安全目标。
- 明确的范围外风险。

完成标准：能够清楚回答系统防谁、防什么、不防什么。

### 阶段 2：建立 OAuth 基线

需要产出：

- lab00 正常运行。
- 合法 token 可以读取完整日历明文。
- 记录 token、scope、consent 和资源访问路径。

完成标准：稳定复现“合法授权仍然导致字段过度暴露”。

### 阶段 3：实现任务策略

需要产出：

- 服务端任务目录。
- 用户任务授权。
- consent 页面。
- 任务和字段的交集计算。

完成标准：应用不能自行声明任务需要全部字段；用户必须明确批准任务。

### 阶段 4：实现字段保护

需要产出：

- 字段加密。
- 字段密钥派生。
- 密钥释放接口。
- 客户端解密。

完成标准：freebusy 任务只能解密 start 和 end，其他字段保持锁定。

### 阶段 5：补齐安全边界

需要产出：

- 客户端 PoP 或 DPoP。
- 公钥封装密钥。
- 短期 capability。
- 重放保护。
- key epoch 和撤销。
- 审计日志。

完成标准：只有 token 不足以获得密钥，越权、重放和撤销测试全部通过。

### 阶段 6：完成评价

至少设置三个对照组：

1. 原生 OAuth：完整明文。
2. 服务器字段过滤：只返回必要明文。
3. 提出方案：返回密文，只释放必要字段密钥。

需要输出：

- 数据暴露减少率。
- 功能成功率。
- 延迟和通信开销。
- 越权测试结果。

完成标准：能够定量说明安全收益和性能成本，而不是只展示功能。

### 阶段 7：复现与交付

需要产出：

- Docker 一条命令启动。
- make experiment LAB=01 自动验证。
- 实验数据可以重新生成。
- 最终报告和演示材料。

完成标准：其他人按照 README 能从头复现主要结果。

## 五、建议的下一步顺序

1. 在安装 Docker Desktop 的机器上完成 lab01 端到端联调。
2. 运行 reproduce.py，确认脆弱版本暴露标题、修复版本只显示开始和结束时间。
3. 加入客户端 PoP。
4. 封装字段密钥并增加短期 capability。
5. 增加 key epoch、单任务撤销和审计日志。
6. 补齐攻击测试矩阵。
7. 编写 benchmark 并生成真实数据。
8. 根据数据写报告，最后制作 PPT 和现场 Demo。

## 六、验收总表

| 验收项 | 当前状态 | 完成标准 |
|---|---|---|
| OAuth 基线 | 代码完成，待 Docker 联调 | 能复现完整明文过度访问 |
| 任务目录 | 已实现 | 客户端不能自行扩字段 |
| 任务 consent | 代码完成，待联调 | 页面展示并保存任务 grant |
| 字段加密 | 已实现并单测 | 密文不能跨字段使用 |
| 密钥释放 | 已实现并单测 | 只释放任务允许字段 |
| 客户端 PoP | 未实现 | 只有 token 不能取密钥 |
| Capability | 未实现 | 短期、绑定、不可重放 |
| 密钥轮换 | 未实现 | 撤销后未来数据无法解密 |
| 越权测试 | 部分实现 | 所有越权请求应拒绝 |
| 性能实验 | 未实现 | 有 p50、p95 和开销数据 |
| 自动复现 | 脚本完成，待运行 | vulnerable/fixed 均符合预期 |
| 最终报告 | 未完成 | 研究问题、设计和数据一致 |

## 七、需要你本人确认或操作的事项

- 提供项目截止时间和最终提交要求。
- 在运行 Docker 的电脑上启动 Docker Desktop。
- 配置本地 hosts 域名。
- 决定团队是否继续统一使用 Go。
- 明确老师是否允许本地模拟 OAuth 环境。
- 确认最终是提交报告、PPT、代码、现场演示，还是全部都需要。
