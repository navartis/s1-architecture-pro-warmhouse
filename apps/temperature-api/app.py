import os
import random
from datetime import datetime, timezone

from flask import Flask, jsonify, request

app = Flask(__name__)

LOCATION_BY_SENSOR_ID = {
    "1": "Living Room",
    "2": "Bedroom",
    "3": "Kitchen",
}
SENSOR_ID_BY_LOCATION = {location: sensor_id for sensor_id, location in LOCATION_BY_SENSOR_ID.items()}

UNIT = "°C"


def reading(location: str, sensor_id: str) -> dict:
    return {
        "value": round(random.uniform(18.0, 25.0), 1),
        "unit": UNIT,
        "timestamp": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "location": location,
        "status": "active",
        "sensor_id": sensor_id,
        "sensor_type": "temperature",
        "description": f"Датчик температуры: {location}",
    }


@app.get("/health")
def health():
    return jsonify({"status": "ok"})


@app.get("/temperature")
def temperature_by_location():
    location = request.args.get("location", "").strip()
    if not location:
        return jsonify({"error": "location is required"}), 400

    sensor_id = SENSOR_ID_BY_LOCATION.get(location, "0")
    return jsonify(reading(location, sensor_id))


@app.get("/temperature/<sensor_id>")
def temperature_by_sensor(sensor_id: str):
    location = LOCATION_BY_SENSOR_ID.get(sensor_id, "Unknown")
    return jsonify(reading(location, sensor_id))


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", "8081")))
