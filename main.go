package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	brazilAPIURL   = "https://brasilapi.com.br/api/cep/v1/"
	viaCEPURL      = "https://viacep.com.br/ws/"
	requestTimeout = 1 * time.Second
)

type Response struct {
	API  string
	Data map[string]interface{}
}

func main() {
	cep := getCEPFromUser()

	responseChannel := make(chan Response, 2)

	go fetchFromAPI(brazilAPIURL, cep, "BrasilAPI", responseChannel)
	go fetchFromAPI(viaCEPURL, cep+"/json/", "ViaCEP", responseChannel)

	select {
	case fastestResponse := <-responseChannel:
		handleResponse(fastestResponse, responseChannel)
	case <-time.After(requestTimeout):
		printTimeout()
	}
}

func getCEPFromUser() string {
	var cep string
	fmt.Print("Digite o CEP: ")
	fmt.Scanln(&cep)
	return cep
}

func fetchFromAPI(baseURL, cep, apiName string, ch chan<- Response) {
	url := baseURL + cep
	client := http.Client{
		Timeout: requestTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		ch <- Response{
			API:  apiName,
			Data: map[string]interface{}{"error": err.Error()}}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Response{
			API:  apiName,
			Data: map[string]interface{}{"error": err.Error()}}
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		ch <- Response{
			API: apiName,
			Data: map[string]interface{}{"error": err.Error()}}
		return
	}

	ch <- Response{API: apiName, Data: data}
}

func handleResponse(fastestResponse Response, ch <-chan Response) {
	select {
	case secondResponse := <-ch:
		printResponse("Resultado mais rápido", fastestResponse)
		printResponse("Resultado da segunda API (mais lenta)", secondResponse)
	case <-time.After(requestTimeout):
		printResponse("Resultado mais rápido", fastestResponse)
	}
}

func printResponse(label string, response Response) {
	fmt.Printf("%s: API %s\n", label, response.API)
	fmt.Printf("Dados: %v\n", response.Data)
}

func printTimeout() {
	fmt.Println("Ambas as APIs excederam o tempo de resposta.")
}
