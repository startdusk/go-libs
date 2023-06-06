package unsafe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UnsafeAccessorField(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	accessor := NewUnsafeAccessor(&User{Name: "Tom", Age: 18})
	val, err := accessor.Field("Age")
	require.NoError(t, err)
	assert.Equal(t, 18, val)

	err = accessor.SetField("Age", 81)
	require.NoError(t, err)
	val, err = accessor.Field("Age")
	require.NoError(t, err)
	assert.Equal(t, 81, val)
}
