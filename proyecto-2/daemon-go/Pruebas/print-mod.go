package main

import (
	"bufio"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

// Estructura para parsear el JSON
type Process struct {
    PID         int     `json:"PID"`
    Name        string  `json:"Name"`
    Cmdline     string  `json:"Cmdline"`
    Vsz         int64   `json:"vsz"`
    Rss         int64   `json:"rss"`
    MemoryUsage float64 `json:"Memory_Usage"`
    CPUUsage    float64 `json:"CPU_Usage"`
}

type ProcessList struct {
    Processes []Process `json:"Processes"`
}

type ContInfo struct {
    PID         int
    Name        string
    VSZ         int64
    RSS         int64
    Cmd         string
    MemoryUsage float64
    CPUUsage    float64
}

type ProcessSys struct {
	PID int
	Name string
	RSSMB int64
	VSZMB int64
	State string
	CmdLine string
	CPUUsage float64
}

var (
	contProcPath = flag.String("contProc", "/proc/continfo_so1_202000774", "proc cont info" )
	sysProcPath = flag.String("sysProc", "/proc/sysinfo_so1_202000774", "proc sys info")
	pollInterval = flag.Duration("interval", 10*time.Second, "Polling Interval")
	cronJobScript = flag.String("cronScript", "/home/gerson/Escritorio/Sopes1/LAB/PROYECTO_1/proyecto-2/cron/run_containers.sh", "Path to container creation script")
	cronJobSchedule = flag.String("cronSchedule", "* * * * * /home/gerson/Escritorio/Sopes1/LAB/PROYECTO_1/proyecto-2/cron/run_containers.sh ", "Cron schedule for container creation")
	pathDatabase = flag.String("database", "/home/gerson/Escritorio/Sopes1/LAB/PROYECTO_1/proyecto-2/grafana/database.db", "database")
)

/*var (
	cronJobActive = prometheus.NewGauge(prometheus.GaugeOpts{Name: "cronjob_active", Help: "Cronjob status (1=active, 0=inactive)"})
)*/


func ensureDB(db *sql.DB) error {
		tables := []string{`
		CREATE TABLE IF NOT EXISTS SysInfo(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at  INTEGER,
			total_ram REAL,
			free_ram REAL,
			used_ram REAL,
			total_process INTEGER
		);`,	

		`CREATE TABLE IF NOT EXISTS ContInfo(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at INTEGER,
			total_ram REAL,
			free_ram REAL,
			used_ram REAL,
			total_cont INTEGER, 
			cpu REAL, 
			cmd TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS TopRamSys(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			pid REAL,
			rss_mb REAL,
			created_at INTEGER

		);`,

		`CREATE TABLE IF NOT EXISTS TopCpuSys(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			pid REAL,
			cpu REAL,
			created_at INTEGER
		);`,

		`CREATE TABLE IF NOT EXISTS TopRamCont(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			pid REAL,
			rss REAL,
			created_at INTEGER
		)`,

		`CREATE TABLE IF NOT EXISTS TopCpuCont(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			pid REAL, 
			cpu REAL,
			created_at INTEGER
		)`,
		}
		for _, t := range tables {
			if _, err := db.Exec(t); err != nil {
				return err
			}
		}
		return nil
}

// Función para agregar el cronjob
func setupCronJob() error {
    //cronEntry := fmt.Sprintf("%s %s >> /var/log/container-creator.log 2>&1\n", 
      //  *cronJobSchedule, *cronJobScript)
    cronEntry := fmt.Sprintf("%s\n", *cronJobSchedule)

    // Leer crontab actual
    cmd := exec.Command("crontab", "-l")
    currentCrontab, err := cmd.Output()
    if err != nil && err.Error() != "exit status 1" {
        return fmt.Errorf("error reading crontab: %v", err)
    }
    
    // Verificar si el cronjob ya existe
    if strings.Contains(string(currentCrontab), *cronJobScript) {
        log.Println("Cronjob already exists")
        return nil
    }
    
    // Agregar nuevo cronjob
    newCrontab := string(currentCrontab) + cronEntry
    
    // Escribir nuevo crontab
    cmd = exec.Command("crontab", "-")
    cmd.Stdin = strings.NewReader(newCrontab)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error writing crontab: %v", err)
    }
    
    log.Printf("Cronjob added: %s", cronEntry)
    //cronJobActive.Set(1)
    return nil
}


