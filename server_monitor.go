package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SystemInfo struct {
	CPU       CPUInfo           `json:"cpu"`
	Memory    MemoryInfo        `json:"memory"`
	Swap      SwapInfo          `json:"swap"`
	Disks     []DiskInfo        `json:"disks"`
	DiskTotal DiskTotal         `json:"disk_total"`
	Networks  []NetworkIf       `json:"networks"`
	NetStats  NetStats          `json:"net_stats"`
	Processes []ProcessInfo     `json:"processes"`
	Ports     []PortService     `json:"ports"`
	Services  map[string]string `json:"services"`
	Versions  Versions          `json:"versions"`
	Logs      []LogEntry        `json:"logs"`
	Uptime    string            `json:"uptime"`
	Timestamp string            `json:"timestamp"`
}

type CPUInfo struct {
	Model   string  `json:"model"`
	Cores   int     `json:"cores"`
	LoadAvg string  `json:"load_avg"`
	Temp    float64 `json:"temp"`
}

type MemoryInfo struct {
	TotalMB int     `json:"total_mb"`
	UsedMB  int     `json:"used_mb"`
	FreeMB  int     `json:"free_mb"`
	Percent float64 `json:"percent"`
}

type SwapInfo struct {
	TotalMB int `json:"total_mb"`
	UsedMB  int `json:"used_mb"`
}

type DiskInfo struct {
	Device  string `json:"device"`
	Total   string `json:"total"`
	Used    string `json:"used"`
	Percent string `json:"percent"`
	Mount   string `json:"mount"`
}

type DiskTotal struct {
	TotalGB float64 `json:"total_gb"`
	UsedGB  float64 `json:"used_gb"`
	Percent string  `json:"percent"`
}

type NetworkIf struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	IPv4   string `json:"ipv4"`
	IPv6   string `json:"ipv6"`
	RxMB   int    `json:"rx_mb"`
	TxMB   int    `json:"tx_mb"`
}

type NetStats struct {
	TCPConns    int `json:"tcp_conns"`
	UDPConns    int `json:"udp_conns"`
	TimeWait    int `json:"time_wait"`
	Established int `json:"established"`
}

type ProcessInfo struct {
	PID     string  `json:"pid"`
	Name    string  `json:"name"`
	CPU     float64 `json:"cpu"`
	Mem     float64 `json:"mem"`
	User    string  `json:"user"`
	Command string  `json:"command"`
}

type PortService struct {
	Port    string `json:"port"`
	Service string `json:"service"`
	IPv4    string `json:"ipv4"`
	IPv6    string `json:"ipv6"`
	Proto   string `json:"proto"`
	Type    string `json:"type"`
	Star    bool   `json:"star"`
	Process string `json:"process"`
	PID     string `json:"pid"`
}

type Versions struct {
	GCC      string `json:"gcc"`
	Make     string `json:"make"`
	CMake    string `json:"cmake"`
	Python   string `json:"python"`
	Pip      string `json:"pip"`
	Node     string `json:"node"`
	Npm      string `json:"npm"`
	Go       string `json:"go"`
	Java     string `json:"java"`
	Postgres string `json:"postgres"`
	MariaDB  string `json:"mariadb"`
	Redis    string `json:"redis"`
	Docker   string `json:"docker"`
	Compose  string `json:"compose"`
	Git      string `json:"git"`
	Vim      string `json:"vim"`
}

type LogEntry struct {
	Time    string `json:"time"`
	Service string `json:"service"`
	Message string `json:"message"`
}

var importantPorts = map[string]bool{
	"22":    true,
	"80":    true,
	"443":   true,
	"3306":  true,
	"5432":  true,
	"6379":  true,
	"27017": true,
}

var systemServices = map[string]bool{
	"sshd":         true,
	"nginx":        true,
	"apache2":      true,
	"postgres":     true,
	"mysqld":       true,
	"mariadbd":     true,
	"redis-server": true,
	"systemd":      true,
	"cron":         true,
	"dockerd":      true,
	"containerd":   true,
	"tailscaled":   true,
}

func readLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func getCPU() CPUInfo {
	cpu := CPUInfo{}
	for _, line := range readLines("/proc/cpuinfo") {
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cpu.Model = strings.TrimSpace(parts[1])
			}
			break
		}
	}
	if lines := readLines("/proc/loadavg"); len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			cpu.LoadAvg = parts[0] + " " + parts[1] + " " + parts[2]
		}
	}
	for _, line := range readLines("/proc/cpuinfo") {
		if strings.HasPrefix(line, "processor") {
			cpu.Cores++
		}
	}
	if data, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp"); err == nil {
		tempStr := strings.TrimSpace(string(data))
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			cpu.Temp = temp / 1000.0
		}
	}
	return cpu
}

