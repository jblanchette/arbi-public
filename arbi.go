package main

import (
  "fmt"
  "math"
)

const (
  ACTION_NOOP = 0
  ACTION_ARBI = 1
  ACTION_REBALANCE = 2
)

type ArbiTarget struct {
  lowId string
  highId string
  lowTarget CurrencyPair
  highTarget CurrencyPair
  lowFinal float64
  highFinal float64
  delta float64
  lowRate float64
  highRate float64
}

type ArbiAction struct {
  kind int
  amount float64
  target ArbiTarget
}

// the min price target to find an arbi trade for
var minArbDelta float64 = 0.0001

// the min target price delta found in an order book to make a trade
var minTargetDelta float64 = 0.001

// the min target delta to rebalance funds
var minRebalanceDelta float64 = 0.00005

var tradeSize float64 = 0.15
var profitTotal float64 = 0

func getExchangeTarget () (ArbiTarget, bool) {
  target := ArbiTarget{
    lowId: "",
    highId: "",
    lowTarget: CurrencyPair{ value: 10000000 },
    highTarget: CurrencyPair{ value: 0 } }

  for id, exchange := range exchanges {
    products := exchange.GetProducts()
    eth := products.pairs["eth"]
    usd := products.pairs["usd"]

    if (eth.value < target.lowTarget.value) {
      target.lowId = id
      target.lowTarget = eth
      target.lowRate = usd.value
    }

    if (eth.value > target.highTarget.value) {
      target.highId = id
      target.highTarget = eth
      target.highRate = usd.value
    }
  }

  if (target.lowId == "" || target.highId == "") {
    return ArbiTarget{}, false
  }

  // calculate ideal price delta
  target.delta = target.highTarget.value - target.lowTarget.value
  return target, true
}

func swapTargetOrders (target ArbiTarget) ArbiTarget {
  return ArbiTarget {
    lowId: target.highId,
    highId: target.lowId,
    lowTarget: target.highTarget,
    highTarget: target.lowTarget,
    lowFinal: target.highFinal,
    highFinal: target.lowFinal,
    lowRate: target.lowRate,
    highRate: target.highRate,
    delta: target.delta }
}

// todo: maybe make this the trade size func
func checkProfitAction (target ArbiTarget) int {
  var priceDelta = target.highTarget.value - target.lowTarget.value

  if (priceDelta <= minRebalanceDelta) {
    if tradeBook.open {
      fmt.Printf("Price under arb delta, rebalance. %f\n", priceDelta)
      return ACTION_REBALANCE
    } else {
      fmt.Printf("Even prices, noop.\n")
      return ACTION_NOOP
    }
  } else {
    if tradeBook.open {
      // todo: maybe do something else in this situation
      //       when the price isn't rebalacable 
      fmt.Printf("Skipping price check since no rebalance. delta: %.5f min: %.5f\n",
        priceDelta, minRebalanceDelta)

      fmt.Printf("Low order: %.5f Low Target: %.5f\n", 
        tradeBook.lowOrder.price, target.lowTarget.value)

      fmt.Printf("High order: %.5f High Target: %.5f\n", 
        tradeBook.highOrder.price, target.highTarget.value)

      return ACTION_NOOP
    }
  }

  var spread = (target.highTarget.value / target.lowTarget.value) - 1
  var profits = (priceDelta * target.highRate)

  profitTotal += profits

  if priceDelta < minArbDelta {
    fmt.Printf("Price delta below min arb delta. price: %.5f min: %.5f\n", 
      priceDelta, minArbDelta)
    return ACTION_NOOP
  }

  fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
  fmt.Printf("Winner: %s\n", target.highId)
  fmt.Printf("Price delta: %f spread: %%%f\n", priceDelta, spread)
  fmt.Printf("Profits: $%.2f\n", profits)
  fmt.Printf("Profit Total: $%.2f\n", profitTotal)
  fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

  return ACTION_ARBI
}

func findTargetPrice (target ArbiTarget) (ArbiAction, bool) {

  lowExchange := exchanges[target.lowId]
  highExchange := exchanges[target.highId]

  lowOrderBook := lowExchange.GetOrderBook(target.lowTarget.ticker)
  highOrderBook := highExchange.GetOrderBook(target.highTarget.ticker)

  lowPrice, lowOk := FindClosestLeaf(
    &lowOrderBook.buyRates, target.lowTarget.value, minTargetDelta)

  highPrice, highOk := FindClosestLeaf(
    &highOrderBook.sellRates, target.highTarget.value, minTargetDelta)

  lowHasFunds := lowExchange.GetBank().hasBalance("BTC", tradeSize)
  highHasFunds := highExchange.GetBank().hasBalance("ETH", tradeSize)

  if (lowOk && highOk && lowHasFunds && highHasFunds) {
    lowPriceDelta := math.Abs(target.lowTarget.value - lowPrice.rate)
    highPriceDelta := math.Abs(target.highTarget.value - highPrice.rate)

    if (lowPriceDelta > minTargetDelta || highPriceDelta > minTargetDelta) {
      fmt.Printf("Price delta above min target delta.\n")
      fmt.Printf("lowprice: %.5f highprice: %.5f min: %.5f\n", 
        lowPriceDelta, highPriceDelta, minTargetDelta)

      return ArbiAction {
        kind: ACTION_NOOP}, false
    }

    priceDelta := highPrice.rate - lowPrice.rate

    fmt.Println("Found arbi target.")
    fmt.Printf("Final price delta: %.5f Ideal delta: %.5f\n", priceDelta, target.delta)

    target.lowFinal = lowPrice.rate
    target.highFinal = highPrice.rate

    return ArbiAction {
      kind: ACTION_ARBI,
      target: target,
      amount: tradeSize }, true
  }

  return ArbiAction {
      kind: ACTION_NOOP }, false
}

func arbiPrices () {
  target, ok := getExchangeTarget()

  if (!ok) {
    fmt.Println("No arbi opp.")

    return
  }

  arbiActionKind := checkProfitAction(target)

  switch arbiActionKind {
  case ACTION_REBALANCE:
    fmt.Println("Doing a rebalance...")
      
    target = swapTargetOrders(target)
    if tradeAction, ok := findTargetPrice(target); ok {
      if openTrade(tradeAction) {
        fmt.Printf("Opened rebalance order: %v\n", tradeBook)

        printBanks()

        // todo: temp way of closing trade book
        clearTradeBook()
      }
    } else {
      fmt.Println("Didnt open order.")
    }

    

  case ACTION_ARBI:
    fmt.Println("Doing an arbi!")
    if tradeAction, ok := findTargetPrice(target); ok {
      if openTrade(tradeAction) {
        fmt.Printf("Opened order: %v\n", tradeBook)

        printBanks()
      }
    } else {
      fmt.Println("Didnt open order.")
    }
  }
  
}
