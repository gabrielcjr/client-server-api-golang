package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	//Make request to server
	req, err := http.Get("http://localhost:8080")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	//Read response
	resp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(resp))

}
