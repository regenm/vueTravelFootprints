# server.py
from flask import Flask, request, jsonify
import json
import os
from flask_cors import CORS

app = Flask(__name__)
CORS(app, origins=["http://localhost:5173"])
JSON_FILE = "markers.json"

# 确保 JSON 文件存在
if not os.path.exists(JSON_FILE):
    with open(JSON_FILE, "w", encoding="utf-8") as f:
        json.dump([], f, ensure_ascii=False, indent=2)

@app.route("/add-marker", methods=["POST"])
def add_marker():
    """新增一个标记"""
    try:
        data = request.get_json()
        name = data.get("name", "").strip()
        longitude = str(data.get("longitude", "")).strip()
        latitude = str(data.get("latitude", "")).strip()
        photos = data.get("photos", [])

        if not name or not longitude or not latitude:
            return jsonify({"success": False, "message": "名称或坐标不能为空"})

        # 读取现有数据
        with open(JSON_FILE, "r", encoding="utf-8") as f:
            markers = json.load(f)

        # 添加新数据
        marker = {
            "name": name,
            "longitude": longitude,
            "latitude": latitude,
            "photos": photos
        }
        markers.append(marker)

        # 写回 JSON 文件
        with open(JSON_FILE, "w", encoding="utf-8") as f:
            json.dump(markers, f, ensure_ascii=False, indent=2)

        # 返回所有数据
        return jsonify({"success": True, "data": markers})

    except Exception as e:
        return jsonify({"success": False, "message": str(e)})
@app.route("/markers", methods=["GET"])
def get_markers():
    try:
        with open(JSON_FILE, "r", encoding="utf-8") as f:
            markers = json.load(f)
        return jsonify({"success": True, "data": markers})
    except Exception as e:
        return jsonify({"success": False, "message": str(e)})


if __name__ == "__main__":
    app.run(port=5000, debug=True)
