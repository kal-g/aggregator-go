package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	address = "localhost:50051"
)

func getPayload() map[string]interface{} {
	v := make(map[string]interface{})
	v["id"] = 1
	v["test1"] = 2
	v["test2"] = 2
	v["test3"] = 1
	v["test4"] = 2
	return v
}

func main() {
	successCount := 0
	url := "http://localhost:50051/consume"
	bodyJSON := make(map[string]interface{})
	bodyJSON["payload"] = getPayload()
	bodyJSON["namespace"] = "test"
	bodyData, _ := json.Marshal(bodyJSON)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))

	client := &http.Client{}

	if len(os.Args) >= 2 {
		count, _ := strconv.Atoi(os.Args[1])
		log.Printf("Sending %d messages\n", count)
		// Contact the server
		for i := 0; i < count; i++ {
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
			resp, err := client.Do(req)
			fmt.Printf("%d: %+v, %+v\n", i, resp, err)
			if err == nil {
				if resp.StatusCode == 200 {
					bytes, err := ioutil.ReadAll(resp.Body)
					if err == nil {
						fmt.Printf("%+v\n", string(bytes))
					}

					successCount++
				}
				resp.Body.Close()
			}
		}
	} else {
		for {
			client.Do(req)
		}
	}
	fmt.Printf("Sent %d messages successfully\n", successCount)

}
