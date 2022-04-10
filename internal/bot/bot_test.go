package bot

import (
	"testing"

	"github.com/felicson/topd/internal/session"

	gradix "github.com/armon/go-radix"
)

var key = session.UUID(14)

func BenchmarkTree(b *testing.B) {

	b.StopTimer()
	btree := gradix.New()

	_, _ = btree.Insert(key, true)

	for i := 0; i < 100000; i++ {
		_, _ = btree.Insert(session.UUID(14), true)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = btree.Get(key)
	}
}