// Función para eliminar el cronjob
func removeCronJob() error {
    // Leer crontab actual
    cmd := exec.Command("crontab", "-l")
    currentCrontab, err := cmd.Output()
    if err != nil {
        if err.Error() == "exit status 1" {
            // Crontab vacío, nada que remover
            return nil
        }
        return fmt.Errorf("error reading crontab: %v", err)
    }
    
    // Filtrar nuestra entrada del cronjob
    lines := strings.Split(string(currentCrontab), "\n")
    var newLines []string
    
    for _, line := range lines {
        if !strings.Contains(line, *cronJobScript) && strings.TrimSpace(line) != "" {
            newLines = append(newLines, line)
        }
    }
    
    // Reconstruir crontab sin nuestra entrada
    newCrontab := strings.Join(newLines, "\n")
    if newCrontab != "" {
        newCrontab += "\n"
    }
    
    // Escribir nuevo crontab
    cmd = exec.Command("crontab", "-")
    cmd.Stdin = strings.NewReader(newCrontab)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error writing crontab: %v", err)
    }
    
    log.Println("Cronjob removed successfully")
    //cronJobActive.Set(0)
    return nil
}



func parseContProcJSON(path string) ([]ContInfo, error) {
    // Leer el archivo
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    // Parsear JSON
    var processList ProcessList
    err = json.Unmarshal(data, &processList)
    if err != nil {
        return nil, err
    }

    // Convertir a la estructura ContInfo incluyendo CPU
    var out []ContInfo
    for _, p := range processList.Processes {
        ci := ContInfo{
            PID:         p.PID,
            Name:        p.Name,
            VSZ:         p.Vsz,
            RSS:         p.Rss,
            Cmd:         p.Cmdline,
            MemoryUsage: p.MemoryUsage,
            CPUUsage:    p.CPUUsage,
        }
        out = append(out, ci)
    }

    return out, nil
}

func parseSysProc(path string) ([]ProcessSys, int64, int64, int64, error) {
		
	  var processes []ProcessSys
		var totalRAM, freeRAM, usedRAM int64
    f, err := os.Open(path)
    if err != nil {
        return nil, 0, 0, 0, err
    }
    defer f.Close()
    
    scanner := bufio.NewScanner(f)
		inProcessSection := false
		
    for scanner.Scan() {
        line := scanner.Text()
				line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "Total_RAM_MB:") {
            s := strings.TrimSpace(strings.TrimPrefix(line, "Total_RAM_MB:"))
						totalRAM, _ = strconv.ParseInt(s, 10, 64)
        } else if strings.HasPrefix(line, "Free_RAM_MB:") {
					  s := strings.TrimSpace(strings.TrimPrefix(line, "Free_RAM_MB:"))
						freeRAM, _ = strconv.ParseInt(s, 10, 64)
        } else if strings.HasPrefix(line, "Used_RAM_MB:") {
            s := strings.TrimSpace(strings.TrimPrefix(line, "Used_RAM_MB:"))
						usedRAM, _ = strconv.ParseInt(s, 10, 64)
        }
        
				if strings.Contains(line, "--- Procesos del sistema ---"){
					inProcessSection = true;
					continue
				}

				if inProcessSection && strings.HasPrefix(line, "PID:"){
					process, err := parseProcessLine(line)
					if err == nil {
						processes = append(processes, process)
					}
				}
    }
		fmt.Printf("DEBUG: Termine de leer sysinfo, procesos: %d\n", len(processes))

    return processes, totalRAM, freeRAM, usedRAM, scanner.Err()
}

