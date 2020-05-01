package process

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"encoding/json"
	"time"
)

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

func getSingleTransaction() (unprocessedTransaction, error) {
	url := "https://7np770qqk5.execute-api.eu-west-1.amazonaws.com/prod/get-transaction"
	t := new(unprocessedTransaction)
	err := getJSON(url, &t)
	return *t, err
}

func getExchangeRates(t unprocessedTransaction, base string) (exchangeRate, error) {
	url := fmt.Sprintf("https://api.exchangeratesapi.io/%s?base=%s", t.CreatedAt.Format("2006-01-02"), base)
	fmt.Print(url)
	er := new(exchangeRate)
	err := getJSON(url, &er)
	return *er, err
}

/* Process */
func Process(ctx *gin.Context) {
	t, err := getSingleTransaction()
	if err != nil {
		fmt.Print("ERROR")
	} else {
		fmt.Print(prettyPrint(t))
	}
	rate, err := getExchangeRates(t, "EUR")
	if err != nil {
		fmt.Print("ERROR")
	} else {
		fmt.Print(prettyPrint(rate))
	}
}