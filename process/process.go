package process

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"encoding/json"
	"time"
	"math"
)

func roundTo(n float64, decimals int) float64 {
	exp := math.Pow10(decimals)
    return math.Round(n * exp) / exp
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
	CreatedAt time.Time 		`json:"createdAt"`
	Currency string 			`json:"currency"`
	ConvertedAmount float64		`json:"convertedAmount"`
	Checksum string				`json:"checksum"`
}

type postProcessTransaction struct {
	Transactions [] processedTransaction `json:"transactions"`
}

type postProcessResponse struct {
	Success bool
	Passed int64
	Failed int64
}

func getSingleTransaction() (unprocessedTransaction, error) {
	url := "https://7np770qqk5.execute-api.eu-west-1.amazonaws.com/prod/get-transaction"
	t := new(unprocessedTransaction)
	err := getJSON(url, &t)
	return *t, err
}

func getExchangeRates(t unprocessedTransaction, base string) (exchangeRate, error) {
	url := fmt.Sprintf("https://api.exchangeratesapi.io/%s?base=%s", t.CreatedAt.UTC().Format("2006-01-02"), base)
	er := new(exchangeRate)
	err := getJSON(url, &er)
	return *er, err
}

func makeSingleProcessedTransaction(base string, ut unprocessedTransaction) (processedTransaction) {
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

/* Process returns nothing*/
func Process(ctx *gin.Context) {
	ppt := new(postProcessTransaction)

	ppt.Transactions = [] processedTransaction{}

	for i := 0; i < 10; i++ {
		ut, err := getSingleTransaction()
		if err != nil {
			fmt.Print("ERROR")
		} else {
			fmt.Print(prettyPrint(ut))
		}
		ppt.Transactions = append(ppt.Transactions, makeSingleProcessedTransaction("EUR", ut))
	}

	url := "https://7np770qqk5.execute-api.eu-west-1.amazonaws.com/prod/process-transactions"

	s, _ := json.Marshal(ppt);

	r, err := http.Post(url, "application/json", bytes.NewBuffer(s))

	if err != nil {
		fmt.Print("ERROR")
	} else {
		defer r.Body.Close()

		ppr := new(postProcessResponse)

		json.NewDecoder(r.Body).Decode(ppr)

		fmt.Print(prettyPrint(ppr))
	}
}