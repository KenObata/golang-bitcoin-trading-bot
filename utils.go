package main

import (
	"bytes"
	"github.com/shopspring/decimal"
	"github.com/toorop/go-bittrex"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	orderIsCancel     = false
)

func makeDecision(b *bittrex.Bittrex) {
	log.Printf("makeDecision is called.")
	if openOrder && !orderIsCancel && lastPrice > 0.0000001 {
		// Should we close the open order?
		for _, o := range orders {
			ppu, _ := o.PricePerUnit.Float64()
			if ppu > 0.0000 {
				//log.Printf("Order percent: %.8f\n", ppu/lastPrice)
				//log.Printf("Allowed order variance: %.8f\n", ORDER_VARIANCE)
				//log.Printf("%.8f > %.8f || %.8f < %.8f\n", ppu/lastPrice, 1.00+ORDER_VARIANCE, ppu/lastPrice, 1.00-ORDER_VARIANCE)
				if ppu/lastPrice > (1.00+ORDER_VARIANCE) || ppu/lastPrice < (1.00-ORDER_VARIANCE) {
					log.Println("Canceled order: ", o.OrderUuid)
					err := b.CancelOrder(o.OrderUuid)
					// We assume we only have one order at a time
					if err != nil {
						log.Println("ERROR ", err)
					} else {
						log.Println("Confirmed cancel")
						orderIsCancel = true
					}
				}
			}
		}
	}
	// If we have no open order should we buy or sell?
	if !openOrder {
		if buySellIndex > BUY_TRIGGER {
			if !buyTriggerActive {
				log.Println("BUY TRIGGER ACTIVE!")
				buyTriggerActive = true
			}
			for _, bals := range balances {
				bal, _ := bals.Balance.Float64()
				if BUY_STRING == bals.Currency {
					//log.Printf("Bal: %.4f %s == %s\n", bal/lastPrice, SELL_STRING, bals.Currency)
				}
				//I need to have at least 0.01 BTC to buy.
				if bal > 0.01 && BUY_STRING == bals.Currency && lastPrice > 0.00000001 {
					// Place buy order
					log.Printf("Placed buy order of %.4f %s at %.8f\n=================================================\n", (bal/lastPrice)-5, BUY_STRING, lastPrice)
					order, err := b.BuyLimit(MARKET_STRING, decimal.NewFromFloat((bal/lastPrice)-5), decimal.NewFromFloat(lastPrice))
					if err != nil {
						log.Println("ERROR ", err)
					} else {
						log.Println("Confirmed: ", order)
					}
					lastBuyPrice = lastPrice
					openOrder = true
				}
			}
		} else if buySellIndex < SELL_TRIGGER {
			if !sellTriggerActive {
				log.Println("SELL TRIGGER ACTIVE!")
				sellTriggerActive = true
			}
			for _, bals := range balances {
				bal, _ := bals.Balance.Float64()
				if SELL_STRING == bals.Currency {
					//allow := "false"
					//if allowSell() {
					//	allow = "true"
					//}
					//log.Printf("Bal: %.4f %s == %s && %s\n", bal, BUY_STRING, bals.Currency, allow)
				}
				if bal > 0.01 && SELL_STRING == bals.Currency && lastPrice > 0.00 && allowSell() {
					// Place sell order
					log.Printf("Placed sell order of %.4f %s at %.8f\n=================================================\n", bal, SELL_STRING, lastPrice)
					order, err := b.SellLimit(MARKET_STRING, decimal.NewFromFloat(bal), decimal.NewFromFloat(lastPrice))
					if err != nil {
						log.Println("ERROR ", err)
					} else {
						log.Println("Confirmed: ", order)
					}
					openOrder = true
				}
			}
		}
	}
}

func allowSell() bool {
	if lastBuyPrice > 0 {
		gain := lastPrice / lastBuyPrice
		if gain < (1.00 - MAX_LOSS) {
			return true
		}
		if gain < (1.00 + MIN_GAIN) {
			return false
		}
	}
	return true
}

func subscribeMarket(b *bittrex.Bittrex, ch chan bittrex.ExchangeState) {
	log.Println("Connecting to:", MARKET_STRING)
	err := b.SubscribeExchangeUpdate(MARKET_STRING, ch, nil)
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("Reconnecting....")
	go subscribeMarket(b, ch)//use go routine for parallel execution
}

func checkIP() (string, error) {
	rsp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}