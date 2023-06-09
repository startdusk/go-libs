package valuer

import (
	"testing"
)

func Test_Reflect_SetColumns(t *testing.T) {
	testSetColumns(t, NewReflectValue)
}
