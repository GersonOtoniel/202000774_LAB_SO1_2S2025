package main

import (
	"bufio"
	"flag"
	//"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	//_ "github.com/mattn/go-sqlite3"
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
)

/*
func ensureDB(db *sql.DB) error {
		schema := `
		CREATE TABLE IF NOT EXIST SysInfo(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at  INTEGER,
			total_mb INTEGER,
			free_mb INTEGER,
			used_mb INTEGER
		);

		CREATE TABLE IF NOT EXIST ContInfo(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at INTEGER,
			pid INTEGER,

		)
		`	
  return nil
}
*/


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
			//cpuStr = strings.TrimSuffix(cpuStr, "%")
			val, err := strconv.ParseFloat(cpuStr, 64)
			if err == nil {
				p.CPUUsage = val
			}
		}
	}
	return p, nil

}

func main() {
	  flag.Parse()
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
}

func min( a, b int ) int {
	if a < b {
		return a
	}
	return b
}
