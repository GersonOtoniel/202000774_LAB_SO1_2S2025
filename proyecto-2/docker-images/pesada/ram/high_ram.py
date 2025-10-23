import time

data = []
start_time = time.time()
while time.time() - start_time < 3600:  # 1 hora
    # Ocupa RAM progresivamente pero limitado
    data.append(' ' * 1024 * 50)  # 50 KB por iteración
    if len(data) > 50:  # máximo ~2.5 MB para no saturar
        data.pop(0)
    time.sleep(0.1)

