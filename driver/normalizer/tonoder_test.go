package normalizer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeToNoder(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	n, err := NativeToNoder.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
	fmt.Println("NODE", n)
}
