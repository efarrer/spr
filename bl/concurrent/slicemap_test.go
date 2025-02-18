package concurrent_test

import (
	"slices"
	"testing"

	"github.com/ejoffe/spr/bl/concurrent"
	"github.com/stretchr/testify/require"
)

func TestSliceMap(t *testing.T) {
	in := []int{1, 2, 3}
	out, err := concurrent.SliceMap(in, func(i int) (int, error) {
		return i + 1, nil
	})

	require.NoError(t, err)

	slices.Sort(out)

	require.Equal(t, []int{2, 3, 4}, out)
}
