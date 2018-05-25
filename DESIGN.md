

trades -

 * fetch prices
 * fetch order books
 * if no open trade -
    * find low/high currency pair exchanges
    * calculate trade success probability:
        * check profit threshold 
        * check order book volume
        * check price history
        
    * open trade on low / high markets

------------------------------------

trade list:

* no trades open:
 - openTrade (ArbiAction)
