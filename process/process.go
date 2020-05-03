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

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("function %s took %s", name, elapsed)
}

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

func getSingleTransaction(ch chan unprocessedTransaction) (unprocessedTransaction, error) {
	url := "https://7np770qqk5.execute-api.eu-west-1.amazonaws.com/prod/get-transaction"
	t := new(unprocessedTransaction)
	err := getJSON(url, &t)
	ch <- *t
	return *t, err
}

func getExchangeRates(t unprocessedTransaction, base string) (exchangeRate, error) {
	url := fmt.Sprintf("https://api.exchangeratesapi.io/%s?base=%s", t.CreatedAt.UTC().Format("2006-01-02"), base)
	er := new(exchangeRate)
	err := getJSON(url, &er)
	return *er, err
}

func processSingleTransaction(base string, ch chan unprocessedTransaction, pCh chan processedTransaction) (processedTransaction, error) {
	ut := <- ch
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

	pCh <- *pt

	return *pt, err
}

func processNTransactions(base string, pCh chan processedTransaction) {
	utCh := make(chan unprocessedTransaction)
	go getSingleTransaction(utCh)
	go processSingleTransaction(base, utCh, pCh)
}

/* Process returns nothing*/
func Process(ctx *gin.Context) {
	defer timeTrack(time.Now(), "Process")
	ppt := new(postProcessTransaction)

	ppt.Transactions = [] processedTransaction{}

	ch := make(chan processedTransaction)
	for i := 0; i < 10; i++ {
		go processNTransactions("EUR", ch)
	}

	for i := 0; i < 10; i++ {
		ppt.Transactions = append(ppt.Transactions, <- ch)
	}

	fmt.Print(prettyPrint(ppt))

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