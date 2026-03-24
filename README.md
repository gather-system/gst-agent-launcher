# gst-agent-launcher

GST Multi-Agent Launcher — Bubble Tea TUI for managing and monitoring coding agent terminals.

## Features

### Agent Management
- **Agent 清單** — 分群顯示所有 Agent（PM / Core / App / Leyu）
- **多選啟動** — Space 勾選，Enter 一鍵開啟多個 Windows Terminal tab
- **群組快選** — `c`/`p`/`o`/`l` 一鍵切換群組
- **專案模式** — `P` 鍵選擇預設 Agent 組合
- **搜尋過濾** — `/` 鍵 fuzzy 搜尋
- **依賴感知** — 啟動 App Agent 時自動提示未啟動的 Core 依賴

### Status Awareness (v0.4.0)
- **健康檢查** — 自動偵測路徑、Git repo、merge conflict 狀態
- **Git 狀態** — 顯示 branch name、dirty count、Issue ID
- **PR 偵測** — 透過 `gh` CLI 偵測 open PR，顯示 `[PR]` badge
- **Process 偵測** — 偵測已運行的 Agent，顯示 `[R]` badge
- **Dashboard** — `d` 鍵切換表格視圖，30 秒自動刷新

### Batch Operations (v0.4.0)
- **批次 Pull** — `g` 鍵對所有 repo 執行 `git pull`
- **批次 Status** — `G` 鍵對所有 repo 執行 `git status`
- **結果表格** — 顯示成功/失敗狀態與 output

## Build

```bash
# Development build
go build -o gst-launcher.exe .

# Production build (with version info)
pwsh ./build.ps1
```

## Usage

```bash
./gst-launcher.exe
./gst-launcher.exe --version
```

## Keybindings

| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | Navigate |
| `Space` | Toggle selection |
| `Enter` | Launch selected agents |
| `a` | Select/deselect all |
| `c` `p` `o` `l` | Toggle group |
| `d` | Dashboard view |
| `g` | Git Pull all |
| `G` | Git Status all |
| `m` | Toggle Monitor |
| `M` | Launch Monitor only |
| `/` | Search filter |
| `P` | Project presets |
| `e` | Edit config |
| `r` | Restore last session |
| `?` | Help overlay |
| `q` | Quit |

## Configuration

Config file locations (priority order):
1. `~/.config/gst-launcher/agents.json`
2. `agents.json` next to executable
3. Embedded default

## Requirements

- Windows 10/11 with Windows Terminal
- Go 1.22+
- `git` (optional, for status features)
- `gh` CLI (optional, for PR detection)
