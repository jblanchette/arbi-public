package main

import (
    "os"
    "fmt"
    "encoding/json"
    ws "github.com/gorilla/websocket"
    gdax "github.com/preichenberger/go-gdax"
    "github.com/avdva/go-bittrex"
)

const (
    WS_TIMEOUT = 30
)

type ExchangeConfiguration struct {
    BittrexKey     string
    BittrexSecret  string
    GdaxSecret     string
    GdaxKey        string
    GdaxPassphrase string
}

type ExchangeApi interface {
    Connect() ExchangeApi
    ConnectWS()
    // DisconnectWS()


    SetProducts(Products) ExchangeApi
    GetProducts() Products

    SetBank(BankAccount) ExchangeApi
    GetBank() BankAccount
    
    GetTickerPrice(string) float64 
    GetBalances() map[string]BankBalance
    
    GetOrderBook(string) OrderBook
    FetchOrderBook(string) OrderBook
    SetOrderBook(string, OrderBook) ExchangeApi
}

type BittrexExchange struct {
    products Products
    orderBooks map[string]OrderBook
    bank BankAccount
    client *bittrex.Bittrex
}

type GdaxExchange struct {
    products Products
    orderBooks map[string]OrderBook
    bank BankAccount
    client *gdax.Client
}

func LoadConfig() ExchangeConfiguration {
    file, _ := os.Open("apiconfig.json")
    decoder := json.NewDecoder(file)
    configuration := ExchangeConfiguration{}
    err := decoder.Decode(&configuration)
    
    if err != nil {
      fmt.Println("error:", err)
    }

    return configuration
}

//
// Connect
//

func (e BittrexExchange) Connect() ExchangeApi {
    fmt.Println("Connecting to bittrex.")
    configuration := LoadConfig()

    e.client = bittrex.New(configuration.BittrexKey, configuration.BittrexSecret)
    return e
}

func (e GdaxExchange) Connect() ExchangeApi {
    fmt.Println("Connecting to gdax.")
    configuration := LoadConfig()

    e.client = gdax.NewClient(configuration.GdaxSecret, 
                              configuration.GdaxKey, 
                              configuration.GdaxPassphrase)
    return e
}

//
// ConnectWS
//

func (e BittrexExchange) ConnectWS() {
    fmt.Println("Connecting bittrex ws")

    var markets []string

    for _, pair := range e.GetProducts().pairs {
        markets = append(markets, pair.ticker)
    }

    fmt.Printf("Using markets: %v\n", markets)

    ch := make(chan bittrex.SummaryState, 16)
    errCh := make(chan error)

    go func() {
        for data := range ch {
            fmt.Println("got data")
            fmt.Println(data)
        }
    }()
    
    go func() {
        errCh <- e.JSubscribeSummaryUpdate(ch, nil, markets...)
    }()

    select {
    case err := <-errCh:
        if err != nil {
            fmt.Println(err)
        }
    }
}

func (e GdaxExchange) ConnectWS() {
  return

  var wsDialer ws.Dialer
  wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
  if err != nil {
    fmt.Println("GDAX WS Error")
    fmt.Println(err.Error())
  }

  subscribe := map[string]string{
    "type": "subscribe",
    "product_id": "BTC-USD",
  }
  if err := wsConn.WriteJSON(subscribe); err != nil {
    fmt.Println(err.Error())
  }

  go func () {
    message:= gdax.Message{}
    for true {
      if err := wsConn.ReadJSON(&message); err != nil {
        fmt.Println("GDAX ws error")
        fmt.Println(err.Error())
        break
      }

      if message.Type == "match" {
        fmt.Printf("gdax message: %v\n", message)
      }
    }
  }()
}

//
// GetBank
//

func (e BittrexExchange) GetBank() BankAccount {
    return e.bank
}

func (e GdaxExchange) GetBank() BankAccount {
    return e.bank
}

func (e BittrexExchange) SetBank(bank BankAccount) ExchangeApi {
    e.bank = bank
    return e
}

func (e GdaxExchange) SetBank(bank BankAccount) ExchangeApi {
    e.bank = bank
    return e
}

//
// GetProducts
//

func (e BittrexExchange) GetProducts() Products {
    return e.products
}

func (e GdaxExchange) GetProducts() Products {
    return e.products
}

