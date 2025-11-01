package main

import (
	"fmt"
	"os"

	"github.com/gurleensethi/cradle/config"
)

func main() {
	err := config.Init()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
