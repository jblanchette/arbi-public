package main

import (
  "fmt"
)

type BankBalance struct {
  currency string
  balance float64
  available float64
  hold float64
}

type BankAccount struct {
  accounts map[string]BankBalance
}

func (b BankAccount) printBalance () {
  for key, account := range b.accounts {
    if key != "BTC" && key != "ETH" {
      continue
    }

    if account.balance > 0 {
      fmt.Printf("%s: %f avail: %f\n", key, account.balance, account.available)  
    }
  }
}

func (b BankAccount) setBalance (currency string, bankBalance BankBalance) {
  b.accounts[currency] = bankBalance
}

func (b BankAccount) hasBalance (market string, amount float64) bool {
  if account, ok := b.accounts[market]; ok {
    return account.available >= amount
  } else {
    fmt.Printf("Could not find account for market: %v\n", market)

    return false
  }
}

func (b BankAccount) withdrawAccount (market string, action string, amount float64) (BankAccount, bool) {
  if account, ok := b.accounts[market]; ok {
      fmt.Println("Found account...")
      
      switch action {
      case "available":
        if account.available >= amount {
          account.available -= amount
          account.hold += amount

          fmt.Printf("withdrew available funds in %v from %v\n", amount, market)
        } else {
          fmt.Printf("insufficent funds available from %v\n", market)
          fmt.Printf("available balance: %v requested: %v\n", account.available, amount)
          return b, false
        }
      case "hold":
        if account.hold >= amount {
          account.hold -= amount
          account.balance -= amount

          fmt.Printf("withdrew held funds in %v from %v\n", amount, market)
          fmt.Printf("remaining balance: %v\n", account.balance)
        } else {
          fmt.Printf("insufficent funds held from %v\n", market)
          fmt.Printf("available balance: %v requested: %v\n", account.hold, amount)
          return b, false
        }
      default:
        fmt.Printf("unexpected switch action found: %v\n", action)
        return b, false
      }

      fmt.Printf("Setting account: %v\n", account)
      b.accounts[market] = account
      return b, true
    } else {
      fmt.Printf("Could not find account for market: %v\n", market)
      fmt.Printf("Accounts: %v\n", b.accounts)
      return b, false
    }
}

func (b BankAccount) depositAccount (market string, action string, amount float64) bool {
  if account, ok := b.accounts[market]; ok {
      account.balance += amount

      return true
    } else {
      fmt.Printf("Could not find account for market: %v\n")
      return false
    }
}
