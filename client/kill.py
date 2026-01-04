import os
import sys
import subprocess
import platform
import shutil
import time

# Данные должны совпадать с теми, что в клиенте
REG_NAME = "WinDefenderUpdater"
EXE_NAME = "WinDefenderSmart.exe"
FOLDER_NAME = "DefenderUpdate"

def kill_process(name):
    """Принудительное завершение процесса по имени"""
    try:
        if platform.system() == "Windows":
            # Используем taskkill без окна консоли
            subprocess.run(f"taskkill /F /IM {name} /T", 
                           shell=True, 
                           capture_output=True, 
                           creationflags=0x08000000) # CREATE_NO_WINDOW
        else:
            subprocess.run(f"pkill -f {name}", shell=True)
        print(f"[*] Process {name} terminated.")
    except Exception as e:
        print(f"[!] Error killing process: {e}")

def remove_from_startup():
    """Удаление из реестра автозагрузки"""
    if platform.system() == "Windows":
        try:
            import winreg
            key = winreg.OpenKey(winreg.HKEY_CURRENT_USER, 
                                 r"Software\Microsoft\Windows\CurrentVersion\Run", 
                                 0, winreg.KEY_SET_VALUE)
            winreg.DeleteValue(key, REG_NAME)
            winreg.CloseKey(key)
            print("[*] Registry startup key removed.")
        except FileNotFoundError:
            print("[?] Registry key already gone.")
        except Exception as e:
            print(f"[!] Registry error: {e}")

def wipe_files():
    """Полное удаление папки с файлами"""
    app_data = os.getenv("APPDATA")
    target_dir = os.path.join(app_data, "Microsoft", "Windows", FOLDER_NAME)
    
    if os.path.exists(target_dir):
        try:
            # Небольшая пауза, чтобы процесс успел закрыться
            time.sleep(1)
            shutil.rmtree(target_dir)
            print(f"[*] Directory {target_dir} wiped.")
        except Exception as e:
            print(f"[!] File removal error: {e}")
    else:
        print("[?] Directory not found.")

def main():
    print("--- PROJECT ZENITH CLEANER ---")
    
    # 1. Останавливаем работу программы
    kill_process(EXE_NAME)
    kill_process("pythonw.exe") # Если запущен как .pyw
    
    # 2. Чистим автозагрузку
    remove_from_startup()
    
    # 3. Удаляем файлы
    wipe_files()
    
    print("[+] Cleanup complete. Zenith removed.")
    
    # Самоликвидация (опционально)
    # os.remove(sys.argv[0]) 

if __name__ == "__main__":
    main()