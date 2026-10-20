import os

import psycopg2
from flask import Flask, jsonify, request
from psycopg2.extras import RealDictCursor

app = Flask(__name__)

DATABASE_URL = os.getenv(
    "DATABASE_URL", "postgresql://postgres:postgres@telemetry-db:5432/telemetry"
)


def connect():
    return psycopg2.connect(DATABASE_URL)


def point_to_json(row: dict) -> dict:
    return {
        "id": row["id"],
        "deviceId": row["device_id"],
        "value": row["value"],
        "unit": row["unit"],
        "recordedAt": row["recorded_at"].isoformat(),
    }


@app.get("/health")
def health():
    try:
        with connect() as conn, conn.cursor() as cur:
            cur.execute("SELECT 1")
    except psycopg2.Error as err:
        return jsonify({"status": "error", "detail": str(err)}), 503
    return jsonify({"status": "ok"})


@app.post("/telemetry")
def ingest():
    payload = request.get_json(silent=True) or {}
    device_id = payload.get("deviceId")
    value = payload.get("value")

    if not device_id or value is None:
        return jsonify({"code": "validation_error", "message": "deviceId и value обязательны"}), 400

    unit = payload.get("unit") or "°C"

    with connect() as conn, conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute(
            """
            INSERT INTO telemetry (device_id, value, unit)
            VALUES (%s, %s, %s)
            RETURNING id, device_id, value, unit, recorded_at
            """,
            (str(device_id), float(value), unit),
        )
        row = cur.fetchone()

    return jsonify(point_to_json(row)), 201


@app.get("/telemetry")
def list_points():
    device_id = request.args.get("deviceId")
    if not device_id:
        return jsonify({"code": "validation_error", "message": "deviceId обязателен"}), 400

    try:
        limit = int(request.args.get("limit", 100))
    except ValueError:
        return jsonify({"code": "validation_error", "message": "limit должен быть числом"}), 400
    limit = max(1, min(limit, 1000))

    with connect() as conn, conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute(
            """
            SELECT id, device_id, value, unit, recorded_at
            FROM telemetry
            WHERE device_id = %s
            ORDER BY recorded_at DESC
            LIMIT %s
            """,
            (str(device_id), limit),
        )
        rows = cur.fetchall()

    return jsonify([point_to_json(row) for row in rows])


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", "8083")))