func getMemory() MemoryInfo {
	mem := MemoryInfo{}
	for _, line := range readLines("/proc/meminfo") {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, _ := strconv.Atoi(parts[1])
				mem.TotalMB = kb / 1024
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, _ := strconv.Atoi(parts[1])
				mem.FreeMB = kb / 1024
			}
		}
	}
	mem.UsedMB = mem.TotalMB - mem.FreeMB
	if mem.TotalMB > 0 {
		mem.Percent = float64(mem.UsedMB) / float64(mem.TotalMB) * 100
	}
	return mem
}

func getSwap() SwapInfo {
	swap := SwapInfo{}
	for _, line := range readLines("/proc/meminfo") {
		if strings.HasPrefix(line, "SwapTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, _ := strconv.Atoi(parts[1])
				swap.TotalMB = kb / 1024
			}
		} else if strings.HasPrefix(line, "SwapFree:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, _ := strconv.Atoi(parts[1])
				swap.UsedMB = swap.TotalMB - kb/1024
			}
		}
	}
	return swap
}

func getDisks() ([]DiskInfo, DiskTotal) {
	mounts, _ := ioutil.ReadFile("/proc/mounts")
	var disks []DiskInfo
	var totalKB, usedKB int64
	for _, line := range strings.Split(string(mounts), "\n") {
		if !strings.HasPrefix(line, "/dev/") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		device := parts[0]
		mount := parts[1]
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err != nil {
			continue
		}
		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - free
		totalKB += int64(total / 1024)
		usedKB += int64(used / 1024)
		percent := 0
		if total > 0 {
			percent = int(used * 100 / total)
		}
		disks = append(disks, DiskInfo{
			Device:  device,
			Total:   formatSize(total),
			Used:    formatSize(used),
			Percent: fmt.Sprintf("%d%%", percent),
			Mount:   mount,
		})
	}
	dt := DiskTotal{
		TotalGB: float64(totalKB) / 1024 / 1024,
		UsedGB:  float64(usedKB) / 1024 / 1024,
	}
	if totalKB > 0 {
		dt.Percent = fmt.Sprintf("%.0f%%", float64(usedKB)/float64(totalKB)*100)
	}
	return disks, dt
}

func formatSize(bytes uint64) string {
	gb := float64(bytes) / 1024 / 1024 / 1024
	if gb >= 1 {
		return fmt.Sprintf("%.0fG", gb)
	}
	mb := float64(bytes) / 1024 / 1024
	return fmt.Sprintf("%.0fM", mb)
}

func getNetworks() []NetworkIf {
	var nets []NetworkIf
	links, _ := ioutil.ReadDir("/sys/class/net")
	for _, link := range links {
		name := link.Name()
		status := "DOWN"
		if data, err := ioutil.ReadFile("/sys/class/net/" + name + "/operstate"); err == nil {
			status = strings.TrimSpace(string(data))
			status = strings.ToUpper(status)
		}
		var ipv4, ipv6 string
		var rxMB, txMB int
		out, _ := exec.Command("ip", "addr", "show", name).Output()
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "inet ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					ipv4 = parts[1]
				}
			} else if strings.HasPrefix(line, "inet6 ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 && !strings.HasPrefix(parts[1], "fe80:") {
					ipv6 = parts[1]
				}
			}
		}
		if rx, _ := ioutil.ReadFile("/sys/class/net/" + name + "/statistics/rx_bytes"); len(rx) > 0 {
			rxVal, _ := strconv.ParseInt(strings.TrimSpace(string(rx)), 10, 64)
			rxMB = int(rxVal / 1024 / 1024)
		}
		if tx, _ := ioutil.ReadFile("/sys/class/net/" + name + "/statistics/tx_bytes"); len(tx) > 0 {
			txVal, _ := strconv.ParseInt(strings.TrimSpace(string(tx)), 10, 64)
			txMB = int(txVal / 1024 / 1024)
		}
		nets = append(nets, NetworkIf{
			Name: name, Status: status, IPv4: ipv4, IPv6: ipv6,
			RxMB: rxMB, TxMB: txMB,
		})
	}
	return nets
}

