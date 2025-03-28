package main

import (
	"fmt"
	"os"

	_ "github.com/gnzdotmx/ishinobu/pkg/bundles/full"
	"github.com/gnzdotmx/ishinobu/pkg/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
