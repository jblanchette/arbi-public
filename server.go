package main

import (
    "fmt"
    "time"
)

var debug bool = true
var exchanges map[string]ExchangeApi
var tradeBook TradeBook

func makeTradeBook () {
  tradeBook = TradeBook{ open: false }
}

func clearTradeBook () {
  tradeBook = TradeBook{ open: false }
}

func makeExchanges () {
  fmt.Println("Creating exchange instances.")

  exchanges = make(map[string]ExchangeApi)

  exchanges["bittrex"] = BittrexExchange{
    bank: BankAccount{
      accounts: map[string]BankBalance{} },
    orderBooks: map[string]OrderBook{},
    products: Products{
      pairs: map[string]CurrencyPair{
        "eth": CurrencyPair{ market: "ETH", ticker: "BTC-ETH" },
        "usd": CurrencyPair{ market: "BTC", ticker: "USDT-BTC" } } } }

  exchanges["gdax"] = GdaxExchange{
    bank: BankAccount{
      accounts: map[string]BankBalance{} },
    orderBooks: map[string]OrderBook{},
    products: Products{
      pairs: map[string]CurrencyPair{
        "eth": CurrencyPair{ market: "ETH", ticker: "ETH-BTC" },
        "usd": CurrencyPair{ market: "BTC", ticker: "BTC-USD" } } } }

  fmt.Println("Finished creating instances.")
}

func connectExchanges() {
  fmt.Println("Connecting to exchanges...")

  for key, exchange := range exchanges {
    exchanges[key] = exchange.Connect()
  }
}

func connectExchangesWS() {
  fmt.Println("connecting exchange ws...")

  for _, exchange := range exchanges {
    exchange.ConnectWS()
  }
}

func fetchExchangeBanks() {
  fmt.Println("Fetching exchange banks...")

  for _, exchange := range exchanges {
    balances := exchange.GetBalances()
    bank := exchange.GetBank()

    for currency, bankBalance := range balances {
      if (debug) {
        // set fake balances

        if currency == "BTC" {
          bankBalance.balance = 0.5
          bankBalance.available = 0.5 
        } else if currency == "ETH" {
          bankBalance.balance = 5
          bankBalance.available = 5
        }
      }

      bank.setBalance(currency, bankBalance)
    }

    exchange.SetBank(bank)
    exchange.GetBank().printBalance()
  }
}

func printBanks () {
  for key, exchange := range exchanges {
    fmt.Printf("ID: %v\n", key)
    exchange.GetBank().printBalance()
  }
}

func tickHandler () {
  fmt.Println("$*************************************************$")
  fetchOrderBooks()
  fetchPrices()
  
  arbiPrices()

  fmt.Println("$*************************************************$")
}

func StartServer() {
  fmt.Println("###################################################")
  fmt.Println("Starting server...")

  makeTradeBook()
  makeExchanges()

  connectExchanges()
  connectExchangesWS()
  fetchExchangeBanks()

  fmt.Println("Scheduling task every 25 seconds.")
  fmt.Println("###################################################")  

	ticker := time.NewTicker(time.Second * 25)
	for {
		tickHandler()

    fmt.Println("tick...")
		<- ticker.C
	}
}
