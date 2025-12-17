# AMLX

> **Advanced Music Lyrics eXtended Platform**\
> 一个面向 TTML 歌词的全生命周期平台：管理、投稿、统计、搜索、分发。

---

## ✨ 项目简介

**AMLX** 是围绕 [amll-ttml-db](https://github.com/Steve-xmh/amll-ttml-db) 构建的下一代歌词基础设施平台。

它并不是简单的“歌词网站”，而是一个 **歌词的控制平面（Control Plane）**，负责：

- 歌词数据的统一管理
- 可视化投稿与审核
- TTML 的结构化解析与版本控制
- 数据统计、搜索与分析
- 向歌词 CDN（AMLX-CDN）下发控制指令

> **设计目标**：
>
> - 对贡献者友好
> - 对客户端高效
> - 对维护者可控
> - 对未来可扩展

---

## 🧩 核心特性

### 1️⃣ TTML 结构化管理

- TTML → AST（抽象语法树）
- 统一的歌词中间表示
- 支持版本 Diff / Patch

### 2️⃣ 可视化投稿流程

- 无需 Git / PR 经验
- 时间轴拖拽编辑
- 实时歌词预览（AMLL 风格）
- 自动规范校验

### 3️⃣ 自动 GitHub 同步

- 投稿即生成 PR
- 双向同步 amll-ttml-db
- 保留 GitHub 作为权威存档

### 4️⃣ 歌词搜索与统计

- 按歌名 / 歌词内容 / 语言搜索
- 投稿者贡献统计
- 歌词请求与使用分析

### 5️⃣ CDN 控制与调度

- 向 AMLX-CDN 发布歌词版本
- 控制分发格式（Binary / TTML / Diff）
- 监测 CDN 节点状态

---

## 🏗️ 系统架构

```
Client / Player
      │
      ▼
  AMLX-CDN  ◄──────►  AMLX (Control Plane)
      │                    │
      ▼                    ▼
 Binary Lyrics         GitHub Repo
```

- **AMLX**：控制面 & 管理平台
- **AMLX-CDN**：数据面 & 分发节点

---

## 📦 仓库职责

| 仓库         | 职责                |
| ---------- | ----------------- |
| `AMLX`     | 管理 / 投稿 / 搜索 / 控制 |
| `AMLX-CDN` | 歌词缓存 / 分发 / 加速    |

---

## 🛠 技术栈（规划）

- Backend：Go
- Lyrics Codec：自研 TTML AST / Binary Pack
- Frontend：Vue
- Search：Bleve
- Sync：GitHub API

---

## 🚧 当前状态

> **WIP / Early Design Stage**

-

---

## 🤝 参与贡献

欢迎：

- TTML 规范讨论
- Codec / Binary 格式设计
- CDN 架构建议

> 本项目目标是：**让高质量歌词更容易被创作、分发与使用**。

---
