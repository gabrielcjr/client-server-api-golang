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

type ResponseAPI struct {
	gorm.Model
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
	ID         int    `gorm:"primaryKey"`
	ResponseAPI ResponseAPI `gorm:"-"` 
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

	select {
	case <-ctx.Done():
		println("timeout")
		return ctx.Err()
	default:
		db.WithContext(ctx).Create(&Prices{
			ResponseAPI: ResponseAPI{
				Usdbrl: struct {
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
				}{
					Code:       "USD",
					Codein:     "BRL",
					Name:       "US Dollar to Brazilian Real",
					High:       "5.50",
					Low:        "5.40",
					VarBid:     "0.10",
					PctChange:  "1.85",
					Bid:        "5.45",
					Ask:        "5.55",
					Timestamp:  "2023-05-15 10:30:00",
					CreateDate: "2023-05-15",
				},
		}})
	}
	return nil
}