func getNetStats() NetStats {
	stats := NetStats{}
	out2, _ := exec.Command("ss", "-tan").Output()
	for _, line := range strings.Split(string(out2), "\n") {
		if strings.Contains(line, "ESTAB") {
			stats.Established++
		}
		if strings.Contains(line, "TIME-WAIT") {
			stats.TimeWait++
		}
		if strings.HasPrefix(line, "LISTEN") || strings.Contains(line, "ESTAB") || strings.Contains(line, "TIME-WAIT") {
			stats.TCPConns++
		}
	}
	return stats
}

func getProcesses() []ProcessInfo {
	var processes []ProcessInfo
	out, _ := exec.Command("ps", "aux", "--sort=-pcpu").Output()
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 || i > 10 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		cmd := fields[10]
		if len(cmd) > 25 {
			cmd = cmd[:25] + "..."
		}
		processes = append(processes, ProcessInfo{
			PID: fields[1], User: fields[0], CPU: cpu, Mem: mem, Command: cmd,
		})
	}
	return processes
}

func getExePath(pid string) string {
	exe, _ := os.Readlink(fmt.Sprintf("/proc/%s/exe", pid))
	return exe
}

func detectServiceType(name string, pid string) (string, string) {
	cmdlinePath := fmt.Sprintf("/proc/%s/cmdline", pid)
	cmdline, err := ioutil.ReadFile(cmdlinePath)
	if err != nil {
		return "other", name
	}
	cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
	cmdStr = strings.TrimSpace(cmdStr)
	cmdLower := strings.ToLower(cmdStr)
	nameLower := strings.ToLower(name)

	if strings.Contains(cmdLower, "docker") || strings.Contains(cmdLower, "containerd") ||
		strings.Contains(nameLower, "docker") || strings.Contains(nameLower, "containerd") {
		return "docker", cmdStr
	}
	if systemServices[name] || systemServices[nameLower] {
		return "system", cmdStr
	}
	exePath := getExePath(pid)
	if strings.Contains(exePath, "server_monitor") {
		return "go", cmdStr
	}
	if strings.Contains(cmdLower, "python") || strings.Contains(cmdLower, ".py") ||
		strings.Contains(cmdLower, "gunicorn") || strings.Contains(cmdLower, "uwsgi") {
		return "python", cmdStr
	}
	if strings.Contains(cmdLower, "node") || strings.Contains(cmdLower, ".js") ||
		strings.Contains(cmdLower, "npm") || strings.Contains(cmdLower, "pm2") {
		return "node", cmdStr
	}
	if strings.Contains(cmdLower, "java") || strings.Contains(cmdLower, ".jar") ||
		strings.Contains(cmdLower, "tomcat") || strings.Contains(cmdLower, "jetty") {
		return "java", cmdStr
	}
	if strings.Contains(cmdLower, "php") || strings.Contains(cmdLower, "php-fpm") {
		return "php", cmdStr
	}
	if strings.Contains(cmdLower, "ruby") || strings.Contains(cmdLower, "rails") ||
		strings.Contains(cmdLower, "puma") || strings.Contains(cmdLower, "unicorn") {
		return "ruby", cmdStr
	}
	if strings.Contains(cmdLower, "cargo") || strings.Contains(exePath, "rust") {
		return "rust", cmdStr
	}
	return "other", cmdStr
}

func getPorts() []PortService {
	portMap := make(map[string]*PortService)
	out, _ := exec.Command("ss", "-tlnp").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		localAddr := fields[3]
		addrParts := strings.Split(localAddr, ":")
		if len(addrParts) < 2 {
			continue
		}
		port := addrParts[len(addrParts)-1]
		ip := strings.Join(addrParts[:len(addrParts)-1], ":")
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")

		name := ""
		pid := ""
		for i := 4; i < len(fields); i++ {
			if strings.Contains(fields[i], "users:") {
				re := strings.Split(fields[i], "\"")
				if len(re) >= 2 {
					name = re[1]
				}
				if strings.Contains(fields[i], "pid=") {
					pidStart := strings.Index(fields[i], "pid=") + 4
					pidEnd := strings.Index(fields[i][pidStart:], ",")
					if pidEnd > 0 {
						pid = fields[i][pidStart : pidStart+pidEnd]
					}
				}
				break
			}
		}

		if name != "" {
			key := port + "_" + name + "_" + pid
			if existing, ok := portMap[key]; ok {
				if strings.Contains(ip, ":") {
					existing.IPv6 = ip
				} else {
					existing.IPv4 = ip
				}
			} else {
				svcType, process := detectServiceType(name, pid)
				ps := &PortService{
					Port: port, Service: name, IP: ip, Proto: "tcp",
					Type: svcType, Star: importantPorts[port], Process: process, PID: pid,
				}
				if strings.Contains(ip, ":") {
					ps.IPv6 = ip
					ps.IPv4 = ""
				} else {
					ps.IPv4 = ip
					ps.IPv6 = ""
				}
				portMap[key] = ps
			}
		}
	}

	var ports []PortService
	for _, ps := range portMap {
		ports = append(ports, *ps)
	}
	return ports
}

