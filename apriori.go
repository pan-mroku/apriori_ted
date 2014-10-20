package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Product struct {
	Name           string
	BoughtQuantity int
	Support        float64
	MinimalSupport bool
}
type Basket struct {
	Id       string
	Products []*Product
}

func main() {
	var filename = flag.String("filename", "marketbasket.csv", "csv file to load")
	flag.StringVar(filename, "f", "marketbasket.csv", "csv file to load")
	var minimalSupport = flag.Float64("support", 0.002, "Minimal support")
	flag.Float64Var(minimalSupport, "s", 0.002, "Minimal support")
	flag.Parse()
	file, err := os.Open(*filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	firstRecord, _ := reader.Read()
	var Products []Product
	Products = make([]Product, len(firstRecord)-1) //pierwsza colmuna to basket id
	for i := 1; i < len(firstRecord); i++ {
		Products[i-1] = Product{strings.TrimSpace(firstRecord[i]), 0, 0, true}
	}

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var Baskets = make([]Basket, len(records))
	for key, record := range records {
		Baskets[key].Id = record[0]
		for i := 1; i < len(record); i++ {
			var field = strings.TrimSpace(record[i])
			if field == "true" {
				Baskets[key].Products = append(Baskets[key].Products, &(Products[i-1]))
				Products[i-1].BoughtQuantity++
			}
		}
	}

	for i := 0; i < len(Products); i++ {
		Products[i].Support = float64(Products[i].BoughtQuantity) / float64(len(Baskets))
		Products[i].MinimalSupport = Products[i].Support >= *minimalSupport
	}

	fmt.Println(Products)
}
