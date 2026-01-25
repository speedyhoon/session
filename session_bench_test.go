package session

import "testing"

// go test -short -bench . -benchmem -benchtime=30s -shuffle=on

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateID()
	}
}
