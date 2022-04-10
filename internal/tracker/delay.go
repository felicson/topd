package tracker

import (
	"fmt"
	"time"
)

func Delay(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s name took %s \n", name, elapsed/time.Nanosecond)
}
