import math
import time

def cpu_light_task():
    print("Ejecutando carga LIGERA de CPU por 30s...")
    end_time = time.time() + 30
    while time.time() < end_time:
        [math.sqrt(i) for i in range(10**4)]  # menos operaciones

if __name__ == "__main__":
    cpu_light_task()
    print("Finalizado CPU ligero.")

