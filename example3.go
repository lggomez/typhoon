package typhoon

import "bytes"
import "fmt"
import "strings"
import "time"

func main() {
	fmt.Println(strings.TrimSpace(string(bytes.ToLower([]byte(time.Now().String())))))
}
