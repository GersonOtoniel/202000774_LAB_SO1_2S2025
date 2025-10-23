import time

start_time = time.time()
while time.time() - start_time < 3600:  # 1 hora
    # Carga moderada de CPU
    [x**2 for x in range(1000000)]
    time.sleep(0.1)  # pausa corta para no saturar CPU

