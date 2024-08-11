import ctypes
import string
import psutil
import time
import re
import sqlite3
import requests
import signal
import sys
import concurrent.futures
import toml

def find_pid_by_name(process_name):
    for proc in psutil.process_iter(['pid', 'name']):
        if proc.info['name'] == process_name:
            return proc.info['pid']
    return None

def read_memory(process_handle, address, size):
    buffer = ctypes.create_string_buffer(size)
    bytes_read = ctypes.c_ulong(0)
    if ctypes.windll.kernel32.ReadProcessMemory(process_handle, address, buffer, size, ctypes.byref(bytes_read)):
        return buffer.raw.decode('utf-8', errors='ignore').strip()
    return None

class MemoryScanner:
    def __init__(self,config_file='config.toml'):
        config = toml.load(config_file)
        self.host = config['server']['host']
        self.api_key = config['server']['api_key']
        self.conn = sqlite3.connect(config['database']['path'])
        self.mode = config['mode']['operation_mode']
        self.conn = sqlite3.connect('netflix.db')
        self.cursor = self.conn.cursor()
        self.added_values = set()
        self.running = True

        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS cached (
                kid TEXT PRIMARY KEY,
                key TEXT NOT NULL
            )
        ''')
        self.conn.commit()

    def insert_key(self, kid, key):
        self.cursor.execute('INSERT OR REPLACE INTO cached (kid, key) VALUES (?, ?)', (kid, key))
        self.conn.commit()

    def send_to_remote(self, kid, key):
        if self.mode == 1:
            return

        url = f"{self.host}/api/keys"
        headers = {
            "Content-Type": "application/json",
            "Cookie": f"api-key={self.api_key}"
        }
        payload = {"kid": kid, "key": key}

        try:
            response = requests.post(url, json=payload, headers=headers)
            response.raise_for_status()
        except requests.RequestException as e:
            print(f"Error: {response.json()["error"]}")

    def scan_memory(self, pid):
        process_handle = ctypes.windll.kernel32.OpenProcess(0x10 | 0x20, False, pid)
        if not process_handle:
            print(f"Failed to open process {pid}.")
            return

        try:
            with concurrent.futures.ThreadPoolExecutor() as executor:
                for address in range(0x000000000, 0x7FFFFFFF, 4096):
                    if not self.running:
                        break
                    data = read_memory(process_handle, address, 512)
                    if data:
                        match = re.search(r"([0-9a-f]{32}):([0-9a-f]{32})", data)
                        if match:
                            extracted_value = match.group(0)
                            if extracted_value.startswith("00000"):
                                if extracted_value not in self.added_values:
                                    print("Match found:", extracted_value)
                                    kid, key = extracted_value.split(':')
                                    future_insert = executor.submit(self.insert_key, kid, key)
                                    future_send = executor.submit(self.send_to_remote, kid, key)
                                    concurrent.futures.wait([future_insert, future_send])
                                    self.added_values.add(extracted_value)

        finally:
            ctypes.windll.kernel32.CloseHandle(process_handle)

    def run(self):
        process_name = "StreamFab64.exe"
        pid = find_pid_by_name(process_name)

        if not pid:
            print(f"Process {process_name} not found.")
            return

        print(f"Found PID: {pid}")
        while self.running:
            self.scan_memory(pid)
            time.sleep(1)

    def stop(self):
        self.running = False
        self.conn.close()
        print("Scanning stopped.")

def signal_handler(signum, frame):
    print("Received signal to stop. Exiting gracefully...")
    scanner.stop()
    sys.exit(0)

if __name__ == "__main__":
    scanner = MemoryScanner()
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        scanner.run()
    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        scanner.stop()