func getUptime() string {
	data, _ := ioutil.ReadFile("/proc/uptime")
	parts := strings.Fields(string(data))
	if len(parts) > 0 {
		secs, _ := strconv.ParseFloat(parts[0], 64)
		days := int(secs) / 86400
		hours := (int(secs) % 86400) / 3600
		mins := (int(secs) % 3600) / 60
		if days > 0 {
			return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
		}
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return ""
}

func getVer(cmd string) string {
	out, _ := exec.Command("sh", "-c", cmd).Output()
	return strings.TrimSpace(string(out))
}

func getVersions() Versions {
	return Versions{
		GCC:      getVer("gcc --version | head -1 | awk '{print $3}' | cut -d'-' -f1"),
		Make:     getVer("make --version | head -1 | awk '{print $3}'"),
		CMake:    getVer("cmake --version | head -1 | awk '{print $3}'"),
		Python:   getVer("python3 --version | awk '{print $2}'"),
		Pip:      getVer("pip3 --version 2>/dev/null | awk '{print $2}' | cut -d'/' -f1"),
		Node:     strings.TrimPrefix(getVer("node --version"), "v"),
		Npm:      getVer("npm --version"),
		Go:       strings.TrimPrefix(getVer("go version | awk '{print $3}'"), "go"),
		Java:     getVer("java -version 2>&1 | head -1 | cut -d'\"' -f2"),
		Postgres: getVer("psql --version 2>/dev/null | awk '{print $3}'"),
		MariaDB:  getVer("mysql --version 2>/dev/null | awk '{print $5}' | tr -d ','"),
		Redis:    getVer("redis-server --version 2>/dev/null | awk -F'=' '{print $2}'"),
		Docker:   getVer("docker --version 2>/dev/null | grep -oP 'version \\K[0-9.]+'"),
		Compose:  getVer("docker-compose --version 2>/dev/null | grep -oP 'version \\K[0-9.]+'"),
		Git:      getVer("git --version | awk '{print $3}'"),
		Vim:      getVer("vim --version | head -1 | awk '{print $5}'"),
	}
}

func getLogs() []LogEntry {
	var logs []LogEntry
	out, _ := exec.Command("journalctl", "-n", "8", "--no-pager", "-o", "short-iso").Output()
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		timeStr := fields[0]
		service := fields[4]
		if strings.HasPrefix(service, "[") && strings.HasSuffix(service, "]") {
			service = service[1 : len(service)-1]
		} else {
			service = "kernel"
		}
		msg := strings.Join(fields[5:], " ")
		if len(msg) > 50 {
			msg = msg[:50] + "..."
		}
		logs = append(logs, LogEntry{Time: timeStr[11:19], Service: service, Message: msg})
	}
	return logs
}

