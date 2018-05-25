package main

import (
    "fmt"
)

type CurrencyPair struct {
    market string
	ticker string
    value float64
}

type Products struct {
	pairs map[string]CurrencyPair 
}

type OrderBook struct {
    buys []BookEntry
    buyRates OrderTree
    sells []BookEntry
    sellRates OrderTree
}

type BookEntry struct {
    quantity float64
    rate float64
}

func printPrices() {
    for id, exchange := range exchanges {
        products := exchange.GetProducts()

        fmt.Printf("Exchange prices: %s\n", id)
        for productName, pair := range products.pairs {
            fmt.Printf("%s (%s): %f\n", productName, pair.ticker, pair.value)
        }
    }
}

func printOrderBook (orderBook OrderBook) {
    fmt.Println("printing order book:")

    fmt.Println("buys:")
    for i := 0; i < 20; i++ {
        if i < len(orderBook.buys) {
            buy := orderBook.buys[i]

            fmt.Printf("Rate: %f Quant: %f\n", buy.rate, buy.quantity)
        }
    }

    fmt.Println("buys:")
    for i := 0; i < 20; i++ {
        if i < len(orderBook.sells) {
            sell := orderBook.sells[i]

            fmt.Printf("Rate: %f Quant: %f\n", sell.rate, sell.quantity)
        }
    }
}

func fetchOrderBooks() {
    for key, exchange := range exchanges {
        products := exchange.GetProducts()

        for _, product := range products.pairs {
            orderBook := exchange.FetchOrderBook(product.ticker)
            exchanges[key] = exchange.SetOrderBook(product.ticker, orderBook)
        }
    }
}

func fetchPrices() {
    for key, exchange := range exchanges {
        ret := Products{ pairs: make(map[string]CurrencyPair) }
        products := exchange.GetProducts()

        for productName, pair := range products.pairs {
            ret.pairs[productName] = CurrencyPair{
                value: exchange.GetTickerPrice(pair.ticker),
                ticker: pair.ticker,
                market: pair.market }
        }

        exchanges[key] = exchange.SetProducts(ret)
    }

    printPrices()
}
