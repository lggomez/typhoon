package typhoon

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

func main() {
	fmt.Println(strings.TrimSpace(string(bytes.ToLower([]byte(time.Now().String())))))
}
