package main

import (
	"github.com/toorop/go-bittrex"
	"log"
	"time"
)

const (
	API_KEY    = "3211f0dcac7b4de88847d7860ed51837"
	API_SECRET = "db2f8a39cf3848ba86a6c76aee328b2f"
	BUY_STRING    = "BTC"
	SELL_STRING   = "TUSD"
	MARKET_STRING = BUY_STRING + "-" + SELL_STRING
	MIN_GAIN      = 0.02
	MAX_LOSS      = 0.02
	ORDER_RANGE      = 0.02

	BUY_TRIGGER    = 10000.0
	SELL_TRIGGER   = -10000.0
	ORDER_VARIANCE = 0.02
)

var (
	balances          []bittrex.Balance
	orders            []bittrex.Order
	ticker            = bittrex.Ticker{}
	lastPrice         float64
	lastBuyPrice      = 0.00
	buySellIndex      = 0.00
	openOrder         = false

	readyToRun        = false//set to false for the first 60 seconds as my script gathers data
	buyTriggerActive  = false
	sellTriggerActive = false

	BotIp string

	highIndex = 0.00//to get an idea for what the max and min are in one minute
	lowIndex  = 0.00
)


//for myself
type TickerResult struct {
	Last      float64
	Bid       float64
	Ask       float64
	High      float64
	Low       float64
	Volume    float64
	Timestamp int
}

//noinspection ALL
func main() {
	// Bittrex client
	bittrexClient := bittrex.New(API_KEY, API_SECRET)
	/* remember this
	type ExchangeState struct {
		MarketName string
		Nounce     int (number only used once)
		Buys       []OrderUpdate
		Sells      []OrderUpdate
		Fills      []Fill
		Initial    bool
	}
	*/
	ch := make(chan bittrex.ExchangeState, 16)

	//fmt.Println(ch)
	go updateStats(bittrexClient, ch)

	// 取引判断のときにも毎回取得するが、初期値として一定回数分のtickerを取得しておく
	var tick_history []TickerResult
	for i := 0; i < 10; i++ {
		tick_history = append(tick_history, &bittrexClient.GetTicker(MARKET_STRING))
		time.Sleep(1000 * time.Millisecond)
	}
	log.Println("tick_history:",tick_history)
	/*
	for st := range ch {

		//example) log.Println("st:",st)
		//st: {BTC-TUSD 0 [] [{{2800 0.00009689} 0} {{11.1504038 0.00009909} 2}] [] false}
		//   ={MarketName, nounce, buys=[],sells[Quantity], sell[Rate], sell[Type], fills, initial }
		// Order placed?
		updatedIndex := false
		for _, b := range st.Buys {
			//this is real time market movement based on market-string.
			log.Println("Buy: ", b.Quantity, " for ", b.Rate, " as ", b.Type)
			quantity, _ := b.Quantity.Float64()
			rate, _ := b.Rate.Float64()

			//every minutes when buy order occures, index is updated.
			calculateIndex(true, quantity, rate)
			updatedIndex = true
		}
		for _, s := range st.Sells {
			//Sells[Type]=[0:'Market Order',1:'Limit Order',2:'Ceiling',3:'Good-tilCancel',4:'Imediate or cancel']
			log.Println("Sell: ", s.Quantity, " for ", s.Rate, " as ", s.Type)
			quantity, _ := s.Quantity.Float64()
			rate, _ := s.Rate.Float64()
			calculateIndex(false, quantity, rate)
			updatedIndex = true

		}
		// Order actually fills
		for _, f := range st.Fills {
			log.Println("Fill: ", f.Quantity, " for ", f.Rate, " as ", f.OrderType)
			// We could say that lastPrice is technically the fill price
			tmpLPrice, _ := f.Rate.Float64()
			if tmpLPrice > 0.0000001 {
				log.Printf("Latest price: 		%v\n", f.Rate)
				lastPrice = tmpLPrice
			}
		}
		if updatedIndex {
			log.Printf("BuySellIndex: 		%.8f\n", buySellIndex)
			decideBuySell(bittrexClient)
		}
	}

	 */

}





func calculateIndex(buy bool, q float64, r float64) {
	log.Printf("calculateIndex is called.")
	// q is quantity VTC
	// r is the rate)
	percent := 0.00
	// Calculate percentage of rate
	if r > 0.0000 && q > 0.0000 && lastPrice > 0.0000 && readyToRun {
		percent = lastPrice / r
		if buy {
			//log.Printf("Buy percent: %.8f\n", percent)
			//log.Printf("Buy quantity: %.8f\n", q)
			//log.Printf("Buy?: %.8f > %.8f && %.8f < %.8f\n", percent, 1.00-ORDER_RANGE, percent, 1.00+ORDER_RANGE)
			if percent > (1.00-ORDER_RANGE) && percent < (1.00+ORDER_RANGE) {
				buySellIndex = buySellIndex + (percent * q)
			}
		} else {
			//log.Printf("Sell percent: %.4f\n", percent)
			//log.Printf("Sell quantity: %.4f\n", q)
			//log.Printf("Sell?: %.8f > %.8f && %.8f < %.8f\n", percent, 1.00-ORDER_RANGE, percent, 1.00+ORDER_RANGE)
			if percent > (1.00-ORDER_RANGE) && percent < (1.00+ORDER_RANGE) {
				percent = percent - 2.00 // Reverse percent, lower is higher
				buySellIndex = buySellIndex + (percent * q)
			}
		}
	}
	if buySellIndex > highIndex {
		highIndex = buySellIndex
	}
	if buySellIndex < lowIndex {
		lowIndex = buySellIndex
	}
	// Reset really high or low numbers due to startup
	if highIndex > 5000000.00 || lowIndex < -5000000.00 {
		highIndex = 0.00
		lowIndex = 0.00
		buySellIndex = 0.00
	}
}

