package main

import (
	"encoding/csv"
	"fmt"
	//"io"
	"container/list"
	"flag"
	"os"
	"strings"
)

type Product string
type Basket struct {
	Id       string
	Products list.List
}

func main() {
	var filename = flag.String("filename", "marketbasket.csv", "csv file to load")
	flag.StringVar(filename, "f", "marketbasket.csv", "csv file to load")
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
	for _, field := range firstRecord {
		fmt.Println(strings.TrimSpace(field))
	}
	records, _ := reader.ReadAll()
	fmt.Println(records[0])

}
