import time

start_time = time.time()
while time.time() - start_time < 3600:  # 1 hora
    # Muy ligera
    [x**2 for x in range(10000)]  # carga mínima
    time.sleep(0.5)

