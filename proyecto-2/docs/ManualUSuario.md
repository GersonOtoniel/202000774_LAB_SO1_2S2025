# Manual de Usuario

## Instrucciones claras para el usuario final
Este sistema permite monitorear en tiempo real el consumo de **CPU** y **Memoria RAM** de los procesos en el sistema.  
La información es recolectada por un **módulo del kernel en C**, almacenada en una base de datos SQLite mediante un **daemon en Go**, y visualizada en un **dashboard de Grafana**.

### Pasos básicos para el usuario:
1. Asegurarsee de que el módulo del kernel esté cargado.
2. Verifique que el daemon de Go esté ejecutándose en segundo plano.
3. Abra Grafana en su navegador para acceder al dashboard.

---

## Ejemplos de uso práctico
- **Visualizar memoria usada y libre**: El dashboard muestra métricas de uso de RAM en MB.  
- **Monitorear procesos**: Cada proceso con su PID, uso de CPU y memoria aparece registrado.  
- **Análisis histórico**: Puede ver la evolución del consumo de recursos a lo largo del tiempo gracias a los datos guardados en SQLite.

---

# Guía de Instalación

## Requisitos del sistema
- **Sistema Operativo**: Linux (probado en Ubuntu/Archcraft).  
- **Kernel**: Compatible con carga de módulos (>= 5.x).  
- **Docker**: Versión 27 o superior.  
- **Go**: Versión 1.21 o superior (para el daemon).  

## Pasos detallados de instalación y configuración
1. **Compilar e instalar el módulo del kernel**:
   ```bash
   make
   sudo insmod continfo.ko
    ```
2. **Ejecutar el daemon en Go**:
    ```bash
    go run daemon.go
    ```

Este se encargará de leer /proc/continfo_so1_202000774 y guardar la información en containers.db.


3. **Levantar Grafana con Docker**:
    ```bash
    docker-compose up --build -d
    ```

Acceder a Grafana desde el navegador en http://localhost:3000.

## Diagramas y Arquitectura
### Diagrama de Flujo

```
[Módulo Kernel en C] ---> [/proc/continfo_so1_#CARNET]
                                |
                                v
                        [Daemon en Go]
                                |
                                v
                    [Base de Datos SQLite]
                                |
                                v
                      [Grafana Dashboard]

```

### Arquitectura del sistema

- Módulo Kernel en C: Recolecta métricas de procesos (RAM, CPU) y expone la información en /proc.

- Daemon en Go: Lee periódicamente la información del módulo y la guarda en containers.db (SQLite).

- Grafana: Usando un datasource SQLite, presenta gráficos en tiempo real con la información registrada.

- Esta arquitectura permite que usuarios no técnicos solo tengan que abrir Grafana para visualizar métricas sin preocuparse por la complejidad del backend.