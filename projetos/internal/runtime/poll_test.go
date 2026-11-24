package runtime_test

import (
	"testing"
	"time"

	"github.com/kyriosdata/runner/internal/runtime"
)

func TestPollFicaProntoNaSegundaTentativa(t *testing.T) {
	calls := 0
	err := runtime.Poll(5, time.Millisecond, func() bool {
		calls++
		return calls == 2
	})
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if calls != 2 {
		t.Fatalf("esperado 2 chamadas, obteve %d", calls)
	}
}

func TestPollEsgotaTentativas(t *testing.T) {
	calls := 0
	err := runtime.Poll(3, time.Millisecond, func() bool {
		calls++
		return false
	})
	if err == nil {
		t.Fatal("esperado erro quando nenhuma tentativa fica pronta")
	}
	if calls != 3 {
		t.Fatalf("esperado 3 chamadas, obteve %d", calls)
	}
}

func TestPollZeroTentativas(t *testing.T) {
	called := false
	err := runtime.Poll(0, time.Millisecond, func() bool {
		called = true
		return true
	})
	if err == nil {
		t.Fatal("esperado erro com zero tentativas")
	}
	if called {
		t.Fatal("ready não deveria ser chamada")
	}
}
