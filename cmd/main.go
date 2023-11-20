package main

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New()
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fmt.Sprintf(`[{"id": "c", "type": storage.Counter, "delta": 7 }]`))
	r, err := req.Post("http://127.0.0.1:8080/updates/")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r.Status()))
}
