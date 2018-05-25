package main

import (
    "fmt"
)

const (
    ORDER_ERROR = -1
    ORDER_CREATED = 0
    ORDER_OPEN = 1
    ORDER_FILLED = 2
    ORDER_CANCELED = 3
)

type OrderStatus struct {
    open bool
    reason int
    // todo: some kind of timestamp(s)
}

type Order struct {
    exchangeId string
    kind string
    status OrderStatus
    market string
    size float64
    price float64
    
    // todo: some kind of timestamp
}

type TradeBook struct {
    open bool
    action ArbiAction
    lowOrder Order
    highOrder Order
}

func createOrder (
    exchangeId string, 
    market string, 
    kind string, 
    price float64, 
    size float64) (Order, bool) {
    
    exchange := exchanges[exchangeId]

    order := Order {
        exchangeId: exchangeId,
        kind: kind,
        status: OrderStatus {
            open: false,
            reason: ORDER_CREATED },
        market: market,
        size: size,
        price: price }

    bank, ok := exchange.GetBank().withdrawAccount(
        market,
        "available",
        price * size)

    if ok {
        fmt.Printf("withdrawing %v from %v\n", (price * size), market)
        exchanges[exchangeId] = exchange.SetBank(bank)
    }

    return order, ok
}

func openTrade (action ArbiAction) bool {

    if action.kind == ACTION_NOOP {
        fmt.Println("Got a noop.")
        return false
    }

    if tradeBook.open {
        fmt.Println("Trade already open. cannot open another.")
        return false
    }

    target := action.target

    lowOrder, lowOk := createOrder(
        target.lowId, "BTC", "sell", target.lowTarget.value, tradeSize)
    highOrder, highOk := createOrder(
        target.highId, "ETH", "buy", target.highTarget.value, tradeSize)

    if (lowOk && highOk) {
        tradeBook = TradeBook {
            open: true,
            action: action,
            lowOrder: lowOrder,
            highOrder: highOrder }

        fmt.Printf("Updated trade book.\n")
        return true
    } else {
        fmt.Println("Problem opening orders.")
        return false
    }

}
