package process

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"encoding/json"
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
	CreatedAt string
	Currency string
	Amount float64
	ExchangeURL string
	Checksum string
}

func getSingleTransaction() (unprocessedTransaction, error) {
	url := "https://7np770qqk5.execute-api.eu-west-1.amazonaws.com/prod/get-transaction"
	transaction := new(unprocessedTransaction)
	err := getJSON(url, &transaction)
	return *transaction, err
}

/* Get and Process transactions */
func Process(ctx *gin.Context) {
	transaction, err := getSingleTransaction()
	if err != nil {
		fmt.Print("ERROR")
	} else {
		fmt.Print(prettyPrint(transaction))
	}
}