# XTerminal - Linux 系统状态监控面板

一个轻量级的 Linux 系统实时状态监控 Web 面板，通过 Shell 脚本采集系统数据，前端页面自动刷新展示。

## 功能

- **CPU** - 使用率、温度、频率
- **内存** - 使用情况、Swap
- **磁盘** - 分区大小、IO 统计
- **网络** - 流量统计
- **GPU** - 显卡信息（如有）
- **进程** - Top 进程列表
- **系统** - OS 版本、内核、运行时间
- **时间** - 系统时间、时区

## 部署

1. 将 `sub/` 目录和 `info.sh` 放到 `~/.xterminal/` 下
2. 将 `web/index.html` 放到 Web 服务器根目录（如 `/var/www/html/`）
3. 设置 cron 定时执行 `info.sh`：

```bash
* * * * * ~/.xterminal/info.sh cpu memory network_stats fs_size fs_stats gpu os process time
```

## 技术栈

- 前端：纯 HTML + CSS + JavaScript（无依赖）
- 后端：Bash 脚本
- Web 服务器：Nginx

## 截图

![预览](web/preview.png)

## 版本

当前版本：18
