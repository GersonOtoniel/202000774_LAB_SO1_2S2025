import math
import time

def cpu_heavy_task():
    end_time = time.time() + 60  # 60 segundos
    while time.time() < end_time:
        # Mucho cálculo matemático para usar CPU
        [math.sqrt(i) for i in range(10**6)]

if __name__ == "__main__":
    print("Ejecutando carga PESADA de CPU por 60s...")
    cpu_heavy_task()
    print("Finalizado CPU pesado.")

