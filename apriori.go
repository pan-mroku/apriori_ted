package main

import (
	"container/list"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Product struct {
	Name           string
	BoughtQuantity int
	Support        float64
}
type Basket struct {
	Id       string
	Products list.List
}

type ProductPair struct {
	First      *Product
	Second     *Product
	Quantity   int
	Support    float64
	Confidence float64
	Gain       float64
}

var minimalSupport *float64

func SplitBasketIntoProductPairs(basket *Basket, output chan *ProductPair, wg *sync.WaitGroup) {
	defer wg.Done()
	for product1 := basket.Products.Front(); product1 != basket.Products.Back(); product1 = product1.Next() {
		if product1.Value.(*Product).Support < *minimalSupport {
			continue
		}
		for product2 := product1.Next(); product2 != nil; product2 = product2.Next() {
			if product2.Value.(*Product).Support < *minimalSupport {
				continue
			}

			var newpair = ProductPair{product1.Value.(*Product), product2.Value.(*Product), 1, 0, 0, 0}
			output <- &newpair

		}
	}
}

func MisterTallyMan(Baskets *list.List, output chan *ProductPair, comm chan bool) {
	var wg sync.WaitGroup
	for basket := Baskets.Front(); basket != nil; basket = basket.Next() {
		wg.Add(1)
		go SplitBasketIntoProductPairs(basket.Value.(*Basket), output, &wg)
	}
	wg.Wait()
	comm <- true
}

var mutex sync.Mutex

func ManaGatherer(Pairs *list.List, pair *ProductPair, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	mutex.Lock()
	defer mutex.Unlock()

	for pairIterator := Pairs.Front(); pairIterator != nil; pairIterator = pairIterator.Next() {
		if pairIterator.Value.(*ProductPair).First == pair.First && pairIterator.Value.(*ProductPair).Second == pair.Second {
			pairIterator.Value.(*ProductPair).Quantity++
			return
		}
	}

	Pairs.PushFront(pair)
}

func main() {
	//Wczytywanie plików i konfiguracji
	var filename = flag.String("filename", "marketbasket.csv", "csv file to load")
	flag.StringVar(filename, "f", "marketbasket.csv", "csv file to load")
	minimalSupport = flag.Float64("support", 0.002, "Minimal support")
	flag.Float64Var(minimalSupport, "s", 0.002, "Minimal support")
	var minimalConfidence = flag.Float64("confidence", 0.07, "Minimal confidence")
	flag.Float64Var(minimalConfidence, "c", 0.07, "Minimal confidence")
	flag.Parse()
	fmt.Println("Minimal support = ", *minimalSupport, ". Minimal confidence = ", *minimalConfidence)
	var StartTime = time.Now()
	fmt.Println("Opening file ", *filename, ". Time = 0")
	file, err := os.Open(*filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	firstRecord, _ := reader.Read()
	//Pierwsza linijka to definicja kolumn
	var Products []Product
	Products = make([]Product, len(firstRecord)-1) //pierwsza colmuna to basket id
	for i := 1; i < len(firstRecord); i++ {
		Products[i-1] = Product{strings.TrimSpace(firstRecord[i]), 0, 0}
	}

	//Czytanie koszyków
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var Baskets = list.New()
	{
		for _, record := range records {
			var basket Basket
			basket.Id = record[0]
			for i := 1; i < len(record); i++ {
				var field = strings.TrimSpace(record[i])
				if field == "true" {
					//Używam wskaźników, żeby uniknąć porównywania stringów
					basket.Products.PushBack(&Products[i-1])
					Products[i-1].BoughtQuantity++
				}
			}
			Baskets.PushBack(&basket)
		}

		//Liczenie wsparcia dla poszczególnych produktów, by móc część odrzucić przy tworzeniu par
		for i := 0; i < len(Products); i++ {
			Products[i].Support = float64(Products[i].BoughtQuantity) / float64(Baskets.Len())
		}
	}

	//Współbieżne tworzenie par
	fmt.Println("Read ", Baskets.Len(), " baskets. Products support counted. Time = ", time.Now().Sub(StartTime))

	var Pairs = list.New()
	{
		var comm = make(chan bool)
		var pairChannel = make(chan *ProductPair)

		go MisterTallyMan(Baskets, pairChannel, comm)

		var Banana = false
		var pair *ProductPair
		var GathererGroup sync.WaitGroup

		for !Banana {
			select {
			case Banana = <-comm:
				fmt.Println("All interesting product pairs created. Time = ", time.Now().Sub(StartTime))
				break
			case pair = <-pairChannel:
				GathererGroup.Add(1)
				go ManaGatherer(Pairs, pair, &GathererGroup)
				continue
			default:
				continue
			}
		}
		GathererGroup.Wait()
	}

	fmt.Println("Quantities of product pairs summed. Time = ", time.Now().Sub(StartTime))

	//Liczenie parametrów par
	for pair := Pairs.Front(); pair != nil; {
		var tmp *ProductPair
		tmp = pair.Value.(*ProductPair)
		tmp.Support = float64(tmp.Quantity) / float64(Baskets.Len())
		if tmp.Support < *minimalSupport {
			var del = pair
			pair = pair.Next()
			Pairs.Remove(del)
			continue
		}
		var switchedPair = ProductPair{tmp.Second, tmp.First, tmp.Quantity, tmp.Support, 0, 0}
		Pairs.InsertAfter(&switchedPair, pair)
		tmp.Confidence = tmp.Support / tmp.First.Support
		switchedPair.Confidence = switchedPair.Support / switchedPair.First.Support
		for i := 0; i < 2; i++ {
			var del = pair
			pair = pair.Next()
			if del.Value.(*ProductPair).Confidence < *minimalConfidence {
				Pairs.Remove(del)
			}
		}
	}

	fmt.Println("Product pairs with too low support or confidence deleted. Time = ", time.Now().Sub(StartTime))
	fmt.Println("Created pairs (", Pairs.Len(), "): ")

	//Wydruk par
	fmt.Println("First,First_Support,Second,Second_Support,Pair_Support,Confidence,Gain")
	for pair := Pairs.Front(); pair != nil; pair = pair.Next() {
		var tmp = pair.Value.(*ProductPair)
		tmp.Gain = tmp.Confidence / tmp.Second.Support
		fmt.Printf("%v,%v,%v,%v,%v,%v,%v\n", (*tmp.First).Name, (*tmp.First).Support, (*tmp.Second).Name, (*tmp.Second).Support, tmp.Support, tmp.Confidence, tmp.Gain)
	}
}
