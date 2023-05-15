package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Price struct {
	Usdbrl struct {
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

type Prices struct {
	gorm.Model
	ID    int `gorm:"primaryKey"`
	Price *Price
}

func main() {
	http.HandleFunc("/cotacao", Handler)
	http.ListenAndServe(":8080", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	resp, err := getUSD()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = insertPrice(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
	// err = insertPrice(resp)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
}

func getUSD() (*Price, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil && ctx.Err() != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var price Price
	err = json.Unmarshal(body, &price)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

func insertPrice(price *Price) error {
	db, err := gorm.Open(sqlite.Open("price.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Prices{})
	db.Create(&Prices{Price: price})

	return nil
}
