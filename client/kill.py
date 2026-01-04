import os
import subprocess
import shutil
import time

# --- ТЕ ЖЕ НАСТРОЙКИ, ЧТО В АГЕНТЕ ---
EXE_NAME = "WinDefenderSmart.exe"
REG_NAME = "WinDefenderUpdate"
# ------------------------------------

def kill_process(name):
    print(f"[*] Попытка завершить процесс {name}...")
    try:
        # Принудительно завершаем дерево процессов
        subprocess.run(['taskkill', '/F', '/IM', name, '/T'], capture_output=True)
        print("[+] Процесс остановлен.")
    except Exception as e:
        print(f"[-] Ошибка при остановке: {e}")

def remove_from_registry(reg_name):
    print(f"[*] Удаление из автозагрузки...")
    reg_path = r"HKCU\Software\Microsoft\Windows\CurrentVersion\Run"
    try:
        subprocess.run(['reg', 'delete', reg_path, '/v', reg_name, '/f'], capture_output=True)
        print("[+] Запись в реестре удалена.")
    except Exception as e:
        print(f"[-] Ошибка реестра: {e}")

def remove_files():
    appdata = os.getenv("APPDATA")
    target_dir = os.path.join(appdata, "Microsoft", "Windows", "Defender")
    
    print(f"[*] Очистка файлов в {target_dir}...")
    if os.path.exists(target_dir):
        try:
            # Ждем секунду, чтобы дескрипторы файлов освободились после taskkill
            time.sleep(1)
            shutil.rmtree(target_dir)
            print("[+] Файлы агента полностью удалены.")
        except Exception as e:
            print(f"[-] Ошибка при удалении папки: {e}")
    else:
        print("[!] Папка агента не найдена.")

if __name__ == "__main__":
    print("=== ZENITH CLEANER TOOL ===")
    kill_process(EXE_NAME)
    remove_from_registry(REG_NAME)
    remove_files()
    print("="*27)
    print("[УСПЕХ] Система очищена.")
    input("Нажми Enter, чтобы выйти...")