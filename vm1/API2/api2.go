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
