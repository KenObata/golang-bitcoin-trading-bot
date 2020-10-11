package main

import (
	"github.com/toorop/go-bittrex"
	"log"
	"time"
)


func Statistics(b *bittrex.Bittrex, ch chan bittrex.ExchangeState) {
	log.Printf("Statistics is called.")

	var err error = nil
	for {
		go func(b *bittrex.Bittrex) {
			balances, err = b.GetBalances()
			if err != nil {
				log.Println("Error:", err)
				// Pause calculations in case of error
				readyToRun = false
			}
			orders, err = b.GetOpenOrders(MARKET_STRING)
			if err != nil {
				log.Println("Error:", err)
				// Pause calculations in case of error
				readyToRun = false
			}

			ticker, err = b.GetTicker(MARKET_STRING)
			if err != nil {
				log.Println("Error:", err)
				// Pause calculations in case of error
				readyToRun = false
			}

			log.Printf("====================================\n")
			log.Printf("Last price: 		%v\n", ticker.Last)
			log.Printf("Index: 			%.4f\n", buySellIndex)
			log.Printf("High Index: 		%.4f\n", highIndex)
			log.Printf("Low Index: 			%.4f\n", lowIndex)
			tmpLast, _ := ticker.Last.Float64()
			if tmpLast > 0.00 {
				lastPrice = tmpLast
			}
			buySellIndex = 0.00
			buyTriggerActive = false
			sellTriggerActive = false

			log.Printf("Bid:			%v\n", ticker.Bid)
			log.Printf("Ask:			%v\n", ticker.Ask)

			// Do we have an open order?
			openOrder = len(orders) > 0
			orderIsCancel = false

			for _, o := range orders {
				log.Println("Pending order: ", o.OrderType, " Quanitity: ", o.QuantityRemaining, "/", o.Quantity, " Price: ", o.PricePerUnit)
			}

			// Where do we have balances
			for _, b := range balances {
				bal, _ := b.Balance.Float64()
				if bal > 0.00 {
					log.Printf("%s:			%v %s %v\n", b.Currency, b.Available, "/", b.Balance)
				}
			}
			log.Printf("====================================\n")

			ip, err := checkIP()
			if err != nil {
				panic(err)
			}
			if BotIp != ip {
				BotIp = ip
				go subscribeMarket(b, ch)
			}

		}(b)
		<-time.After(60 * time.Second)
		// Wait 60 to init and collect data
		readyToRun = true
	}
}
