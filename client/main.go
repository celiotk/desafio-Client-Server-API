package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	// o client.go terá um timeout máximo de 300ms para receber o resultado do server.go
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal(err)
	}
	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Status:%v\n", res.Status)
		io.Copy(os.Stdout, res.Body)
		return
	}

	// receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON)
	c := Cotacao{}
	err = json.NewDecoder(res.Body).Decode(&c)
	if err != nil {
		log.Fatal(err)
	}

	// salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar:%s", c.Bid)), 0666)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Dólar:%s\n", c.Bid)
}
