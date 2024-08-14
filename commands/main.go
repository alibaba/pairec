package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/alibaba/pairec/v2/pairecmd/app"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := app.NewPairecCommand()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
