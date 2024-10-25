package main

import (
	"log"
	"metrics/internal/client"
	"strconv"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {

	agent := client.NewAgent("http://localhost:8080/")

	go func() {
		for {
			agent.CollectData()

			time.Sleep(pollInterval)
		}
	}()

	go func() {
		for {
			stats, err := agent.Stats.Map()
			if err != nil {
				log.Println("error convert stats to map: ", err)
			}
			for types, typesData := range stats {
				//fmt.Println("types", types, "typesData", typesData)
				for k, v := range typesData.(map[string]interface{}) {
					//fmt.Println(k, v)
					vStr := strconv.FormatFloat(v.(float64), 'f', -1, 64)
					agent.PostUpdate(types, k, vStr)

				}
			}

			time.Sleep(reportInterval)
		}
	}()

	for {

	}
}
