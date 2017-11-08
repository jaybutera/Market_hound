package main

import (
   "log"
   "net/http"
   "io/ioutil"
   "encoding/json"
   "time"
   "github.com/gorilla/mux"
   "github.com/gorilla/websocket"
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

type Tuple struct {
   Tick Ticker
   VolumeSpike float32
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

// Configure websockets
var upgrader = websocket.Upgrader{
   ReadBufferSize: 1024,
   WriteBufferSize: 1024,
   CheckOrigin: func(r *http.Request) bool {
      return true
   },
}

func main() {
   // anomaly thresh as a percentage
   const anomThresh = .0
   // CoinMarketCap API
   var urlBase = "https://api.coinmarketcap.com/v1/ticker/"
   // Symbols to watch
   symbols := []string{
      "bitcoin",
      "ethereum",
      "ripple",
      "litecoin",
      "dash",
      "neo",
      "nem",
      "monero",
      "iota",
      "qtum",
      "zcash",
      "bitconnect",
      "lisk",
      "cardano",
      "stellar",
      "hshare",
      "waves",
      "stratis",
      "komodo",
      "ark",
      "electroneum",
      "steem",
      "decred",
      "bitcoindark",
      "bitshares",
      "pivx",
      "vertcoin",
      "monacoin",
      "factom",
      "dogecoin",
   }
   // Channel recieves anomalies as they are found
   anomWatch := make(chan []Tuple)

   // Set up API server
   router := mux.NewRouter()
   router.HandleFunc("/sup", func(w http.ResponseWriter, r *http.Request) {
      json.NewEncoder(w).Encode("nigga")
   })
   router.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
      conn, err := upgrader.Upgrade(w, r, nil)
      if err != nil {
         log.Println(err)
         return
      }

      log.Println("Client subscribed")

      for {
         // Wait for new anomaly list
         anoms := <-anomWatch
         // Serialize list to json
         jsonList, err := json.Marshal(anoms)
         if err != nil {
            log.Println(err)
            continue
         }
         // Send message over ws
         err = conn.WriteMessage(websocket.TextMessage, jsonList)
         if err != nil {
            log.Println(err)
            continue
         }
      }
   })

   go http.ListenAndServe(":8000", router)

   // Store last ticks to compare with current
   lastTicks := make([]Ticker, len(symbols))
   // Initial load
   for i, s := range symbols {
      lastTicks[i] = getTicker(urlBase + s)
   }

   // Invoke channel on repeat to monitor coins
   ticker := time.NewTicker(3 * time.Second)
   // Setup exit strategy
   quit := make(chan struct{})

   // Concurrent channel runs core when invoked
   go func() {
      for {
         select {
            // On invoke
            case <- ticker.C:
               // Start a list to store anomalies detected
               anomalies := make([]Tuple, 0)
               // Check all symbols
               for i, s := range symbols {
                  // Fetch latest data
                  t := getTicker(urlBase + s)
                  //log.Println(t)

                  // Compute perc difference in volume over time
                  volDiff := ((t.Market_cap_usd / lastTicks[i].Market_cap_usd) - 1) * 100
                  // If anomaly, add to list
                  if volDiff >= anomThresh {
                     anomalies = append(anomalies, Tuple{Tick: t, VolumeSpike: volDiff})
                  }

                  // Finally the current becomes the last
                  lastTicks[i] = t
               }

               // Send anomalies list to channel if someones is listening
               select {
               case anomWatch <- anomalies:
               default:
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