func parseProcessLine(line string)(ProcessSys, error){
	var p ProcessSys

	fields := strings.Fields(line)

	for _, field := range fields {
		if strings.HasPrefix(field, "PID:"){
			pidStr := strings.TrimPrefix(field, "PID:")
			p.PID, _ = strconv.Atoi(pidStr)
		} else if strings.HasPrefix(field, "Name:"){
			p.Name = strings.TrimPrefix(field, "Name:")
		} else if strings.HasPrefix(field, "CMD:"){
			p.CmdLine = strings.TrimPrefix(field, "CMD:")
		} else if strings.HasPrefix(field, "VSZ_MB:"){
			vszStr := strings.TrimPrefix(field, "VSZ_MB:")
			p.VSZMB, _ = strconv.ParseInt(vszStr, 10, 64)
		} else if strings.HasPrefix(field, "RSS_MB:"){
			rssStr := strings.TrimPrefix(field, "RSS_MB:")
			p.RSSMB, _ = strconv.ParseInt(rssStr, 10, 64)
		} else if strings.HasPrefix(field, "State:"){
			p.State = strings.TrimPrefix(field, "State:")
		} else if strings.HasPrefix(field, "CPU_usage:"){
			cpuStr := strings.TrimPrefix(field, "CPU_usage:")
			cpuStr = strings.TrimSuffix(cpuStr, "%")
			val, err := strconv.ParseFloat(cpuStr, 64)
			if err == nil {
				p.CPUUsage = val
			}
		}
	}
	return p, nil

}


func stopAndRemoveContainer(containerID string) error {
    log.Printf("Eliminando contenedor: %s", containerID)
    
    cmd := exec.Command("docker", "rm", "-f", containerID)
    out, err := cmd.CombinedOutput()
    
    if err != nil {
        if strings.Contains(string(out), "No such container") {
            log.Printf("Contenedor %s ya no existe", containerID)
            return nil
        }
        return fmt.Errorf("docker rm -f failed: %v - out: %s", err, string(out))
    }
    
    log.Printf("Contenedor %s eliminado exitosamente", containerID)
    return nil
}

func guessContainerIDFromCmd(cmd string) string {
    // Buscar patrones de container ID en el comando
    patterns := []string{
        "-id ",    // containerd-shim pattern
        "--cidfile", // docker run --cidfile
    }
    
    for _, pattern := range patterns {
        if idx := strings.Index(cmd, pattern); idx != -1 {
            // Extraer el ID después del pattern
            remaining := cmd[idx+len(pattern):]
            fields := strings.Fields(remaining)
            if len(fields) > 0 && len(fields[0]) >= 12 && isHex(fields[0]) {
                return fields[0]
            }
        }
    }
    
    // Fallback: buscar cualquier string hexadecimal de 12+ caracteres
    parts := strings.Fields(cmd)
    for _, p := range parts {
        if len(p) >= 12 && isHex(p) {
            return p
        }
    }
    return ""
}

func isHex(s string) bool {
    for _, c := range s {
        if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
            return false
        }
    }
    return true
}

// Obtiene los contenedores Docker actuales: map[ID]Nombre
func getDockerContainers() (map[string]string, error) {
    out, err := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}").Output()
    if err != nil {
        return nil, err
    }
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    containers := make(map[string]string)
    for _, line := range lines {
        parts := strings.Fields(line)
        if len(parts) >= 2 {
            id := parts[0]
            name := parts[1]
            containers[id] = name
        }
    }
    return containers, nil
}

