package reflect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IterateArray(t *testing.T) {
	cases := []struct {
		name     string
		entity   any
		wantVals []any
		wantErr  error
	}{
		{
			name:     "[3]int",
			entity:   [3]int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
		{
			name:     "[]int",
			entity:   []int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vals, err := IterateArray(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantVals, vals)
		})
	}
}

func Test_IterateMap(t *testing.T) {
	cases := []struct {
		name       string
		entity     any
		wantKeys   []any
		wantValues []any
		wantErr    error
	}{
		{
			name: "map[int]string",
			entity: map[int]string{
				1: "1",
				2: "2",
			},
			wantKeys:   []any{1, 2},
			wantValues: []any{"1", "2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			keys, values, err := IterateMap(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantKeys, keys)
			assert.Equal(t, c.wantValues, values)
		})
	}
}
