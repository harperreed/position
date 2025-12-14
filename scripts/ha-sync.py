#!/usr/bin/env python3
# ABOUTME: Syncs Home Assistant entity locations to position
# ABOUTME: Polls HA API and calls position CLI to record locations

# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "requests",
#     "python-dotenv",
# ]
# ///

"""Sync Home Assistant entity locations to position."""

import os
import subprocess
import sys
from pathlib import Path

import requests
from dotenv import load_dotenv

# Entity ID -> position name mapping
# Edit this to match your Home Assistant entities
ENTITIES = {
    "person.harper": "harper",
    # "device_tracker.harpers_iphone": "harper-phone",
    # "device_tracker.model_3": "car",
}

# Path to position binary (adjust if needed)
POSITION_BIN = "position"


def get_entity_state(hass_url: str, hass_token: str, entity_id: str) -> dict | None:
    """Fetch entity state from Home Assistant."""
    url = f"{hass_url}/api/states/{entity_id}"
    headers = {"Authorization": f"Bearer {hass_token}"}

    try:
        resp = requests.get(url, headers=headers, timeout=10)
        if resp.status_code != 200:
            print(f"Failed to fetch {entity_id}: HTTP {resp.status_code}")
            return None
        return resp.json()
    except requests.RequestException as e:
        print(f"Request error for {entity_id}: {e}")
        return None


def sync_entity(hass_url: str, hass_token: str, entity_id: str, position_name: str) -> bool:
    """Sync a single entity to position. Returns True on success."""
    state = get_entity_state(hass_url, hass_token, entity_id)
    if not state:
        return False

    attrs = state.get("attributes", {})
    lat = attrs.get("latitude")
    lng = attrs.get("longitude")

    if lat is None or lng is None:
        print(f"No location for {entity_id} (state: {state.get('state', 'unknown')})")
        return False

    label = attrs.get("friendly_name") or entity_id

    try:
        subprocess.run(
            [
                POSITION_BIN,
                "add",
                position_name,
                "--lat",
                str(lat),
                "--lng",
                str(lng),
                "--label",
                label,
            ],
            check=True,
            capture_output=True,
            text=True,
        )
        print(f"Synced {position_name}: {label} ({lat}, {lng})")
        return True
    except subprocess.CalledProcessError as e:
        print(f"Failed to add position for {position_name}: {e.stderr}")
        return False
    except FileNotFoundError:
        print(f"position binary not found: {POSITION_BIN}")
        return False


def main() -> int:
    """Main entry point. Returns exit code."""
    # Load .env from script directory or current directory
    script_dir = Path(__file__).parent
    load_dotenv(script_dir / ".env")
    load_dotenv()  # Also check current directory

    hass_url = os.environ.get("HASS_URL")
    hass_token = os.environ.get("HASS_TOKEN")

    if not hass_url or not hass_token:
        print("Error: HASS_URL and HASS_TOKEN must be set in .env or environment")
        return 1

    # Strip trailing slash from URL
    hass_url = hass_url.rstrip("/")

    success_count = 0
    for entity_id, position_name in ENTITIES.items():
        if sync_entity(hass_url, hass_token, entity_id, position_name):
            success_count += 1

    print(f"Synced {success_count}/{len(ENTITIES)} entities")
    return 0 if success_count == len(ENTITIES) else 1


if __name__ == "__main__":
    sys.exit(main())
