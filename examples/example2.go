package examples

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

func main22() {
	println("asd")
	println("dsa")
	println("abecedary")
	fmt.Println(strings.TrimSpace(string(bytes.ToLower([]byte(time.Now().String())))))
}
