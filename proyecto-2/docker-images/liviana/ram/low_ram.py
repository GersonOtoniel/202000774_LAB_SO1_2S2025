import time

data = []
start_time = time.time()
while time.time() - start_time < 3600:  # 1 hora
    # Ocupa RAM mínima
    data.append(' ' * 1024 * 5)  # 5 KB por iteración
    if len(data) > 20:  # máximo ~100 KB
        data.pop(0)
    time.sleep(0.5)