//
// SetProducts
//

func (e BittrexExchange) SetProducts(products Products) ExchangeApi {
    e.products = products
    return e
}

func (e GdaxExchange) SetProducts(products Products) ExchangeApi {
    e.products = products
    return e
}

//
// GetTickerPrice
//

func (e BittrexExchange) GetTickerPrice(market string) float64 {
    ticker, err := e.client.GetTicker(market)
    if err != nil {
        fmt.Println("Error fetching bittrex price: ", err.Error())
        return 0
    } else {
        return ticker.Last
    }
}

func (e GdaxExchange) GetTickerPrice(market string) float64 {
    ticker, err := e.client.GetTicker(market)
    if err != nil {
        fmt.Println("Error fetching gdax price: ", err.Error())
        return 0
    } else {
        return ticker.Price
    }
}

//
// GetBalances
//

func (e BittrexExchange) GetBalances() map[string]BankBalance {
    balances, err := e.client.GetBalances()

    if err != nil {
        fmt.Println("Error fetching bittrex balances: ", err.Error())
        return nil
    }

    ret := make(map[string]BankBalance, len(balances))

    for _, balance := range balances {
        ret[balance.Currency] = BankBalance{
            currency: balance.Currency,
            balance: balance.Balance,
            available: balance.Available,
            hold: balance.Pending }
    }

    return ret
}

func (e GdaxExchange) GetBalances() map[string]BankBalance {
    accounts, err := e.client.GetAccounts()
    if err != nil {
        fmt.Println("Error fetching bittrex balances: ", err.Error())
        return nil
    }

    ret := make(map[string]BankBalance, len(accounts))

    for _, balance := range accounts {
        ret[balance.Currency] = BankBalance{
            currency: balance.Currency,
            balance: balance.Balance,
            available: balance.Available,
            hold: balance.Hold }
    }

    return ret
}

//
// SetOrderBook
//

func (e BittrexExchange) SetOrderBook(market string, orderBook OrderBook) ExchangeApi {
    e.orderBooks[market] = orderBook
    return e
}

func (e GdaxExchange) SetOrderBook(market string, orderBook OrderBook) ExchangeApi {
    e.orderBooks[market] = orderBook
    return e
}

//
// GetOrderBook
//

//
// FetOrderBook
//

func (e BittrexExchange) GetOrderBook(market string) OrderBook {
    return e.orderBooks[market]
}

func (e GdaxExchange) GetOrderBook(market string) OrderBook {
    return e.orderBooks[market]
}

func (e BittrexExchange) FetchOrderBook(market string) OrderBook {
    orderBook, err := e.client.GetOrderBook(market, "both", 0)
    if err != nil {
        fmt.Println("Error fetching books for bittrex")
        fmt.Printf("Error: %v\n", err)
        return OrderBook{}
    }
    
    ret := OrderBook{}
    for _, order := range orderBook.Buy {
        quant, _ := order.Quantity.Float64()
        rate,_ := order.Rate.Float64()
        book := BookEntry{ quantity: quant, rate: rate }

        ret.buys = append(ret.buys, book)
        InsertOrderTreeNode(&ret.buyRates, book)
    }

    for _, order := range orderBook.Sell {
        quant, _ := order.Quantity.Float64()
        rate,_ := order.Rate.Float64()
        book := BookEntry{ quantity: quant, rate: rate }

        ret.sells = append(ret.sells, book)
        InsertOrderTreeNode(&ret.buyRates, book)
    }

    return ret
}

func (e GdaxExchange) FetchOrderBook(market string) OrderBook {
    orderBook, err := e.client.GetBook(market, 2)
    if err != nil {
        fmt.Println("Error fetching books for gdax")
        fmt.Printf("Error: %v\n", err)
        return OrderBook{}
    }

    ret := OrderBook{}
    for _, order := range orderBook.Bids {
        book := BookEntry{ quantity: order.Size, rate: order.Price }
        ret.buys = append(ret.buys, book)
        InsertOrderTreeNode(&ret.buyRates, book)
    }

    for _, order := range orderBook.Asks {
        book := BookEntry{ quantity: order.Size, rate: order.Price }
        ret.sells = append(ret.sells, book)
        InsertOrderTreeNode(&ret.sellRates, book)
    }

    return ret
}
