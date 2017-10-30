package examples

import "bytes"
import "fmt"
import "strings"
import "time"

func main3() {
	println("asd")
	println("dsa")
	println("abecedary")
	fmt.Println(strings.TrimSpace(string(bytes.ToLower([]byte(time.Now().String())))))
}
