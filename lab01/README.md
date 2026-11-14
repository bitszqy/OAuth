# Lab 01：日历任务数据过度访问

lab00 保留为原生 OAuth 基线。lab01 用来验证：一个拥有合法 read:calendar token 的第三方应用，能否读取用户当前批准任务之外的字段。

## 一、任务策略

- calendar.freebusy：允许 start、end
- calendar.reminder：允许 start、end、title

日历资源中仍然存在 location、attendees 和 description，但 freebusy 任务不能解密这些字段。

## 二、两种版本

### vulnerable

- 资源服务器返回完整日历明文。
- 任务 grant 不限制 JSON 中可见的字段。
- 用于展示原生 OAuth 粗粒度授权造成的数据过度访问。

### fixed

- 每个日历字段单独加密。
- 密钥接口只返回当前任务允许字段的密钥。
- freebusy 任务只能解密 start 和 end。

## 三、运行方式

从仓库根目录运行：

~~~sh
python3 scripts/run_experiment.py --lab 01
~~~

自动复现脚本会执行：

1. 注册并登录本地用户。
2. 启动 OAuth Authorization Code Flow。
3. 批准 calendar.freebusy 任务。
4. 请求日历资源和任务密钥。
5. 检查应用是否能看到事件标题和描述。

预期结果：

- vulnerable：能够看到标题和描述，exploitable=true。
- fixed：看不到标题和描述，exploitable=false。

## 四、当前状态

Docker 端到端实验已经跑通：

~~~text
lab01: healthy
lab01 vulnerable: expected outcome confirmed
lab01: healthy, state and S256 PKCE present, invalid state rejected
lab01 fixed: expected outcome confirmed
~~~

Windows 环境搭建和常见错误处理见 [Windows 环境搭建说明](../docs/windows-setup.md)。

## 五、尚未完成

- DPoP 或客户端签名证明。
- 客户端公钥封装字段密钥。
- capability 过期和重放保护。
- key epoch 与密钥轮换。
- 单任务撤销。
- 审计日志。
- 性能与功能基准测试。
