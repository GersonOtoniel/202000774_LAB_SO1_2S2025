package main

import (
    "encoding/json"
    "fmt"
    "os"
		"bufio"
		"strings"
		"strconv"
		"github.com/mattn/go-sqlite3"
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

func ensureDB(db *sql.){

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

func parseSysProc(path string) (total, free, used int64, err error) {
    f, err := os.Open(path)
    if err != nil {
        return
    }
    defer f.Close()
    
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "Total_RAM_KB:") {
            s := strings.TrimSpace(strings.TrimPrefix(line, "Total_RAM_KB:"))
            total, _ = strconv.ParseInt(s, 10, 64)
        } else if strings.HasPrefix(line, "Free_RAM_KB:") {
            s := strings.TrimSpace(strings.TrimPrefix(line, "Free_RAM_KB:"))
            free, _ = strconv.ParseInt(s, 10, 64)
        } else if strings.HasPrefix(line, "Used_RAM_KB:") {
            s := strings.TrimSpace(strings.TrimPrefix(line, "Used_RAM_KB:"))
            used, _ = strconv.ParseInt(s, 10, 64)
        }
    }
    err = scanner.Err()
    return
}

func main() {
	  
	  total, free, used, err := parseSysProc("/proc/sysinfo_so1_202000774")
		if err != nil {
			fmt.Printf("Error leyendo sysinfo: %v\n", err)
		} else {
			fmt.Printf("=== INFORMACIÓN DEL SISTEMA ===\n")
      fmt.Printf("Total RAM: %d KB (%.2f GB)\n", total, float64(total)/1024/1024)
      fmt.Printf("Free RAM:  %d KB (%.2f GB)\n", free, float64(free)/1024/1024)
      fmt.Printf("Used RAM:  %d KB (%.2f GB)\n", used, float64(used)/1024/1024)
      fmt.Printf("Uso: %.1f%%\n", float64(used)*100/float64(total))
      fmt.Println()
		}
		

    // función que parsea JSON
    conts, err := parseContProcJSON("/proc/continfo_so1_202000774")
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
