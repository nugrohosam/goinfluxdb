package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		connectInflux()
		return c.SendString("Hello, World 👋!")
	})

	app.Listen(":3000")
}

func connectInflux() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "influxdb")
	// Use blocking write client for writes to desired bucket
	writeAPI := client.WriteAPIBlocking("influxdb-org", "influxdb")
	// Create point using full params constructor
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45.0},
		time.Now())
	// write point immediately
	writeAPI.WritePoint(context.Background(), p)
	// Create point using fluent style
	p = influxdb2.NewPointWithMeasurement("stat").
		AddTag("unit", "temperature").
		AddField("avg", 23.2).
		AddField("max", 45.0).
		SetTime(time.Now())
	writeAPI.WritePoint(context.Background(), p)

	// Or write directly line protocol
	line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
	writeAPI.WriteRecord(context.Background(), line)

	// Get query client
	queryAPI := client.QueryAPI("influxdb-org")
	// Get parser flux query result
	result, err := queryAPI.Query(context.Background(), `from(bucket:"influxdb")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
	if err == nil {
		// Use Next() to iterate over query result lines
		for result.Next() {
			// Observe when there is new grouping key producing new table
			if result.TableChanged() {
				fmt.Println("+++++++++++++ ------ +++++++++++++")
				fmt.Printf("table: %s\n", result.TableMetadata().String())
				fmt.Println("+++++++++++++ ------ +++++++++++++")
			}
			// read result
			fmt.Println("+++++++++++++ ------ +++++++++++++")
			fmt.Printf("row: %s\n", result.Record().String())
			fmt.Println("+++++++++++++ ------ +++++++++++++")
		}
		if result.Err() != nil {
			fmt.Println("+++++++++++++ ------ +++++++++++++")
			fmt.Printf("Query error: %s\n", result.Err().Error())
			fmt.Println("+++++++++++++ ------ +++++++++++++")
		}
	} else {
		fmt.Println("+++++++++++++ ------ +++++++++++++")
		fmt.Println(err)
		fmt.Println("+++++++++++++ ------ +++++++++++++")
	}

	// Ensures background processes finishes
	client.Close()
}
