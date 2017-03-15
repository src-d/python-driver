package normalizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bblfsh/sdk/uast"
	"github.com/stretchr/testify/require"
)

var (
	fixtureDir = "fixtures"
)

func TestAnnotate(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("python_example_1.json")
	require.NoError(err)

	n, err := NativeToNoder.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	err = AnnotationRules.Apply(n)
	require.NoError(err)

	missingRole := make(map[string]bool)
	iter := uast.NewPreOrderPathIter(uast.NewPath(n))
	for {
		n := iter.Next()
		if n.IsEmpty() {
			break
		}

		missingRole[n.Node().InternalType] = true
	}

	for k := range missingRole {
		fmt.Println("NO ROLE", k)
	}
}

func TestAnnotatePrettyAnnotationsOnly(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("python_example_1.json")
	require.NoError(err)

	n, err := NativeToNoder.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	err = AnnotationRules.Apply(n)
	require.NoError(err)

	buf := bytes.NewBuffer(nil)
	err = uast.Pretty(n, buf, uast.IncludeAnnotations|uast.IncludeChildren|uast.IncludeTokens)
	require.NoError(err)
	fmt.Println(buf.String())
}

func TestNodeTokens(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("python_example_1.json")
	require.NoError(err)

	n, err := NativeToNoder.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	tokens := uast.Tokens(n)
	require.True(len(tokens) > 0)
	for _, tk := range tokens {
		fmt.Println("TOKEN", tk)
	}
}

func getFixture(name string) (map[string]interface{}, error) {
	path := filepath.Join(fixtureDir, name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(f)
	data := map[string]interface{}{}
	if err := d.Decode(&data); err != nil {
		_ = f.Close()
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	return data, nil
}