func getSystemInfo() SystemInfo {
	disks, diskTotal := getDisks()
	services := map[string]string{}
	for _, svc := range []string{"docker", "nginx", "redis-server", "postgresql", "mariadb", "ssh"} {
		out, _ := exec.Command("systemctl", "is-active", svc).Output()
		status := strings.TrimSpace(string(out))
		if status != "active" {
			status = "inactive"
		}
		services[svc] = status
	}
	return SystemInfo{
		CPU:       getCPU(),
		Memory:    getMemory(),
		Swap:      getSwap(),
		Disks:     disks,
		DiskTotal: diskTotal,
		Networks:  getNetworks(),
		NetStats:  getNetStats(),
		Processes: getProcesses(),
		Ports:     getPorts(),
		Services:  services,
		Versions:  getVersions(),
		Logs:      getLogs(),
		Uptime:    getUptime(),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

const html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Server Monitor</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:Inter,-apple-system,sans-serif;background:#0d1117;color:#c9d1d9;min-height:100vh;padding:10px}
.wrap{max-width:1600px;margin:0 auto}
.hd{text-align:center;padding:14px 0 10px;border-bottom:1px solid #21262d;margin-bottom:10px}
.hd h1{font-size:1.5em;font-weight:600;color:#58a6ff;display:inline-flex;align-items:center;gap:8px}
.badge{background:linear-gradient(135deg,#f85149,#da3633);color:#fff;padding:2px 8px;border-radius:10px;font-size:.6em;animation:pulse 2s infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.6}}
.badge::before{content:"";display:inline-block;width:5px;height:5px;background:#fff;border-radius:50%;margin-right:4px;animation:blink 1s infinite}
@keyframes blink{0%,100%{opacity:1}50%{opacity:.3}}
.tabs{display:flex;justify-content:center;gap:4px;margin-bottom:10px}
.tab{padding:6px 16px;border-radius:5px;background:#161b22;border:1px solid #30363d;cursor:pointer;font-size:.75em;color:#8b949e}
.tab:hover{background:#21262d;color:#c9d1d9}
.tab.on{background:#388bfd1a;color:#58a6ff;border-color:#58a6ff66}
.page{display:none}
.page.on{display:block}

/* 统计条 */
.stats{display:flex;gap:10px;margin-bottom:12px;flex-wrap:wrap}
.stat{flex:1;min-width:120px;background:#161b22;border:1px solid #30363d;border-radius:6px;padding:12px;text-align:center}
.stat .v{font-size:1.3em;font-weight:700;color:#58a6ff}
.stat .l{font-size:.65em;color:#8b949e;margin-top:2px}

/* 固定布局 - 左右两列，左边固定，右边可变 */
.fixed{display:grid;grid-template-columns:260px 1fr;gap:10px;margin-bottom:10px}
.left{display:flex;flex-direction:column;gap:10px}
.right{display:flex;flex-direction:column;gap:10px}

/* 卡片 */
.c{background:#161b22;border:1px solid #30363d;border-radius:6px;padding:12px;flex-shrink:0}
.c h3{font-size:.75em;font-weight:600;color:#58a6ff;margin-bottom:8px;padding-bottom:5px;border-bottom:1px solid #21262d}
.row{display:flex;justify-content:space-between;padding:2px 0;font-size:.7em}
.k{color:#6e7681}
.v{color:#7ee787;font-family:Consolas,monospace}
.bar{height:3px;background:#21262d;border-radius:2px;margin-top:3px}
.bar>div{height:100%;background:linear-gradient(90deg,#238636,#2ea043);border-radius:2px;transition:width .3s}
.bar.w>div{background:linear-gradient(90deg,#d29922,#e3b341)}
.bar.d>div{background:linear-gradient(90deg,#da3633,#f85149)}

/* 标签 */
.t{display:inline-block;padding:1px 6px;border-radius:3px;font-size:.65em;font-weight:500;margin:1px}
.t.sys{background:#1f6feb33;color:#58a6ff}
.t.ok{background:#23863633;color:#3fb950}
.t.no{background:#da363333;color:#f85149}
.t.dk{background:#8957e533;color:#a371f7}
.t.go{background:#2ea04333;color:#56d364}
.t.py{background:#d2992233;color:#e3b341}
.t.jv{background:#f0883e33;color:#f0883e}

/* 服务标签区 */
.svc{display:flex;flex-wrap:wrap;gap:2px}

/* 动态区域 */
.dyn{display:grid;grid-template-columns:1fr 1fr;gap:10px}

/* 网卡 */
.ni{background:#0d1117;border-radius:4px;padding:6px 8px;margin:3px 0}
.ni>.r{display:flex;align-items:center;gap:5px;margin-bottom:3px}
.ni>.r .nm{color:#58a6ff;font-weight:600;font-size:.72em}
.ni>.r .st{font-size:.58em;padding:0 5px;border-radius:8px;background:#23863633;color:#3fb950}
.ni>.r .st.down{background:#da363333;color:#f85149}

/* 进程表 */
.ph,.pr{display:grid;grid-template-columns:50px 50px 40px 40px 1fr;gap:4px;font-size:.68em;align-items:center}
.ph{color:#58a6ff;border-bottom:1px solid #21262d;padding:3px 0;margin-bottom:3px}
.pr{padding:2px 0;border-bottom:1px solid #21262d22}
.pr:hover{background:#ffffff05}

/* 日志 */
.lh,.lr{display:grid;grid-template-columns:42px 60px 1fr;gap:4px;font-size:.68em;align-items:center}
.lh{color:#58a6ff;border-bottom:1px solid #21262d;padding:3px 0;margin-bottom:3px}
.lr{padding:2px 0;border-bottom:1px solid #21262d22}
.lr .m{color:#8b949e;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}

/* 开发环境 */
.sec{margin-bottom:12px}
.sec .t{font-size:.75em;font-weight:600;color:#58a6ff;margin-bottom:6px;padding-bottom:3px;border-bottom:1px solid #21262d}
.dg{display:grid;grid-template-columns:repeat(auto-fill,minmax(150px,1fr));gap:6px}
.di{background:#0d1117;border-radius:4px;padding:8px;display:flex;justify-content:space-between;align-items:center}
.di .n{color:#6e7681;font-size:.68em}
.di .v{color:#7ee787;font-family:Consolas,monospace;font-size:.68em}

/* 端口表 - 固定列宽，防止错位 */
.ph2,.pr2{display:grid;grid-template-columns:55px 75px 200px 95px 55px 20px;gap:6px;font-size:.7em;align-items:center}
.ph2{color:#58a6ff;border-bottom:1px solid #30363d;padding:5px 0;margin-bottom:4px;font-weight:600;background:#161b22;position:sticky;top:0}
.pr2{padding:4px 0;border-bottom:1px solid #21262d}
.pr2:hover{background:#ffffff08}
.pr2.st{background:#d299220a}
.pn{color:#d29922;font-weight:700}
.st{color:#ffd700}
/* 各列固定样式 */
.c1{font-weight:700;color:#d29922}
.c2{color:#c9d1d9}
.c3{color:#8b949e;font-size:.65em;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.c4{color:#6e7681;font-size:.65em}
.c5{text-align:center}
.c6{text-align:center;color:#ffd700}

.up{color:#484f58;font-size:.65em;text-align:center;margin-top:8px}
.foot{color:#30363d;font-size:.65em;text-align:center;margin-top:8px;padding:8px}
</style>
</head>
<body>
<div class="wrap">
<div class="hd"><h1>🖥️ Debian Server<span class="badge">LIVE</span></h1></div>
<div class="tabs">
<div class="tab on" onclick="sw('sys')">📊 系统信息</div>
<div class="tab" onclick="sw('dev')">🔧 开发环境</div>
<div class="tab" onclick="sw('port')">🔌 端口服务</div>
</div>

<div id="sys" class="page on">
<!-- 顶部统计 -->
<div class="stats">
<div class="stat"><div class="v" id="su">-</div><div class="l">运行时间</div></div>
<div class="stat"><div class="v" id="sc">-</div><div class="l">CPU负载</div></div>
<div class="stat"><div class="v" id="sm">-</div><div class="l">内存使用</div></div>
<div class="stat"><div class="v" id="sd">-</div><div class="l">磁盘使用</div></div>
<div class="stat"><div class="v" id="sn">-</div><div class="l">网络连接</div></div>
</div>

<!-- 固定区域：左边固定，右边可变 -->
<div class="fixed">
<div class="left">
<div class="c" style="height:120px">
<h3>💻 处理器</h3>
<div class="row"><span class="k">型号</span><span class="v" id="cm">-</span></div>
<div class="row"><span class="k">核心</span><span class="v" id="cc">-</span></div>
<div class="row"><span class="k">负载</span><span class="v" id="cl">-</span></div>
<div class="row"><span class="k">温度</span><span class="v" id="ct">-</span></div>
</div>
<div class="c" style="height:95px">
<h3>🧠 内存</h3>
<div class="row"><span class="k">使用</span><span class="v" id="mp">-</span></div>
<div class="row"><span class="k">总计</span><span class="v" id="mt">-</span></div>
<div class="bar"><div id="mb"></div></div>
</div>
<div class="c" style="height:95px">
<h3>💾 磁盘</h3>
<div class="row"><span class="k">使用</span><span class="v" id="dp">-</span></div>
<div class="row"><span class="k">总计</span><span class="v" id="dt">-</span></div>
<div class="bar"><div id="db"></div></div>
</div>
<div class="c" style="height:75px">
<h3>⚙️ 服务</h3>
<div class="svc" id="sv"></div>
</div>
</div>

<div class="right">
<div class="c">
<h3>🌐 网络</h3>
<div id="ni" style="max-height:100px;overflow-y:auto"></div>
<div style="margin-top:6px;padding-top:6px;border-top:1px solid #21262d">
<div class="row"><span class="k">TCP连接</span><span class="v" id="tc">-</span></div>
<div class="row"><span class="k">已建立</span><span class="v" id="te">-</span></div>
</div>
</div>
</div>
</div>

<!-- 动态区域 -->
<div class="dyn">
<div class="c">
<h3>📈 Top进程</h3>
<div class="ph"><span>PID</span><span>用户</span><span>CPU%</span><span>MEM%</span><span>命令</span></div>
<div id="pr" style="max-height:140px;overflow-y:auto"></div>
</div>
<div class="c">
<h3>📋 系统日志</h3>
<div class="lh"><span>时间</span><span>服务</span><span>消息</span></div>
<div id="lg" style="max-height:140px;overflow-y:auto"></div>
</div>
</div>
</div>

<div id="dev" class="page">
<div class="sec"><div class="t">🔨 编译工具</div><div class="dg" id="dv1"></div></div>
<div class="sec"><div class="t">🐹 编程语言</div><div="dg" id="dv2"></div></div>
<div class="sec"><div class="t">🗄️ 数据库</div><div class="dg" id="dv3"></div></div>
<div class="sec"><div class="t">🐳 容器</div><div class="dg" id="dv4"></div></div>
<div class="sec"><div class="t">🛠️ 工具</div><div class="dg" id="dv5"></div></div>
</div>

<div id="port" class="page">
<div class="stats">
<div class="stat"><div class="v" id="pc">-</div><div class="l">监听端口</div></div>
<div class="stat"><div class="v" id="sc2">-</div><div class="l">运行服务</div></div>
<div class="stat"><div class="v" id="dc">-</div><div class="l">Docker</div></div>
</div>
<div style="margin-bottom:6px;font-size:.68em">
<span class="t sys">system</span>
<span class="t dk">docker</span>
<span class="t go">go</span>
<span class="t py">python</span>
<span class="t no">node</span>
<span class="t jv">java</span>
<span style="color:#ffd700;margin-left:6px">★ 重点</span>
</div>
<div class="c">
<h3>🔌 端口监听</h3>
<div class="ph2"><span class="c1">端口</span><span class="c2">服务</span><span class="c3">监听地址</span><span class="c4">类型</span><span class="c5">PID</span><span class="c6">★</span></div>
<div id="pt" style="max-height:calc(100vh - 260px);overflow-y:auto"></div>
</div>
</div>

<div class="up">更新: <span id="ts">-</span></div>
<div class="foot">Debian 11 · 内核 5.10.0 · 每3秒刷新</div>
</div>

<script>
function sw(p){
document.querySelectorAll('.tab').forEach(t=>t.classList.remove('on'));
document.querySelectorAll('.page').forEach(t=>t.classList.remove('on'));
document.querySelector('.tab:nth-child('+(p==='sys'?1:p==='dev'?2:3)+')').classList.add('on');
document.getElementById(p).classList.add('on');
}
var tc={system:'t sys',docker:'t dk',go:'t go',python:'t py',node:'t no',java:'t jv',php:'t py',ruby:'t no',rust:'t go',other:''};
function u(){
fetch('/api').then(r=>r.json()).then(d=>{
document.getElementById('su').textContent=d.uptime;
document.getElementById('sc').textContent=d.cpu.load_avg.split(' ')[0];
document.getElementById('sm').textContent=d.memory.percent.toFixed(0)+'%';
document.getElementById('sd').textContent=d.disk_total.percent;
document.getElementById('sn').textContent=d.net_stats.established;
document.getElementById('cm').textContent=d.cpu.model||'-';
document.getElementById('cc').textContent=d.cpu.cores+'核';
document.getElementById('cl').textContent=d.cpu.load_avg;
document.getElementById('ct').textContent=d.cpu.temp>0?d.cpu.temp.toFixed(0)+'°C':'-';
document.getElementById('mp').textContent=d.memory.percent.toFixed(1)+'% ('+d.memory.used_mb+'M)';
document.getElementById('mt').textContent=d.memory.total_mb+' MB';
document.getElementById('mb').style.width=d.memory.percent+'%';
document.getElementById('mb').className=d.memory.percent>80?'bar d':d.memory.percent>60?'bar w':'bar';
document.getElementById('dp').textContent=d.disk_total.percent+' ('+d.disk_total.used_gb.toFixed(1)+'G)';
document.getElementById('dt').textContent=d.disk_total.total_gb.toFixed(0)+' GB';
document.getElementById('db').style.width=d.disk_total.percent;

let sv='';Object.entries(d.services).forEach(([k,v])=>{sv+='<span class="t '+(v==='active'?'ok':'no')+'">'+k+'</span>'});
document.getElementById('sv').innerHTML=sv;

let ni='';d.networks.forEach(n=>{
const s=n.status==='UP'?'':'down';
ni+='<div class="ni"><div class="r"><span class="nm">'+n.name+'</span><span class="st '+s+'">'+n.status+'</span></div>';
if(n.ipv4)ni+='<div class="row"><span class="k">IPv4</span><span class="v" style="font-size:.68em">'+n.ipv4+'</span></div>';
if(n.ipv6)ni+='<div class="row"><span class="k">IPv6</span><span class="v" style="font-size:.65em">'+n.ipv6+'</span></div>';
ni+='<div class="row"><span class="k">流量</span><span class="v">↓'+n.rx_mb+'M ↑'+n.tx_mb+'M</span></div></div>';
});
document.getElementById('ni').innerHTML=ni;
document.getElementById('tc').textContent=d.net_stats.tcp_conns;
document.getElementById('te').textContent=d.net_stats.established;

let pr='';d.processes.forEach(p=>{pr+='<div class="pr"><span>'+p.pid+'</span><span>'+p.user+'</span><span>'+p.cpu.toFixed(1)+'</span><span>'+p.mem.toFixed(1)+'</span><span>'+p.command+'</span></div>'});
document.getElementById('pr').innerHTML=pr;

let lg='';d.logs.forEach(l=>{lg+='<div class="lr"><span>'+l.time+'</span><span class="t sys">'+l.service+'</span><span class="m">'+l.message+'</span></div>'});
document.getElementById('lg').innerHTML=lg;

const v=d.versions;
document.getElementById('dv1').innerHTML=[['GCC',v.gcc],['Make',v.make],['CMake',v.cmake]].map(x=>'<div class="di"><span class="n">'+x[0]+'</span><span class="v">'+(x[1]||'-')+'</span></div>').join('');
document.getElementById('dv2').innerHTML=[['Python',v.python],['pip',v.pip],['Node',v.node],['npm',v.npm],['Go',v.go],['Java',v.java]].map(x=>'<div class="di"><span class="n">'+x[0]+'</span><span class="v">'+(x[1]||'-')+'</span></div>').join('');
document.getElementById('dv3').innerHTML=[['PostgreSQL',v.postgres],['MariaDB',v.mariadb],['Redis',v.redis]].map(x=>'<div class="di"><span class="n">'+x[0]+'</span><span class="v">'+(x[1]||'-')+'</span></div>').join('');
document.getElementById('dv4').innerHTML=[['Docker',v.docker],['Compose',v.compose]].map(x=>'<div class="di"><span class="n">'+x[0]+'</span><span class="v">'+(x[1]||'-')+'</span></div>').join('');
document.getElementById('dv5').innerHTML=[['Git',v.git],['Vim',v.vim],['tmux','3.1c'],['htop','3.0.5'],['jq','1.6'],['rg','13.0'],['bat','0.13'],['fd','8.3']].map(x=>'<div class="di"><span class="n">'+x[0]+'</span><span class="v">'+(x[1]||'-')+'</span></div>').join('');

let pt='';d.ports.forEach(p=>{
let ip=p.ipv4||'';
if(p.ipv6){ip=ip?(ip+' | '+p.ipv6):p.ipv6;}
if(ip.length>30)ip=ip.substring(0,30)+'...';
pt+='<div class="pr2'+(p.star?' st':'')+'"><span class="c1">'+p.port+'</span><span class="c2">'+p.service+'</span><span class="c3">'+ip+'</span><span class="c4"><span class="'+(tc[p.type]||'')+'">'+p.type+'</span></span><span class="c4">'+p.pid+'</span><span class="c6">'+(p.star?'★':'')+'</span></div>';
});
document.getElementById('pt').innerHTML=pt;
document.getElementById('pc').textContent=d.ports.length;
let us={};d.ports.forEach(p=>{us[p.service]=1});
document.getElementById('sc2').textContent=Object.keys(us).length;
let dc=0;d.ports.forEach(p=>{if(p.type==='docker')dc++});
document.getElementById('dc').textContent=dc;
document.getElementById('ts').textContent=d.timestamp;
})}
u();setInterval(u,3000);
</script>
</body>
</html>`

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(getSystemInfo())
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}