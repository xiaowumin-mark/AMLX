# 用户系统接口文档

基础路径：`/api/v1`

本文件覆盖用户、认证、权限相关接口。

## 约定

- 所有请求/响应均为 JSON。
- `Content-Type: application/json`
- 时间格式为 RFC3339（Go `time.Time` 默认 JSON 格式）。
- 错误响应结构：
```json
{"error":"message"}
```

## 鉴权说明

- Access Token：放在 `Authorization: Bearer <access_token>`。
- Refresh Token：在请求体 JSON 中传递。
- Token 对示例：
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "access_expires_at": "2026-02-08T10:00:00Z",
  "refresh_expires_at": "2026-02-15T10:00:00Z"
}
```

## 认证接口

### 注册

- `POST /auth/register`
- 是否需要登录：否
- 请求：
```json
{
  "name": "alice",
  "email": "alice@example.com",
  "password": "your-password",
  "role_id": 1
}
```
- 响应 `201`：
```json
{
  "user": {
    "id": 1,
    "name": "alice",
    "email": "alice@example.com",
    "role_id": 1,
    "ban": false,
    "created_at": "2026-02-08T10:00:00Z",
    "updated_at": "2026-02-08T10:00:00Z"
  },
  "tokens": {
    "access_token": "...",
    "refresh_token": "...",
    "access_expires_at": "2026-02-08T10:00:00Z",
    "refresh_expires_at": "2026-02-15T10:00:00Z"
  }
}
```

说明：
- 当 `auth.allow_register_role=false` 时，会忽略 `role_id`，使用 `auth.default_role_id`。

### 登录

- `POST /auth/login`
- 是否需要登录：否
- 请求：
```json
{
  "email": "alice@example.com",
  "password": "your-password"
}
```
- 响应 `200`：与注册相同结构。

### 刷新 Token

- `POST /auth/refresh`
- 是否需要登录：否
- 请求：
```json
{"refresh_token":"..."}
```
- 响应 `200`：
```json
{"tokens":{...}}
```

说明：
- 默认启用 refresh token 轮换（见 `auth.refresh_token_reuse`）。

### 登出

- `POST /auth/logout`
- 是否需要登录：否（只要提供 refresh token 即可）
- 请求：
```json
{"refresh_token":"..."}
```
- 响应 `200`：
```json
{"ok":true}
```

### 当前用户

- `GET /auth/me`
- 是否需要登录：是（access token）
- 响应 `200`：
```json
{
  "user": {
    "id": 1,
    "name": "alice",
    "email": "alice@example.com",
    "role_id": 1,
    "ban": false,
    "created_at": "2026-02-08T10:00:00Z",
    "updated_at": "2026-02-08T10:00:00Z"
  }
}
```

### 全部登出

- `POST /auth/logout_all`
- 是否需要登录：是（access token）
- 响应 `200`：
```json
{"ok":true}
```

### 修改密码

- `POST /auth/change_password`
- 是否需要登录：是（access token）
- 请求：
```json
{
  "old_password": "old-pass",
  "new_password": "new-pass"
}
```
- 响应 `200`：
```json
{"ok":true}
```

## 用户接口（仅管理员）

所有用户接口要求：
- 有效的 access token
- 具备 `admin` 权限

### 创建用户

- `POST /users`
- 请求：
```json
{
  "name": "bob",
  "email": "bob@example.com",
  "password": "your-password",
  "role_id": 2
}
```
- 响应 `201`：`user` 对象（同上）。

### 通过 ID 查询用户

- `GET /users/:id`
- 响应 `200`：`user` 对象。

### 通过邮箱查询用户

- `GET /users?email=alice@example.com`
- 响应 `200`：`user` 对象。

### 更新用户

- `PATCH /users/:id`
- 请求：
```json
{
  "name": "new-name",
  "email": "new-email@example.com",
  "password": "new-pass",
  "role_id": 3
}
```
- 响应 `200`：`user` 对象。

### 封禁/解封用户

- `PUT /users/:id/ban`
- 请求：
```json
{"ban":true}
```
- 响应 `200`：
```json
{"ok":true}
```

## 权限接口（仅管理员）

所有权限接口要求：
- 有效的 access token
- 具备 `admin` 权限

### 创建角色

- `POST /roles`
- 请求：
```json
{"name":"editor","description":"Content editor"}
```
- 响应 `201`：
```json
{"role":{"id":1,"name":"editor","description":"Content editor"}}
```

### 创建权限

- `POST /permissions`
- 请求：
```json
{"name":"post.write","description":"Write posts"}
```
- 响应 `201`：
```json
{"permission":{"id":1,"name":"post.write","description":"Write posts"}}
```

### 查看角色权限

- `GET /roles/:id/permissions`
- 响应 `200`：
```json
{"permissions":[{"id":1,"name":"post.write","description":"Write posts"}]}
```

### 给角色添加权限

- `POST /roles/:id/permissions`
- 请求：
```json
{"permission_id":1}
```
- 响应 `200`：
```json
{"ok":true}
```

### 移除角色权限

- `DELETE /roles/:id/permissions/:perm_id`
- 响应 `200`：
```json
{"ok":true}
```

## 常见状态码

- `200` OK
- `201` Created
- `400` 参数/JSON 错误
- `401` 未授权（缺少/无效 token）
- `403` 无权限（权限不足或注册关闭）
- `404` 未找到
- `409` 冲突（邮箱/角色/权限已存在）
- `500` 服务端错误
