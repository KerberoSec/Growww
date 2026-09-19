"""
Device Fingerprinting, Geofencing & Suspicious Login Detection (Prompt 074)
Detects concurrent geographically impossible logins, new hardware fingerprints, and VPN anomalies.
"""

import math
from typing import Dict, Any, List
from datetime import datetime

def haversine_km(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    R = 6371.0 # Earth radius in km
    dLat = math.radians(lat2 - lat1)
    dLon = math.radians(lon2 - lon1)
    a = (math.sin(dLat / 2) ** 2 +
         math.cos(math.radians(lat1)) * math.cos(math.radians(lat2)) *
         math.sin(dLon / 2) ** 2)
    c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))
    return R * c

class DeviceFingerprintGuard:
    def __init__(self):
        self.user_devices: Dict[str, List[str]] = {} # user -> list of device hashes
        self.last_login_geo: Dict[str, Dict[str, Any]] = {}

    def evaluate_login(self, user_id: str, device_hash: str, ip: str, lat: float, lon: float, is_vpn: bool) -> Dict[str, Any]:
        now = datetime.utcnow()
        known_devices = self.user_devices.setdefault(user_id, [])

        is_new_device = device_hash not in known_devices
        impossible_speed = False

        if user_id in self.last_login_geo:
            prev = self.last_login_geo[user_id]
            dist_km = haversine_km(prev["lat"], prev["lon"], lat, lon)
            hours_diff = (now - prev["time"]).total_seconds() / 3600.0
            if hours_diff > 0:
                speed_kmh = dist_km / hours_diff
                if speed_kmh > 900.0 and dist_km > 100.0: # Exceeds commercial flight speed
                    impossible_speed = True

        self.last_login_geo[user_id] = {"lat": lat, "lon": lon, "time": now}
        if not is_new_device:
            known_devices.append(device_hash)

        requires_2fa = is_new_device or impossible_speed or is_vpn
        status = "FLAGGED_SUSPICIOUS" if impossible_speed else ("STEP_UP_CHALLENGE" if requires_2fa else "ALLOWED")

        return {
            "status": status,
            "is_new_device": is_new_device,
            "impossible_travel": impossible_speed,
            "vpn_detected": is_vpn,
            "challenge_required": requires_2fa
        }
