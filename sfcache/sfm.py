import re
import pymem
import pymem.process
import time
import sqlite3
import requests
import signal
import sys
import toml


class MemoryScanner:
    def __init__(self, config_file='config.toml'):
        config = toml.load(config_file)
        self.host = config['server']['host']
        self.api_key = config['server']['api_key']
        self.conn = sqlite3.connect(config['database']['path'])
        self.mode = config['mode']['operation_mode']
        self.cursor = self.conn.cursor()
        self.added_values = set()
        self.running = True
        self.pattern = re.compile(r"([0-9a-f]{32}):([0-9a-f]{32})")

        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS cached (
                kid TEXT PRIMARY KEY,
                key TEXT NOT NULL
            )
        ''')
        self.conn.commit()

    def insert_key(self, kid, key):
        try:
            self.cursor.execute(
                'INSERT OR REPLACE INTO cached (kid, key) VALUES (?, ?)',
                (kid, key)
            )
            self.conn.commit()
            print(f"Successfully inserted {kid} {key}")
        except sqlite3.Error as e:
            print(f"Error inserting {kid} {key}: {e}")

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
        except requests.exceptions.ConnectionError:
            print("Connection error: Unable to reach the server.")
        except requests.exceptions.HTTPError:
            error_message = response.json().get('error', 'Unknown error')
            print(f"HTTP error: {error_message}")
        except requests.RequestException as e:
            print(f"Request error: {e}")

    def scan_memory(self):
        process_name = "StreamFab64.exe"
        try:
            pm = pymem.Pymem(process_name)
            address = 0x00000000
            max_address = 0x7FFFFFFF  
            pattern = r"([0-9a-f]{32}):([0-9a-f]{32})"
            while address < max_address:
                    try:
                        memory_chunk = pm.read_bytes(address, 0x1000)  
                        data = memory_chunk.decode('utf-8', errors='ignore').strip()
                        matches = re.findall(pattern, data)
                        for match in matches:
                            extracted_value = f"{match[0]}:{match[1]}"
                            if extracted_value.startswith("00000") and extracted_value not in self.added_values:
                                print(f"Match found at {hex(address)}: {extracted_value}")
                                kid, key = extracted_value.split(':')
                                self.insert_key(kid, key)
                                self.send_to_remote(kid, key)
                                self.added_values.add(extracted_value)
                    except (pymem.exception.MemoryReadError, pymem.exception.MemoryWriteError):
                        pass 

                    address += 0x1000  

        except Exception as e:
            print(f"An error occurred during memory scan: {e}")
        finally:
            if 'pm' in locals():
                pm.close_process()

    def run(self):
        while self.running:
            self.scan_memory()
            time.sleep(1)  

    def stop(self):
        self.running = False
        self.conn.close()

def signal_handler(signum, frame):
    print("Exiting gracefully...")
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
