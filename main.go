package main

import (
   "log"
   "net/http"
   "io/ioutil"
   "encoding/json"
   "time"
)

// Ticker JSON structure
type Ticker struct {
   Id string
   Name string
   Symbol string
   Rank int `json:",string"`
   Price_usd float32 `json:",string"`
   Price_btc float32 `json:",string"`
   H24_volume_usd float32 `json:"24h_volume_usd,string"`
   Market_cap_usd float32 `json:",string"`
   Avaialble_supply float32 `json:",string"`
   Total_supply float32 `json:",string"`
   Percent_change_1h float32 `json:",string"`
   Percent_change_24h float32 `json:",string"`
   Percent_change_7d float32 `json:",string"`
   Last_updated uint32 `json:",string"`
}

func (t Ticker) VolSpike (perc float32) {
}

func getTicker (url string) Ticker {
   // Allocate struct for json
   var t Ticker

   // Call API
   resp, err := http.Get(url)
   if err != nil {
      log.Println(err)
   }
   // Read response
   defer resp.Body.Close()
   body, _ := ioutil.ReadAll(resp.Body)

   // Bytearray to struct
   //log.Println(string(body[2:len(body)-2]))
   err1 := json.Unmarshal(body[1:len(body)-2], &t)
   if err1 != nil {
      log.Println(err1)
   }

   return t
}

func main() {
   // anomaly thresh as a percentage
   const anomThresh = .02
   // CoinMarketCap API
   var urlBase = "https://api.coinmarketcap.com/v1/ticker/"
   // Symbols to watch
   symbols := []string{
      "bitcoin",
      "iota",
   }

   // Store last ticks to compare with current
   lastTicks := make([]Ticker, len(symbols))
   // Initial load
   for i, s := range symbols {
      lastTicks[i] = getTicker(urlBase + s)
   }

   // Invoke channel on repeat to monitor coins
   ticker := time.NewTicker(1 * time.Second)
   // Setup exit strategy
   quit := make(chan struct{})

   // Concurrent channel runs core when invoked
   go func() {
      for {
         select {
            // On invoke
            case <- ticker.C:
               // Start a list to store anomalies detected
               anomalies := make([]Ticker, 0)
               // Check all symbols
               for i, s := range symbols {
                  // Fetch latest data
                  t := getTicker(urlBase + s)
                  // If anomaly, add to list
                  if (t.Market_cap_usd -
                        lastTicks[i].Market_cap_usd) > anomThresh {
                     anomalies = append(anomalies, t)
                  }

                  // Finally the current becomes the last
                  lastTicks[i] = t
               }
            // On close channel
            case <- quit:
               ticker.Stop()
               return
         }
      }
   }()

   // Wait forever
   <-make(chan int)
}
