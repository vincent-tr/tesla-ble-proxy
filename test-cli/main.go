package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type CarCredentialsRequest struct {
	PrivateKey []byte `json:"privateKey"`
	VIN        string `json:"vin"`
}

type ResponseCode string

const ResponseCredentialsNeeded ResponseCode = "credentials-needed"
const ResponseError ResponseCode = "error"
const ResponseOK ResponseCode = "success"

type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
}

func main() {

	privKey, err := os.ReadFile("../tesla-auth/vehicle-private-key.pem")
	if err != nil {
		panic(err)
	}

	vin := os.Getenv("TESLA_VIN")

	fmt.Printf("VIN: %s\n", vin)

	creds := CarCredentialsRequest{
		PrivateKey: privKey,
		VIN:        vin,
	}

	post("/car-credentials", &creds)

	post("/unlock", nil)

}

func post(path string, reqObj any) (*Response, error) {
	var data []byte

	if reqObj != nil {
		jsonStr, err := json.Marshal(reqObj)
		if err != nil {
			panic(err)
		}

		data = jsonStr
	}

	fmt.Printf("req: %s\n", string(data))

	req, err := http.NewRequest("POST", "http://rpi-ble-proxy:8080"+path, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	if reqObj != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("response Body:", string(body))

	// TODO

	return nil, nil
}
