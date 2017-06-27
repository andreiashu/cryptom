package main

import (
	"fmt"
	"github.com/andreiashu/kraken-go-api-client"
	"github.com/deckarep/gosx-notifier"
	"github.com/spf13/viper"
	"time"
	"sync"
	"github.com/labstack/gommon/log"
	"os"
)

type OrderStatus struct {
	Open map[string]krakenapi.Order
	sync.Mutex
}

const KRAKEN_LINK = "https://www.kraken.com/login"
// in seconds
const POLL_INTERVAL = 8

func main() {
	// config setup
	viper.SetConfigType("toml")
	viper.SetConfigName(".cryptom") // name of config file (without extension)
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		log.Printf("Fatal error trying to read config: %s \n", err)
		os.Exit(1)
	}

	key := viper.GetString("kraken-key")
	secret := viper.GetString("kraken-secret")
	api := krakenapi.New(key, secret)
	s := &OrderStatus{
		Open: make(map[string]krakenapi.Order),
	}
	orderc, errc := KrakenAccPoll(api, time.Duration(POLL_INTERVAL)*time.Second)
	for {
		select {
		case resp := <-orderc:
			new := map[string]krakenapi.Order{}
			notes := []*gosxnotifier.Notification{}
			obsolete := map[string]krakenapi.Order{}
			s.Lock()
			for orderid, order := range resp.Open {
				if _, ok := s.Open[orderid]; !ok {
					new[orderid] = order
					if len(notes) < 3 {
						notes = append(notes, &gosxnotifier.Notification{
							Title:   fmt.Sprintf("cryptom - New %s Order", order.Description.AssetPair),
							Message: fmt.Sprintf("%s", order.Description.Order),
							Link: KRAKEN_LINK,
							Sender: "com.apple.Safari",
							Sound:   gosxnotifier.Funk,
						})
					}
				}
			}

			for orderid, order := range s.Open {
				if _, ok := resp.Open[orderid]; !ok {
					obsolete[orderid] = order
					notes = append(notes, &gosxnotifier.Notification{
						Title:   fmt.Sprintf("cryptom - Completed %s Order", order.Description.AssetPair),
						Message: fmt.Sprintf("%s", order.Description.Order),
						Link: KRAKEN_LINK,
						Sender: "com.apple.Safari",
						Sound:   gosxnotifier.Funk,
					})
				}
			}

			// delete obsolete orders
			for orderid, _ := range obsolete {
				delete(s.Open, orderid)
			}
			// add new orders
			for orderid, order := range new {
				s.Open[orderid] = order
			}

			s.Unlock()

			if len(notes) > 1 {
				notes = append(notes, &gosxnotifier.Notification{
					Title:   fmt.Sprintf("cryptom %d open; %d finished", len(new), len(obsolete)),
					Message: "Check notifications panel for more...",
					Link: KRAKEN_LINK,
					Sender: "com.apple.Safari",
					Sound:   gosxnotifier.Funk,
				})
			}

			for _, note := range notes {
				err := note.Push()
				if err != nil {
					log.Printf("Notification push error: %s", err)
				}
			}
		case err := <-errc:
			log.Printf("Cryptom error: %s", err)
		}
	}

}

func KrakenAccPoll(api *krakenapi.KrakenApi, interval time.Duration) (<-chan *krakenapi.OpenOrdersResponse, <-chan error) {
	orderc := make(chan *krakenapi.OpenOrdersResponse)
	errc := make(chan error, 1)

	go func() {
		for c := time.Tick(interval); ; <-c {
			orderdata, err := api.OpenOrders(map[string]string{})
			if err != nil {
				errc <- err
			} else {
				//t, _ := json.Marshal(orderdata)
				//fmt.Printf("Orderdata: %s\n", string(t))
				orderc <- orderdata
			}
		}
	}()

	return orderc, errc
}
