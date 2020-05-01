package process

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"encoding/json"
	"time"
	"math"
)

func roundTo(n float64, decimals uint32) float64 {
    return math.Round(n * float64(decimals)) / float64(decimals)
}

func prettyPrint(i interface{}) string {
    s, _ := json.MarshalIndent(i, "", "\t")
    return string(s)
}

func getJSON(url string, target interface{}) error {
    r, err := http.Get(url)
    if err != nil {
        return err
    }
    defer r.Body.Close()

    return json.NewDecoder(r.Body).Decode(target)
}

type unprocessedTransaction struct {
	CreatedAt time.Time
	Currency string
	Amount float64
	ExchangeURL string
	Checksum string
}

type exchangeRate struct {
	Rates map[string] float64
}

type processedTransaction struct {
	CreatedAt time.Time
	Currency string
	ConvertedAmount float64
	Checksum string
}

func getSingleTransaction() (unprocessedTransaction, error) {
	url := "https://7np770qqk5.execute-api.eu-west-1.amazonaws.com/prod/get-transaction"
	t := new(unprocessedTransaction)
	err := getJSON(url, &t)
	return *t, err
}

func getExchangeRates(t unprocessedTransaction, base string) (exchangeRate, error) {
	url := fmt.Sprintf("https://api.exchangeratesapi.io/%s?base=%s", t.CreatedAt.Format("2006-01-02"), base)
	er := new(exchangeRate)
	err := getJSON(url, &er)
	return *er, err
}

func makeSingleProcessedTransaction(base string) (processedTransaction) {
	ut, err := getSingleTransaction()
	if err != nil {
		fmt.Print("ERROR")
	} else {
		fmt.Print(prettyPrint(ut))
	}
	r, err := getExchangeRates(ut, base)
	if err != nil {
		fmt.Print("ERROR")
	} else {
		fmt.Print(prettyPrint(r))
	}
	pt := new(processedTransaction)
	pt.Checksum = ut.Checksum
	pt.CreatedAt = ut.CreatedAt
	pt.Currency = ut.Currency
	pt.ConvertedAmount = roundTo(ut.Amount / r.Rates[ut.Currency], 4)
	fmt.Print(prettyPrint(pt))

	return *pt
}

/* Process */
func Process(ctx *gin.Context) {
	makeSingleProcessedTransaction("EUR")
}