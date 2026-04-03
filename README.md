# XTerminal - Linux 系统状态监控面板

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-yellow.svg)](https://www.linux.org/)

轻量级 Linux 系统实时监控 Web 面板，Go 单文件实现，内嵌前端，零依赖部署。

## ✨ 功能特性

### 📊 系统信息
- CPU 型号、核心数、负载、温度
- 内存使用率、Swap 状态
- 磁盘分区信息、总容量统计
- 系统运行时间

### 🌐 网络监控
- 网卡状态、IPv4/IPv6 地址
- 实时流量统计（RX/TX）
- TCP 连接数、已建立连接、TIME-WAIT

### 📈 进程管理
- Top 10 CPU 占用进程
- 显示 PID、用户、CPU%、MEM%、命令

### 🔌 端口服务
- 监听端口列表
- 服务类型识别（system/docker/go/python/node/java 等）
- 重要端口标记（22, 80, 443, 3306, 5432, 6379, 27017）
- 进程 PID 关联

### 🔧 开发环境
自动检测以下工具版本：
- **编译工具**: GCC, Make, CMake
- **编程语言**: Python, pip, Node.js, npm, Go, Java
- **数据库**: PostgreSQL, MariaDB, Redis
- **容器**: Docker, Docker Compose
- **常用工具**: Git, Vim, tmux, htop, jq, rg, bat, fd

### ⚙️ 服务状态
- Docker, Nginx, Redis, PostgreSQL, MariaDB, SSH 运行状态

### 📋 系统日志
- 最近 8 条 journalctl 日志
- 实时展示服务名、时间、消息

### 🎨 用户界面
- 🌙/☀️ **昼夜模式切换** - 自动跟随系统，可手动切换
- 📱 **移动端适配** - 响应式布局，手机也能用
- ⚡ **实时刷新** - 每 3 秒自动更新数据

## 📸 截图

| 深色模式 | 浅色模式 |
|:---:|:---:|
| ![深色模式](preview.png) | ![浅色模式](preview-light.png) |

## 🚀 快速部署

### 编译运行

```bash
# 克隆项目
git clone https://github.com/LceAn/xterminal.git
cd xterminal

# 编译（需要 Go 1.18+）
go build -o server_monitor server_monitor.go

# 运行（默认监听 :8080）
./server_monitor

# 指定端口运行
./server_monitor --port 9000
```

### 访问面板

```
http://your-server:8080
```

### systemd 服务（推荐）

创建服务文件 `/etc/systemd/system/server-monitor.service`:

```ini
[Unit]
Description=Server Monitor Panel
After=network.target

[Service]
Type=simple
ExecStart=/path/to/server_monitor
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
sudo systemctl enable server-monitor
sudo systemctl start server-monitor
```

## 🔌 API 接口

| 接口 | 说明 |
|:---|:---|
| `GET /` | 监控面板 HTML 页面 |
| `GET /api` | 系统信息 JSON 接口 |

### JSON 返回示例

```json
{
  "cpu": {
    "model": "AMD EPYC 7002",
    "cores": 8,
    "load_avg": "0.15 0.10 0.05",
    "temp": 45.0
  },
  "memory": {
    "total_mb": 16384,
    "used_mb": 2048,
    "percent": 12.5
  },
  "disk_total": {
    "total_gb": 500,
    "used_gb": 150,
    "percent": "30%"
  },
  "uptime": "15d 8h 32m",
  "timestamp": "2024-01-15 14:30:00"
}
```

## 🛠️ 技术栈

- **后端**: Go (net/http, 纯标准库)
- **前端**: 原生 HTML + CSS + JavaScript（无框架）
- **数据源**: `/proc` 文件系统 + systemctl + ss 命令
- **刷新**: 每 3 秒 AJAX 自动刷新

## 📋 系统要求

- **操作系统**: Linux（推荐 Debian/Ubuntu/CentOS）
- **Go 版本**: 1.18+（编译需要）
- **权限**: 读取 `/proc`、执行 `ss`、`systemctl` 命令

## 🔐 安全建议

- 不要在公网直接暴露端口
- 使用 Nginx 反向代理 + HTTPS
- 或配合 Tailscale/WireGuard 内网访问
- 可添加简单认证（Nginx basic auth）

### Nginx 反向代理示例

```nginx
server {
    listen 443 ssl;
    server_name monitor.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    auth_basic "Monitor";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
    }
}
```

## 📝 版本历史

| 版本 | 更新内容 |
|:---:|:---|
| v19 | 添加昼夜模式切换、移动端适配、更新 README |
| v18 | 端口服务类型识别、开发环境检测 |
| v17 | 网卡状态、流量统计、系统日志 |
| v16 | 基础系统监控面板 |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License - 自由使用、修改、分发。

---

Made with ❤️ by [LceAn](https://github.com/LceAn)