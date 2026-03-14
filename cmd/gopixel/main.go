package main

import (
	"fmt"
	"log"
	"os"

	gopixels "github.com/saran13raj/go-pixels"
)

func main() {
	output, err := gopixels.FromImagePath(os.Args[1], 0, 20, "halfcell", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}
