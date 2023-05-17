package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Usdbrl struct {
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
}

type ResponseAPI struct {
	Usdbrl `json:"USDBRL"`
}

type Prices struct {
	ID          int `gorm:"primaryKey"`
	ResponseAPI `gorm:"ResponseAPI"`
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
}

func getUSD() (*ResponseAPI, error) {
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

	var price ResponseAPI
	err = json.Unmarshal(body, &price)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

func connDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("price.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Prices{})
	return db, nil
}

func insertPrice(price *ResponseAPI) error {
	db, err := connDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	prices := &Prices{
		ResponseAPI: ResponseAPI{
			Usdbrl: Usdbrl{
				Code:       price.Usdbrl.Code,
				Codein:     price.Usdbrl.Codein,
				Name:       price.Usdbrl.Name,
				High:       price.Usdbrl.High,
				Low:        price.Usdbrl.Low,
				VarBid:     price.Usdbrl.VarBid,
				PctChange:  price.Usdbrl.PctChange,
				Bid:        price.Usdbrl.Bid,
				Ask:        price.Usdbrl.Ask,
				Timestamp:  price.Usdbrl.Timestamp,
				CreateDate: price.Usdbrl.CreateDate,
			},
		}}

	fmt.Println(prices)

	select {
	case <-ctx.Done():
		println("timeout")
		return ctx.Err()
	default:
		db.WithContext(ctx).Create(prices)
	}
	return nil
}
