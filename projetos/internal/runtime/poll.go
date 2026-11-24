package runtime

import (
	"fmt"
	"time"
)

// Poll repete ready até ela retornar verdadeiro ou esgotar as tentativas.
// Depois de cada tentativa malsucedida espera delay, inclusive a última.
func Poll(attempts int, delay time.Duration, ready func() bool) error {
	for i := 0; i < attempts; i++ {
		if ready() {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("não ficou pronto após %d tentativas", attempts)
}
