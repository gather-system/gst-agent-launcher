# GST Agent Launcher

Go + Bubble Tea TUI 應用程式，用於管理和啟動 Claude Code Agent。

## 專案資訊

- **Module**: `github.com/gather-system/gst-agent-launcher`
- **產出**: `gst-launcher.exe`
- **技術**: Go 1.22+ / Bubble Tea v2 / Lip Gloss v2

## 建置與執行

```bash
go build -o gst-launcher.exe .
./gst-launcher.exe
```

## 專案結構

```
gst-agent-launcher/
├── main.go              # 程式進入點
├── config/
│   ├── config.go        # 設定檔載入邏輯
│   └── default.json     # 內嵌預設 Agent 定義
├── tui/
│   ├── model.go         # Bubble Tea Model
│   ├── view.go          # 畫面渲染
│   ├── update.go        # 按鍵/事件處理
│   └── styles.go        # Lip Gloss 樣式
├── launcher/
│   └── wt.go            # Windows Terminal 啟動邏輯
└── go.mod
```

## 開發規範

- 使用 `gofmt` 格式化程式碼
- 錯誤處理遵循 Go 慣例（回傳 error，不 panic）
- 設定檔優先順序：`~/.config/gst-launcher/agents.json` > exe 同目錄 > 內嵌預設
- Bubble Tea v2 import 路徑為 `charm.land/bubbletea/v2`
- Lip Gloss v2 import 路徑為 `charm.land/lipgloss/v2`
