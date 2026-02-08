# 配置文件

---

## 配置规范

- `mysql.host` 主机地址
- `mysql.port` 端口
- `mysql.user` 用户
- `mysql.password` 密码
- `mysql.database` 数据库名
- `mysql.charset` 字符集
- `mysql.parse_time` 解析时间字段
- `mysql.loc` 时区
- `mysql.max_open_conns` 最大连接数
- `mysql.max_idle_conns` 最大空闲连接数
- `mysql.conn_max_lifetime` 连接最大生命周期
- `mysql.conn_max_idle_time` 空闲连接最大生命周期
- `mysql.log_level` GORM 日志级别
- `mysql.slow_threshold_ms` 慢查询阈值
- `server.port` 服务端口
- `server.log` 是否输出日志
- `server.log_level` 日志级别
- `server.read_timeout` 读取超时
- `server.write_timeout` 写入超时
- `server.idle_timeout` 空闲超时
## Log Config

- `log.level` log level (debug/info/warn/error)
- `log.format` output format (text/json)
- `log.output` output target (stdout/stderr/file/both/discard)
- `log.file` log file path (required when output=file/both)
- `log.add_source` include source location
- `log.time_format` time format (default RFC3339)

## Auth Config

- `auth.jwt_secret` JWT signing secret (required)
- `auth.issuer` JWT issuer
- `auth.access_ttl` access token ttl (e.g. 15m)
- `auth.refresh_ttl` refresh token ttl (e.g. 168h)
- `auth.bcrypt_cost` bcrypt cost (default 10)
- `auth.allow_register` allow user registration
- `auth.allow_register_role` allow setting role_id on register
- `auth.default_role_id` default role id for register
- `auth.refresh_token_reuse` allow refresh token reuse (false = rotate)
- `auth.bootstrap_admin_role` ensure admin role + permission on startup