func enforceContainerLimit(contenedores []ContInfo, totalLimit int) {
    protectedContainers := []string{"grafana-sqlite", "grafana"}

    dockerMap, err := getDockerContainers()
    if err != nil {
        log.Println("Error obteniendo contenedores Docker:", err)
        return
    }

    var protected, nonProtected []ContInfo

    // Separar protegidos y no protegidos
    for _, cont := range contenedores {
        isProtected := false
        for _, dockerName := range dockerMap {
            for _, protectedName := range protectedContainers {
                if strings.Contains(strings.ToLower(dockerName), strings.ToLower(protectedName)) ||
                   strings.Contains(strings.ToLower(cont.Name), strings.ToLower(protectedName)) {
                    isProtected = true
                    break
                }
            }
            if isProtected {
                break
            }
        }
        if isProtected {
            protected = append(protected, cont)
        } else {
            nonProtected = append(nonProtected, cont)
        }
    }

    // Calcular cuántos no protegidos se pueden mantener
    maxNonProtected := totalLimit - len(protected)
    if maxNonProtected < 0 {
        maxNonProtected = 0
    }

    if len(nonProtected) <= maxNonProtected {
        log.Printf("Contenedores activos: protegidos=%d, no protegidos=%d (total límite=%d)",
            len(protected), len(nonProtected), totalLimit)
        return
    }

    // Ordenar por PID (antigüedad) y eliminar los extras
    sort.Slice(nonProtected, func(i, j int) bool {
        return nonProtected[i].PID < nonProtected[j].PID
    })

    extras := nonProtected[maxNonProtected:]
    log.Printf("Contenedores a eliminar: %d\n", len(extras))

    for _, cont := range extras {
        // Buscar ID real en dockerMap
        containerID := ""
        for id, name := range dockerMap {
            if strings.Contains(name, cont.Name) {
                containerID = id
                break
            }
        }
        if containerID != "" {
            log.Printf("Eliminando contenedor antiguo: %s (Nombre: %s, PID: %d)",
                containerID, cont.Name, cont.PID)
            stopAndRemoveContainer(containerID)
        }
    }
}

// Función para vaciar las tablas al cerrar
func cleanupDatabase() error {
    db, err := sql.Open("sqlite3", *pathDatabase)
    if err != nil {
        return fmt.Errorf("error opening database for cleanup: %v", err)
    }
    defer db.Close()

    tables := []string{
        "SysInfo",
        "ContInfo", 
        "TopRamSys",
        "TopCpuSys",
        "TopRamCont",
        "TopCpuCont",
    }

    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
        if err != nil {
            return fmt.Errorf("error cleaning table %s: %v", table, err)
        }
        log.Printf("Tabla %s limpiada", table)
    }

    // Opcional: Compactar la base de datos después de borrar
    _, err = db.Exec("VACUUM")
    if err != nil {
        log.Printf("Warning: no se pudo compactar la base de datos: %v", err)
    }

    log.Println("Limpieza de base de datos completada")
    return nil
}


