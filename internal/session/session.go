package session

import (
	"math/rand"
	"strings"
	"time"
)

var sessionData = []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func UUID(length int) string {

	var sess strings.Builder
	sess.Grow(length)
	size := len(sessionData) - 1

	for i := 0; i < length; i++ {
		inx := rand.Intn(size)
		sess.WriteByte(sessionData[inx])
	}
	return sess.String()
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
