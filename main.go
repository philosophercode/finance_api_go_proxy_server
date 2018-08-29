package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	ID    float64 `json:"id"`
	Stock string  `json:"stock"`
}

type Response struct {
	Date   []string  `json:"date"`
	Epoch  []int64   `json:"epoch"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Open   []float64 `json:"open"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
	Ok     bool      `json:"ok"`
}

type stockData struct {
	MetaData struct {
		Information string `json:"1. Information"`
		Symbol      string `json:"2. Symbol"`
		Time        string `json:"3. Last Refreshed"`
		Size        string `json:"4. Output Size"`
		TimeZone    string `json:"5. Time Zone"`
	} `json:"Meta Data"`
	TimeSeries map[string]struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	} `json:"Time Series (Daily)"`
}

func handler(request Request) (Response, error) {
	var date []string
	var dateUnix []int64
	var open []float64
	var high []float64
	var low []float64
	var close []float64
	var volume []int64

	var tmpRecords stockData

	apiKey := os.Getenv("KEY")
	symbolStr := request.Stock
	symbol := strings.ToUpper(symbolStr)
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&outputsize=full&apikey=%s", symbol, apiKey)
	response, err := http.Get(url)

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		json.Unmarshal(contents, &tmpRecords)
		if err != nil {
			log.Fatal(err)
		}

		priceDays := tmpRecords.TimeSeries

		var keys []string
		for day := range priceDays {
			keys = append(keys, day)
		}
		sort.Strings(keys)
		date = keys
		for _, day := range keys {
			highPrice := priceDays[day].High
			highPriceFloat, _ := strconv.ParseFloat(highPrice, 64)
			high = append(high, highPriceFloat)

			lowPrice := priceDays[day].Low
			lowPriceFloat, _ := strconv.ParseFloat(lowPrice, 64)
			low = append(low, lowPriceFloat)

			openPrice := priceDays[day].Open
			openPriceFloat, _ := strconv.ParseFloat(openPrice, 64)
			open = append(open, openPriceFloat)

			closePrice := priceDays[day].Close
			closePriceFloat, _ := strconv.ParseFloat(closePrice, 64)
			close = append(close, closePriceFloat)

			volumeDay := priceDays[day].Volume
			volumeDayInt, _ := strconv.ParseInt(volumeDay, 10, 64)
			volume = append(volume, volumeDayInt)

			t, _ := time.Parse("2006-01-02", strings.TrimSpace(day))
			unixTime := (t.UnixNano() / int64(time.Millisecond))
			dateUnix = append(dateUnix, unixTime)
			fmt.Println(day)
			fmt.Println(unixTime)
		}

	}
	return Response{
		Date:   date,
		Epoch:  dateUnix,
		High:   high,
		Low:    low,
		Open:   open,
		Close:  close,
		Volume: volume,
		Ok:     true,
	}, nil

}

func main() {
	lambda.Start(handler)
}
