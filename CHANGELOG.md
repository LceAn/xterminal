# 更新日志

本文件记录 XTerminal 的重要变更，格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [未发布] - 2026-08-31

### 新增
- 新增 `--token` 访问令牌：页面与 `/api` 可要求令牌（`?token=` 或 `Authorization: Bearer`），校验失败返回 404 并记录日志，令牌比较使用 `crypto/subtle` 常数时间算法；前端在首次带 `?token=` 访问后通过 localStorage 自动附带。
- 以非回环地址监听且未设置令牌时，启动日志输出警告。

### 测试
- 新增 `TestTokenAuthMiddleware`（7 个子用例）与默认开放行为回归测试，在 Linux 环境（go1.25.13）全部通过。

## [未发布] - 2026-08-31

### 安全
- Go 工具链基线从 1.22 升级到 1.25.13，规避 1.22 系列标准库已知漏洞（GO-2026-6218/6091/6090/6089 涉及 net/url、html/template、crypto/tls、net/http）；README 编译要求同步更新。

## [未发布] - 2026-08-30

### 新增
- 新增 `CHANGELOG.md` 与 `ROADMAP.md`。
- README 增加文档索引。

### 说明
- 本次为文档整理，服务端代码未改动。

## 历史概要

- 2026-08-04 `fix: secure monitor listener and HTTP handling` — 监听地址与 HTTP 处理加固（默认仅 127.0.0.1、请求超时、404/405 语义）。
- 2026-07-28 `docs: standardize README structure` — 标准化 README 结构。
- 2026-04-03 `fix: 修复 PortService 结构体字段名错误`。
