package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "18.16.32.6/24"
	parts := strings.Split(s, "/")
	fmt.Println(parts[0])

}