func main() {
	  flag.Parse()
		if err := setupCronJob(); err != nil {
			log.Printf("Warning: could not setup cronjob: %v", err)
		}
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
    
		ticker := time.NewTicker(*pollInterval)
		defer ticker.Stop()
		log.Println("Daemon started with cronjob management. Press Ctrl+C to stop gracefully.")

		
		for {
			select {
			case <- shutdown:
				log.Println("Received shutdown signal. Cleaning up...")
				ticker.Stop()

				if err := removeCronJob(); err != nil {
					log.Printf("Error removing cronjob: %v", err)
				}
				// Limpiar base de datos
    		if err := cleanupDatabase(); err != nil {
        	log.Printf("Error cleaning database: %v", err)
    		} else {
        	log.Println("Database cleaned successfully")
    		}
				return
			case <- ticker.C:
				fmt.Println("hola repetido")
				process, total, free, used, err := parseSysProc(*sysProcPath)
        if err != nil {
	      	fmt.Printf("Error leyendo sysinfo: %v\n", err)
        } else {
					db, err := sql.Open("sqlite3", *pathDatabase)
					if err != nil {
						log.Fatal(err)
					}
					defer db.Close()

					if err := ensureDB(db); err != nil {
						log.Fatal("Error creando tablas: ", err)
					}

					createdAt := time.Now().UnixMilli()
					// === Guardar SysInfo ===
					_, err = db.Exec(`INSERT INTO SysInfo (created_at, total_ram, free_ram, used_ram, total_process) 
						VALUES (?, ?, ?, ?, ?)`, createdAt, total, free, used, len(process))
					if err != nil {
						log.Println("Error insertando SysInfo:", err)
					}

					// === Top 5 RAM del sistema ===
					sort.Slice(process, func(i, j int) bool {
						return process[i].RSSMB > process[j].RSSMB
					})
					for i := 0; i < min(5, len(process)); i++ {
						p := process[i]
						_, err := db.Exec(`INSERT INTO TopRamSys (name, pid, rss_mb, created_at) VALUES (?, ?, ?, ?)`,
							p.Name, p.PID, p.RSSMB, createdAt)
						if err != nil {
							log.Println("Error insertando TopRamSys:", err)
						}
					}

					// === Top 5 CPU del sistema ===
					sort.Slice(process, func(i, j int) bool {
						return process[i].CPUUsage > process[j].CPUUsage
					})
					for i := 0; i < min(5, len(process)); i++ {
						p := process[i]
						_, err := db.Exec(`INSERT INTO TopCpuSys (name, pid, cpu, created_at) VALUES (?, ?, ?, ?)`,
							p.Name, p.PID, p.CPUUsage, createdAt)
							fmt.Println("Insertando en SQLite:", p.CPUUsage)

						if err != nil {
							log.Println("Error insertando TopCpuSys:", err)
						}
					}
				}

				// === Contenedores ===
				conts, err := parseContProcJSON(*contProcPath)
				if err != nil {
					fmt.Printf("Error leyendo contenedores: %v\n", err)
				} else {
					db, err := sql.Open("sqlite3", *pathDatabase)
					if err != nil {
						log.Fatal(err)
					}
					defer db.Close()
					
					createdAt := time.Now().UnixMilli()
					// Resumen de contenedores
					var totalCPU, totalMem float64
					for _, c := range conts {
						totalCPU += c.CPUUsage
						totalMem += c.MemoryUsage
					}

					_, err = db.Exec(`INSERT INTO ContInfo (created_at, total_ram, free_ram, used_ram, total_cont) 
						VALUES (?, ?, ?, ?, ?)`, createdAt, totalMem, totalCPU, 0, len(conts))
					if err != nil {
						log.Println("Error insertando ContInfo:", err)
					}

					// === Top 5 RAM contenedores ===
					sort.Slice(conts, func(i, j int) bool {
						return conts[i].RSS > conts[j].RSS
					})
					for i := 0; i < min(5, len(conts)); i++ {
						c := conts[i]
						_, err := db.Exec(`INSERT INTO TopRamCont (name, pid, rss, created_at) VALUES (?, ?, ?, ?)`,
							c.Name, c.PID, c.RSS, createdAt)
						if err != nil {
							log.Println("Error insertando TopRamCont:", err)
						}
					}


					// === Top 5 CPU contenedores ===
					sort.Slice(conts, func(i, j int) bool {
						return conts[i].CPUUsage > conts[j].CPUUsage
					})
					for i := 0; i < min(5, len(conts)); i++ {
						c := conts[i]
						_, err := db.Exec(`INSERT INTO TopCpuCont (name, pid, cpu, created_at) VALUES (?, ?, ?, ?)`,
							c.Name, c.PID, c.CPUUsage, createdAt)
						if err != nil {
							log.Println("Error insertando TopCpuCont:", err)
						}
					}
				}

				enforceContainerLimit(conts, 11)

			}

		}

    /*
	  process, total, free, used, err := parseSysProc(*sysProcPath)
		if err != nil {
			fmt.Printf("Error leyendo sysinfo: %v\n", err)
		} else {
			fmt.Printf("=== INFORMACIÓN DEL SISTEMA ===\n")
      fmt.Printf("Total RAM: %d MB (%.2f GB)\n", total, float64(total)/1024)
      fmt.Printf("Free RAM:  %d MB (%.2f GB)\n", free, float64(free)/1024)
      fmt.Printf("Used RAM:  %d MB (%.2f GB)\n", used, float64(used)/1024)
      if total > 0 {
				fmt.Printf("Uso: %.1f%%\n", float64(used)*100/float64(total))
			}
      fmt.Println()
		}

		//fmt.Printf("\n=== TOP 5 por RAM ===\n")
		fmt.Printf("=== PROCESOS DEL SISTEMA (%d) ===\n", len(process))
	// CORREGIDO: Iterar sobre el slice en lugar de imprimirlo directamente
		for i, p := range process {
			fmt.Printf("%d. %s (PID: %d)\n", i+1, p.Name, p.PID)
			fmt.Printf("   RSS: %d MB\n", p.RSSMB)
			fmt.Printf("   VSZ: %d MB\n", p.VSZMB)
			fmt.Printf("   Estado: %s\n", p.State)
			fmt.Printf("   CPU: %.2f%%\n", p.CPUUsage)
			if p.CmdLine != "" {
				fmt.Printf("   Cmd: %s\n", p.CmdLine)
			}
			fmt.Println()
		}
		
		// PROCESOS DEL SISTEMA QUE MAS RAM ESTAN USANDO
		fmt.Printf("\n=== TOP 5 por RAM procesos del Sistema ===\n")
		sort.Slice(process, func(i, j int) bool {
			return process[i].RSSMB > process[j].RSSMB
		})
		for i:=0; i<min(5, len(process)); i++{
			c := process[i]
			fmt.Printf("%d. %s (PID: %d) - RAM: %d MB\n",
		    i+1, c.Name, c.PID, c.RSSMB)
		}

		// PROCESOS DEL SISTEMA QUE MAS PORCENTAJE DE CPU ESTAN USANDO
		fmt.Printf("\n=== TOP 5 por CPU procesos del Sistema ===\n")
		sort.Slice(process, func(i, j int) bool {
			return process[i].CPUUsage > process[j].CPUUsage
		})
		for i:=0; i<min(5, len(process)); i++{
			c := process[i]
			fmt.Printf("%d. %s (PID: %d) - CPU: %1.f%%\n",
				i+1, c.Name, c.PID, c.CPUUsage)
		}
		

    // función que parsea JSON
    conts, err := parseContProcJSON(*contProcPath)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("=== CONTENEDORES (%d) ===\n", len(conts))
    for i, c := range conts {
				fmt.Printf("%d  Name: %s\n", i+1, c.Name)
        fmt.Printf("   PID: %d\n", c.PID)
        fmt.Printf("   RSS:  %d KB (%.2f MB)\n", c.RSS, float64(c.RSS)/1024)
        fmt.Printf("   VSZ:  %d KB (%.2f MB)\n", c.VSZ, float64(c.VSZ)/1024)
        fmt.Printf("   Mem:  %.1f%%\n", c.MemoryUsage)
        fmt.Printf("   CPU:  %.1f%%\n", c.CPUUsage)
        fmt.Printf("   Cmd:  %s\n", c.Cmd)
        fmt.Println()
    }

		// PROCESOS DE CONTENEDORES QUE MAS RAM ESTAN USANDO
		fmt.Printf("\n=== TOP 5 por RAM CONTENEDORES ===\n")
		sort.Slice(conts, func(i, j int) bool {
			return conts[i].RSS > conts[j].RSS
		})
		for i:=0; i<min(5, len(conts)); i++{
			c := conts[i]
			fmt.Printf("%d. %s (PID: %d) - RAM: %d KB (%.1f MB)\n",
		    i+1, c.Name, c.PID, c.RSS, float64(c.RSS)/1024)
		}

		// PROCESOS DE CONTENEDORES QUE MAS PORCENTAJE DE CPU ESTAN USANDO
		fmt.Printf("\n=== TOP 5 por CPU CONTENEDORES ===\n")
		sort.Slice(conts, func(i, j int) bool {
			return conts[i].CPUUsage > conts[j].CPUUsage
		})
		for i:=0; i<min(5, len(conts)); i++{
			c := conts[i]
			fmt.Printf("%d. %s (PID: %d) - CPU: %1.f%%\n",
				i+1, c.Name, c.PID, c.CPUUsage)
		}

    // Mostrar resumen de uso total
    var totalCPU, totalMem float64
    for _, c := range conts {
        totalCPU += c.CPUUsage
        totalMem += c.MemoryUsage
    }
    fmt.Printf("=== RESUMEN TOTAL ===\n")
    fmt.Printf("CPU Total: %.1f%%\n", totalCPU)
    fmt.Printf("Mem Total: %.1f%%\n", totalMem)
		*/
}

func min( a, b int ) int {
	if a < b {
		return a
	}
	return b
}




