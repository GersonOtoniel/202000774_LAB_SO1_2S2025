import time

def ram_heavy_task():
    print("Reservando RAM intensivamente por 60s...")
    data = []
    end_time = time.time() + 60
    while time.time() < end_time:
        # Cada iteración agrega ~1MB
        data.append("X" * 1024 * 1024)
        time.sleep(0.1)

if __name__ == "__main__":
    ram_heavy_task()
    print("Finalizado RAM pesado.")

