#!/bin/bash

# ===============================
# Configuración de límites
# ===============================
MAX_CPU=70      # máximo % de CPU que queremos usar
MAX_RAM=70      # máximo % de RAM que queremos usar

# Limites de contenedores
PESADOS_CPU=0.6
PESADOS_RAM=256m
LIGEROS_CPU=0.2
LIGEROS_RAM=128m

# Imágenes
pesados=("high_cpu" "high_ram")
ligeros=("low_cpu" "low_ram")

# Contadores
total=0
MAX_CONTAINERS=10

# Funciones para obtener % de uso actual
get_cpu_usage() {
    # suma de todos los CPU actuales en %
    echo $(top -bn1 | grep "Cpu(s)" | awk '{print 100-$8}')
}

get_ram_usage() {
    free | awk '/Mem/ {printf "%.0f", $3/$2*100}'
}

echo "🚀 Iniciando hasta $MAX_CONTAINERS contenedores sin pasar límites de CPU=$MAX_CPU% y RAM=$MAX_RAM%..."
echo

while [ $total -lt $MAX_CONTAINERS ]; do

    # obtener uso actual
    cpu_now=$(get_cpu_usage)
    ram_now=$(get_ram_usage)

    # Elegir tipo contenedor aleatorio
    if (( RANDOM % 2 )); then
        tipo_contenedor=${pesados[$((RANDOM % ${#pesados[@]}))]}
        cpu_limit=$PESADOS_CPU
        ram_limit=$PESADOS_RAM
        tipo="PESADO"
    else
        tipo_contenedor=${ligeros[$((RANDOM % ${#ligeros[@]}))]}
        cpu_limit=$LIGEROS_CPU
        ram_limit=$LIGEROS_RAM
        tipo="LIGERO"
    fi

    # Verificar que no superemos límites
    cpu_needed=$(echo "$cpu_now + $cpu_limit*100" | bc)
    ram_needed=$(echo "$ram_now + ${ram_limit%m}*100/$(free -m | awk '/Mem/ {print $2}')" | bc)

    if (( $(echo "$cpu_needed < $MAX_CPU" | bc) )) && (( $(echo "$ram_needed < $MAX_RAM" | bc) )); then
        # Generar nombre aleatorio
        random_name=$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 6 | head -n 1)
        current_time=$(date +"%H%M%S")
        container_name="${tipo_contenedor//:/-}-${total}-${random_name}-${current_time}"

        docker run -d --name "$container_name" --cpus="$cpu_limit" --memory="$ram_limit" "$tipo_contenedor"

        echo "✅ Contenedor $tipo No.$total -> $container_name ($tipo_contenedor)"
        total=$((total+1))
    else
        echo "⚠️ No se puede lanzar $tipo ahora. CPU: $cpu_now%, RAM: $ram_now%. Esperando..."
        sleep 2
    fi
done

echo
echo "🎉 Todos los contenedores creados. Verifica con: docker ps"

