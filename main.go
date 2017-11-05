package main

import (
   "fmt"
   "net/http"
   "io/ioutil"
   "encoding/json"
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

func main() {
   // CoinMarketCap API
   var url_base = "https://api.coinmarketcap.com/v1/ticker/"
   // Symbols to watch
   symbols := "bitcoin"

   resp, err := http.Get(url_base + symbols)
   if err != nil {
      fmt.Println(err)
   }

   defer resp.Body.Close()
   body, _ := ioutil.ReadAll(resp.Body)

   fmt.Println(string(body[2:len(body)-2]))
   var t Ticker
   json.Unmarshal(body[1:len(body)-2], &t)

   fmt.Println(t)
}
