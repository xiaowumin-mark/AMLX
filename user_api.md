# User System API

Base path: `/api/v1`

This document covers user, auth, and permission APIs implemented in this project.

## Conventions

- All request/response bodies are JSON.
- `Content-Type: application/json`
- Times are RFC3339 (Go `time.Time` default JSON format).
- Error response shape:
```json
{"error":"message"}
```

## Authentication

- Access token: JWT in `Authorization: Bearer <access_token>`.
- Refresh token: sent in JSON body.
- Token pair response:
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "access_expires_at": "2026-02-08T10:00:00Z",
  "refresh_expires_at": "2026-02-15T10:00:00Z"
}
```

## Auth Endpoints

### Register

- `POST /auth/register`
- Auth: public
- Request:
```json
{
  "name": "alice",
  "email": "alice@example.com",
  "password": "your-password",
  "role_id": 1
}
```
- Response `201`:
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

Notes:
- If `auth.allow_register_role=false`, `role_id` will be ignored and `auth.default_role_id` is used.

### Login

- `POST /auth/login`
- Auth: public
- Request:
```json
{
  "email": "alice@example.com",
  "password": "your-password"
}
```
- Response `200`: same shape as Register.

### Refresh

- `POST /auth/refresh`
- Auth: public
- Request:
```json
{"refresh_token":"..."}
```
- Response `200`:
```json
{"tokens":{...}}
```

Notes:
- By default refresh tokens are rotated (see `auth.refresh_token_reuse`).

### Logout

- `POST /auth/logout`
- Auth: public
- Request:
```json
{"refresh_token":"..."}
```
- Response `200`:
```json
{"ok":true}
```

### Me

- `GET /auth/me`
- Auth: access token required
- Response `200`:
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

### Logout All

- `POST /auth/logout_all`
- Auth: access token required
- Response `200`:
```json
{"ok":true}
```

### Change Password

- `POST /auth/change_password`
- Auth: access token required
- Request:
```json
{
  "old_password": "old-pass",
  "new_password": "new-pass"
}
```
- Response `200`:
```json
{"ok":true}
```

## User Endpoints (Admin Only)

All user endpoints require:
- Valid access token
- Permission `admin`

### Create User

- `POST /users`
- Request:
```json
{
  "name": "bob",
  "email": "bob@example.com",
  "password": "your-password",
  "role_id": 2
}
```
- Response `201`: `user` object (same as above).

### Get User By ID

- `GET /users/:id`
- Response `200`: `user` object.

### Get User By Email

- `GET /users?email=alice@example.com`
- Response `200`: `user` object.

### Update User

- `PATCH /users/:id`
- Request:
```json
{
  "name": "new-name",
  "email": "new-email@example.com",
  "password": "new-pass",
  "role_id": 3
}
```
- Response `200`: `user` object.

### Set Ban

- `PUT /users/:id/ban`
- Request:
```json
{"ban":true}
```
- Response `200`:
```json
{"ok":true}
```

## Permission Endpoints (Admin Only)

All permission endpoints require:
- Valid access token
- Permission `admin`

### Create Role

- `POST /roles`
- Request:
```json
{"name":"editor","description":"Content editor"}
```
- Response `201`:
```json
{"role":{"id":1,"name":"editor","description":"Content editor"}}
```

### Create Permission

- `POST /permissions`
- Request:
```json
{"name":"post.write","description":"Write posts"}
```
- Response `201`:
```json
{"permission":{"id":1,"name":"post.write","description":"Write posts"}}
```

### List Role Permissions

- `GET /roles/:id/permissions`
- Response `200`:
```json
{"permissions":[{"id":1,"name":"post.write","description":"Write posts"}]}
```

### Add Permission To Role

- `POST /roles/:id/permissions`
- Request:
```json
{"permission_id":1}
```
- Response `200`:
```json
{"ok":true}
```

### Remove Permission From Role

- `DELETE /roles/:id/permissions/:perm_id`
- Response `200`:
```json
{"ok":true}
```

## Common Status Codes

- `200` OK
- `201` Created
- `400` Invalid input or JSON
- `401` Unauthorized (missing/invalid token)
- `403` Forbidden (permission denied or registration disabled)
- `404` Not found
- `409` Conflict (email/role/permission exists)
- `500` Internal server error
