package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "modernc.org/sqlite"
)

type App struct {
	ctx      context.Context
	db       *sql.DB
	config   *Config
	monitor  *ClipboardMonitor
	tray     *SystemTray
	dataDir  string
	logFile  *os.File
	logMu    sync.Mutex
}

type Config struct {
	Hotkey      string `json:"hotkey"`
	SaveDays    int    `json:"save_days"`
	StoragePath string `json:"storage_path"`
}

type ClipboardRecord struct {
	ID          int64      `json:"id"`
	Content     string     `json:"content"`
	CreatedAt   time.Time  `json:"created_at"`
	IsFavorited bool       `json:"is_favorited"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

func NewApp() *App {
	return &App{
		config: &Config{
			Hotkey:   "Ctrl+Shift+V",
			SaveDays: 30,
		},
	}
}

func (a *App) log(msg string) {
	a.logMu.Lock()
	defer a.logMu.Unlock()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)
	fmt.Print(line)
	if a.logFile != nil {
		a.logFile.WriteString(line)
		a.logFile.Sync()
	}
}

func (a *App) safeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.log(fmt.Sprintf("PANIC in %s: %v", name, r))
			}
		}()
		fn()
	}()
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.log("=== 启动 ===")

	// 第一步：获取数据路径
	var err error
	a.dataDir, err = a.getDataPath()
	if err != nil {
		a.log(fmt.Sprintf("ERROR: 获取数据路径失败: %v", err))
		return
	}

	if a.dataDir == "" {
		a.log("ERROR: 数据路径为空!")
		return
	}

	a.log(fmt.Sprintf("数据目录: %s", a.dataDir))

	// 第二步：创建数据目录
	if err := os.MkdirAll(a.dataDir, 0755); err != nil {
		a.log(fmt.Sprintf("ERROR: 创建数据目录失败: %v", err))
		return
	}

	// 第三步：初始化日志文件
	logPath := filepath.Join(a.dataDir, "log.txt")
	a.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		a.log(fmt.Sprintf("WARNING: 无法打开日志文件: %v", err))
	} else {
		a.log(fmt.Sprintf("日志文件: %s", logPath))
	}

	// 第四步：加载配置
	a.loadConfig()

	// 第五步：初始化数据库
	dbPath := filepath.Join(a.dataDir, "cliplite.db")
	a.log(fmt.Sprintf("数据库: %s", dbPath))

	if err := a.initDatabase(dbPath); err != nil {
		a.log(fmt.Sprintf("ERROR: 数据库初始化失败: %v", err))
	} else {
		a.log("数据库初始化成功 ✓")
	}

	// 第六步：启动其他组件
	a.monitor = NewClipboardMonitor(a)
	a.safeGo("monitor", a.monitor.Start)
	a.log("剪贴板监控已启动 ✓")

	a.safeGo("cleanup", a.cleanupExpiredRecords)

	a.safeGo("tray", a.initTray)

	a.log("=== 初始化完成 ===")
}

func (a *App) getDataPath() (string, error) {
	var candidates []string

	execPath, err := os.Executable()
	if err == nil && execPath != "" {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates, filepath.Join(execDir, "data"))
	}

	wd, err := os.Getwd()
	if err == nil && wd != "" {
		candidates = append(candidates, filepath.Join(wd, "data"))
	}

	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		candidates = append(candidates, filepath.Join(homeDir, ".cliplite"))
	}

	tempDir := os.TempDir()
	if tempDir != "" {
		candidates = append(candidates, filepath.Join(tempDir, "cliplite"))
	}

	for i, path := range candidates {
		a.log(fmt.Sprintf("尝试路径 %d: %s", i+1, path))
		if err := os.MkdirAll(path, 0755); err == nil {
			a.log(fmt.Sprintf("选择路径: %s", path))
			return path, nil
		} else {
			a.log(fmt.Sprintf("  路径不可用: %v", err))
		}
	}

	return "", fmt.Errorf("所有备选路径都不可用: %v", candidates)
}

func (a *App) shutdown(ctx context.Context) {
	a.log("关闭中...")
	if a.tray != nil {
		a.tray.Stop()
	}
	if a.monitor != nil {
		a.monitor.Stop()
	}
	if a.db != nil {
		a.db.Close()
	}
	if a.logFile != nil {
		a.logFile.Close()
	}
	a.log("已关闭")
}

func (a *App) loadConfig() {
	configPath := filepath.Join(a.dataDir, "config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			a.log("配置文件不存在，使用默认配置")
			return
		}
		a.log(fmt.Sprintf("读取配置文件失败: %v", err))
		return
	}

	var savedConfig Config
	if err := json.Unmarshal(data, &savedConfig); err != nil {
		a.log(fmt.Sprintf("解析配置文件失败: %v", err))
		return
	}

	if savedConfig.Hotkey != "" {
		a.config.Hotkey = savedConfig.Hotkey
	}
	if savedConfig.SaveDays > 0 {
		a.config.SaveDays = savedConfig.SaveDays
	}
	if savedConfig.StoragePath != "" {
		a.config.StoragePath = savedConfig.StoragePath
	}

	a.log(fmt.Sprintf("配置已加载: Hotkey=%s, SaveDays=%d", a.config.Hotkey, a.config.SaveDays))
}

func (a *App) saveConfig() {
	configPath := filepath.Join(a.dataDir, "config.json")
	
	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		a.log(fmt.Sprintf("序列化配置失败: %v", err))
		return
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		a.log(fmt.Sprintf("保存配置失败: %v", err))
		return
	}

	a.log("配置已保存")
}

func (a *App) initTray() {
	iconPath := ""
	
	execPath, err := os.Executable()
	if err == nil && execPath != "" {
		execDir := filepath.Dir(execPath)
		candidates := []string{
			filepath.Join(execDir, "icon.ico"),
			filepath.Join(execDir, "build", "windows", "icon.ico"),
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				iconPath = p
				break
			}
		}
	}

	a.tray = NewSystemTray(a)
	a.tray.iconPath = iconPath
	a.log(fmt.Sprintf("托盘图标: %s", iconPath))
	a.tray.Start()
}

func (a *App) ShowWindow() {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
	}
}

func (a *App) HideWindow() {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
	}
}

func (a *App) ExitApp() {
	a.log("用户请求退出")
	go func() {
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()
	if a.tray != nil {
		a.tray.Stop()
	}
	if a.monitor != nil {
		a.monitor.Stop()
	}
	if a.db != nil {
		a.db.Close()
	}
	if a.logFile != nil {
		a.logFile.Close()
	}
	os.Exit(0)
}

func (a *App) initDatabase(dbPath string) error {
	var err error
	a.db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	a.db.Exec("PRAGMA journal_mode=WAL")

	if err := a.db.Ping(); err != nil {
		a.db.Close()
		a.db = nil
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	_, err = a.db.Exec(`
	CREATE TABLE IF NOT EXISTS clipboard_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_favorited BOOLEAN DEFAULT 0,
		expires_at DATETIME
	)`)
	if err != nil {
		a.db.Close()
		a.db = nil
		return fmt.Errorf("创建表失败: %v", err)
	}

	a.db.Exec("CREATE INDEX IF NOT EXISTS idx_created_at ON clipboard_records(created_at)")
	a.db.Exec("CREATE INDEX IF NOT EXISTS idx_is_favorited ON clipboard_records(is_favorited)")
	a.db.Exec("CREATE INDEX IF NOT EXISTS idx_content ON clipboard_records(content)")

	return nil
}

func (a *App) cleanupExpiredRecords() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if a.db != nil {
				a.db.Exec(
					"DELETE FROM clipboard_records WHERE is_favorited = 0 AND expires_at IS NOT NULL AND expires_at < ?",
					time.Now(),
				)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) AddRecord(content string) error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	if strings.TrimSpace(content) == "" {
		return nil
	}

	var expiresAt *time.Time
	if a.config.SaveDays > 0 {
		exp := time.Now().AddDate(0, 0, a.config.SaveDays)
		expiresAt = &exp
	}

	_, err := a.db.Exec(
		"INSERT INTO clipboard_records (content, expires_at) VALUES (?, ?)",
		content, expiresAt,
	)
	return err
}

func (a *App) RecordFromPaste(content string) {
	if content == "" {
		return
	}

	if err := a.AddRecord(content); err != nil {
		a.log(fmt.Sprintf("粘贴记录失败: %v", err))
	} else {
		a.log(fmt.Sprintf("已记录粘贴内容 (长度: %d)", len(content)))
	}
}

func (a *App) GetRecords(search string) ([]ClipboardRecord, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	query := "SELECT id, content, created_at, is_favorited, expires_at FROM clipboard_records"
	args := []interface{}{}

	if search != "" {
		query += " WHERE content LIKE ?"
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY is_favorited DESC, created_at DESC LIMIT 200"

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ClipboardRecord
	for rows.Next() {
		var r ClipboardRecord
		var createdAtStr string
		var expiresAtStr sql.NullString

		err := rows.Scan(&r.ID, &r.Content, &createdAtStr, &r.IsFavorited, &expiresAtStr)
		if err != nil {
			continue
		}

		r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if expiresAtStr.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", expiresAtStr.String)
			r.ExpiresAt = &t
		}

		records = append(records, r)
	}

	return records, nil
}

func (a *App) GetFavorites() ([]ClipboardRecord, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	rows, err := a.db.Query(
		"SELECT id, content, created_at, is_favorited, expires_at FROM clipboard_records WHERE is_favorited = 1 ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ClipboardRecord
	for rows.Next() {
		var r ClipboardRecord
		var createdAtStr string
		var expiresAtStr sql.NullString

		err := rows.Scan(&r.ID, &r.Content, &createdAtStr, &r.IsFavorited, &expiresAtStr)
		if err != nil {
			continue
		}

		r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if expiresAtStr.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", expiresAtStr.String)
			r.ExpiresAt = &t
		}

		records = append(records, r)
	}

	return records, nil
}

func (a *App) ToggleFavorite(id int64) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("数据库未初始化")
	}

	var isFav bool
	var content string
	err := a.db.QueryRow("SELECT is_favorited, content FROM clipboard_records WHERE id = ?", id).Scan(&isFav, &content)
	if err != nil {
		return "", err
	}

	if !isFav {
		var count int
		err := a.db.QueryRow("SELECT COUNT(*) FROM clipboard_records WHERE content = ? AND is_favorited = 1 AND id != ?", content, id).Scan(&count)
		if err != nil {
			return "", err
		}
		if count > 0 {
			return "该内容已收藏过", nil
		}
	}

	_, err = a.db.Exec(
		"UPDATE clipboard_records SET is_favorited = NOT is_favorited, expires_at = CASE WHEN NOT is_favorited THEN NULL ELSE expires_at END WHERE id = ?",
		id,
	)
	if err != nil {
		return "", err
	}

	if isFav {
		return "已取消收藏", nil
	}
	return "已收藏", nil
}

func (a *App) DeleteRecord(id int64) error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := a.db.Exec("DELETE FROM clipboard_records WHERE id = ?", id)
	return err
}

func (a *App) ClearAllRecords() error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := a.db.Exec("DELETE FROM clipboard_records WHERE is_favorited = 0")
	return err
}

func (a *App) GetConfig() *Config {
	return a.config
}

func (a *App) UpdateConfig(hotkey string, saveDays int, storagePath string) error {
	if hotkey != "" {
		a.config.Hotkey = hotkey
	}
	if saveDays > 0 {
		a.config.SaveDays = saveDays
	}
	if storagePath != "" {
		a.config.StoragePath = storagePath
	}

	a.saveConfig()
	return nil
}

func (a *App) SelectDirectory() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择数据存储位置",
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) CopyToClipboard(content string) error {
	defer func() {
		if r := recover(); r != nil {
			a.log(fmt.Sprintf("PANIC in CopyToClipboard: %v", r))
		}
	}()

	if a.monitor != nil {
		a.monitor.SetClipboard(content)
	}
	return nil
}
