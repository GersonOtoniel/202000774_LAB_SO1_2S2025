import time

def ram_light_task():
    print("Reservando poca RAM por 30s...")
    data = []
    end_time = time.time() + 30
    while time.time() < end_time:
        data.append("X" * 1024 * 100)  # ~100KB cada vez
        time.sleep(0.2)

if __name__ == "__main__":
    ram_light_task()
    print("Finalizado RAM ligero.")

