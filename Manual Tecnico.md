# Proyecto: APIs Distribuidas con containerd, Docker y Zot

Este proyecto consiste en implementar **tres máquinas virtuales (VMs)** con APIs en **Go**, contenerizadas y distribuidas mediante `containerd` y `Zot` como registro OCI.

---

## 🔹 Estructura del Proyecto

- **VM1** → API1 (expone puerto 8081, se comunica con API2) es la api + containerd.
- **VM2** → API2 (expone puerto 8082). es la api + containerd
- **VM3** → Registro **Zot**.

---

# Instalación de dependencias para creación de Máquinas Virtuales

Para poder crear y administrar máquinas virtuales en Linux, se instalaron las siguientes herramientas y dependencias.

---

## 1. Actualizar el sistema
Antes de instalar cualquier paquete, es recomendable actualizar los repositorios y el sistema:

### En **Archlinux**
```bash
sudo pacman -Syu 
```

## 2. Instalar Virtualización (KVM + Libvirt)
Necesitamos soporte para virtualización y gestión de máquinas virtuales:
```bash
sudo pacman -S qemu-full virt-manager virt-viewer dnsmasq vde2 bridge-utils openbsd-netcat
```

## 3. Habilitar y arrancar libvirt
Activamos el servicio de libvirtd para poder usar virsh y virt-manager:
```bash
sudo systemctl enable libvirtd
sudo systemctl start libvirtd
systemctl status libvirtd
```

## 🔹 Instalaciones iniciales en VM1 y VM2

###  En todas las VMs la instalacion de dependencias
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git build-essential
```

### Instalacion de Go
```bash
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version

```

### Instalacion de containerd
```bash
sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc


echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update

sudo apt-get install containerd.io
```


## Instalaciones en VM3
### Instalacion de Docker
```bash
sudo apt install -y docker.io
sudo systemctl enable --now docker
docker --version
```

### Instalacion de Zot
```bash
docker run -d -p 5000:5000 --name zot_project ghcr.io/project-zot/zot-minimal:latest
```

## Desarrollo de API's en Go para VM1 y VM2
### API1 VM1 (puerto 8081)
```bash
package main

import (
    "io/ioutil"
    "net/http"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()

    carnet := "202000774"

    app.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "mensaje": "Hola desde la API1: API1 en la VM1, desarrollada por el estudiantes Gerson Gonzalez con carnet: 202000774",
        })
    })

    app.Get("/api1/"+carnet+"/llamar-api2", func(c *fiber.Ctx) error {
        resp, err := http.Get("http://vm1:8082/")
        if err != nil {
            return c.JSON(fiber.Map{"Error al llamar a la api2": err.Error()})
        }
        defer resp.Body.Close()
        body, _:=ioutil.ReadAll(resp.Body)
        return c.Send(body)
    })

    app.Get("/api1/"+carnet+"/llamar-api3", func(c *fiber.Ctx) error {
        resp, err:= http.Get("http://vm2:8083/")
        if err != nil {
            return c.JSON(fiber.Map{"Error al llamar a la api3": err.Error()})
        }
        defer resp.Body.Close()
        body, _:=ioutil.ReadAll(resp.Body)
        return c.Send(body)
    })
    app.Listen("0.0.0.0:8081")
}

```

### API2 VM1 
```bash
package main

import (
    "io/ioutil"
    "net/http"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()
    carnet := "202000774"

    app.Get("/", func(c *fiber.Ctx) error { 
        return c.JSON(fiber.Map{
            "mensaje": "Hola desde la API: API2 en la VM1, desarrollada por el estudiantes Gerson Gonzalez con carnet: 202000774",
        })
    })


    app.Get("/api2/"+carnet+"/llamar-api1", func(c *fiber.Ctx) error {
        resp, err := http.Get("http://vm1:8081/")
        if err != nil {
            return c.JSON(fiber.Map{"Error al llamar a la api1": err.Error()})
        }
        defer resp.Body.Close()
        body, _:=ioutil.ReadAll(resp.Body)
        return c.Send(body)
    })

    app.Get("/api2/"+carnet+"/llamar-api3", func(c *fiber.Ctx) error {
        resp, err:= http.Get("http://vm2:8082/")
        if err != nil {
            return c.JSON(fiber.Map{"Error al llamar a la api3": err.Error()})
        }
        defer resp.Body.Close()
        body, _:=ioutil.ReadAll(resp.Body)
        return c.Send(body)
    })

    app.Listen(":8082")
}

```

### API3 VM2
```bash
package main

import (
    "io/ioutil"
    "net/http"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()
    carnet := "202000774"

    app.Get("/", func(c *fiber.Ctx) error { 
        return c.JSON(fiber.Map{
            "mensaje": "Hola desde la API: API3 en la VM2, desarrollada por el estudiantes Gerson Gonzalez con carnet: 202000774",
        })
    })


    app.Get("/api3/"+carnet+"/llamar-api1", func(c *fiber.Ctx) error {
        resp, err := http.Get("http://vm1:8081/")
        if err != nil {
            return c.JSON(fiber.Map{"Error al llamar a la api1": err.Error()})
        }
        defer resp.Body.Close()
        body, _:=ioutil.ReadAll(resp.Body)
        return c.Send(body)
    })

    app.Get("/api3/"+carnet+"/llamar-api2", func(c *fiber.Ctx) error {
        resp, err:= http.Get("http://vm1:8082/")
        if err != nil {
            return c.JSON(fiber.Map{"Error al llamar a la api2": err.Error()})
        }
        defer resp.Body.Close()
        body, _:=ioutil.ReadAll(resp.Body)
        return c.Send(body)

    })

    app.Listen("0.0.0.0:8083")
}

```

## Archivo dockerfile

Todos los archivos dockerfile fueron creados igual ya que son creados con los mismos nombre pero con carpetas distintas.

### Estructura de las carpetas:
# 📂 Proyecto
- 📂 API1  
  - 📜 main.go  
  - 📜 go.mod  
  - 📜 go.sum  
  - 📜 Dockerfile  
- 📂 API2  
  - 📜 main.go  
  - 📜 go.mod  
  - 📜 go.sum  
  - 📜 Dockerfile  
- 📂 API3  
  - 📜 main.go  
  - 📜 go.mod  
  - 📜 go.sum  
  - 📜 Dockerfile  


### Dockerfile
```bash

FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/api .

# Expone el puerto 8081, 8082 y 8083
EXPOSE 8081

CMD ["./api"]

```

## Construccion de imagenes con containerd (VM1 y VM2)
```bash
sudo apt install -y buildkit

buildctl build \
    --frontend=dockerfile.v0 \
    --local context=. \
    --local dockerfile=. \
    --output type=oci,dest=image.tar

sudo ctr image import image.tar

sudo ctr images ls

```

## Subir imágenes al registro Zot (VM3)
En VM1 y VM2, hacer push hacia el Zot de VM3:

* Etiquetar la imagen:
```bash
sudo ctr images tag <imagen_id> VM3_IP:5000/api1:latest

```
* Subir al Zot
```bash
sudo ctr images push --plain-http VM3_IP:5000/api1:latest
```
