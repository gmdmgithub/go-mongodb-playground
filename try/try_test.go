package try

import (
	"log"
	"testing"
	"time"
)

func TestTryExample(t *testing.T) {
	MaxRetries = 20
	SomeFunction := func() (string, error) { //simulate real function
		return "", nil
	}
	var value string
	err := Do(func(attempt int) (bool, error) {
		var err error
		value, err = SomeFunction()
		log.Print(value)
		return attempt < 5, err // try 5 times
	})
	if err != nil {
		time.Sleep(200 * time.Millisecond) // wait XXX milliseconds
		log.Fatalln("error:", err)
	}
}
