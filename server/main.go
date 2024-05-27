package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "cotacao.db")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS cotacao (id INTEGER PRIMARY KEY AUTOINCREMENT,code text, codein text, name text, high text, low text, varBid text, pctChange text, bid text, ask text, timestamp text, create_date text);"); err != nil {
		db.Close()
		log.Fatal(err)
	}
	db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", cotacao)
	http.ListenAndServe(":8080", mux)
}

func cotacao(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, err := getCotacao(ctx)
	if err != nil {
		fmt.Printf("Falha ao consultar a cotação. %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = saveCotacao(ctx, c)
	if err != nil {
		fmt.Printf("Falha ao salvar cotação na base de dados. %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(c.USDBRL); err != nil {
		fmt.Printf("Falha no encode do json. %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// saveCotacao salva cotação no banco
func saveCotacao(ctx context.Context, c *Cotacao) error {

	// o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	db, err := sql.Open("sqlite3", "cotacao.db")
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO cotacao (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.USDBRL.Code, c.USDBRL.Codein, c.USDBRL.Name, c.USDBRL.High, c.USDBRL.Low, c.USDBRL.VarBid, c.USDBRL.PctChange, c.USDBRL.Bid, c.USDBRL.Ask, c.USDBRL.Timestamp, c.USDBRL.CreateDate,
	)

	return err
}

// getCotacao consulta cotação do Dólar de API externa
func getCotacao(ctx context.Context) (*Cotacao, error) {

	// timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	// consumir a API contendo o câmbio de Dólar e Real
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var c Cotacao
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
