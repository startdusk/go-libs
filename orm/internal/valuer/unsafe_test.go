package valuer

import (
	"testing"
)

func Test_Unsafe_SetColumns(t *testing.T) {
	testSetColumns(t, NewUnsafeValue)
}
