package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/labstack/echo"
)

// config struct
type Config struct {
	Dir     string   `json:"dir"`
	Clients []Client `json:"clients"`
}

func (c Config) AsString() string {
	var clientExport string
	for _, client := range c.Clients {
		clientExport += fmt.Sprintf("%s %s(%s)\n", c.Dir, client.IP, comma_separate(client.Opts))
	}
	return clientExport
}

type Client struct {
	IP   string   `json:"ip"`
	Opts []string `json:"opts"`
}

// Healthcheck
func healthcheck(c echo.Context) error {
	_, err := exec.Command("systemctl", "start", "nfs-server").Output()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return c.String(http.StatusOK, "OK")
}

// Take in ConfigFile and setup NFS server
func config(c echo.Context) error {
	var lstConfig []Config
	if err := c.Bind(&lstConfig); err != nil {
		log.Fatal(err)
		return err
	}

	file, err := os.OpenFile("/etc/exports", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return c.String(500, "Error opening file")
	}
	defer file.Close()

	for _, config := range lstConfig {
		_, err = file.WriteString(config.AsString() + "\n")
		if err != nil {
			log.Fatal(err)
			return c.String(500, "Error writing to file")
		}
	}

	_, err = exec.Command("exportfs", "-a").Output()
	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = exec.Command("systemctl", "restart", "nfs-server").Output()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return c.JSON(http.StatusOK, lstConfig)
}

// UTIL
func comma_separate(s []string) string {
	var str string
	for _, v := range s {
		str += v + ","
	}
	return str[:len(str)-1]
}

// self-destruct
func main() {
	e := echo.New()

	e.GET("/healthcheck", healthcheck)
	e.POST("/config", config)

	log.Fatal(e.Start(":8282"))
}
