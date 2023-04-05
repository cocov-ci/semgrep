package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/cocov-ci/semgrep/mocks"
)

type testHelper struct {
	ctx  *mocks.MockContext
	exec *mocks.Mockexecutor
	ru   *runner
}

func newHelper(ctrl *gomock.Controller) *testHelper {
	exec := mocks.NewMockexecutor(ctrl)
	ctx := mocks.NewMockContext(ctrl)
	ctx.EXPECT().L().
		DoAndReturn(func() *zap.Logger { return zap.NewNop() }).
		AnyTimes()

	return &testHelper{
		ctx:  ctx,
		exec: exec,
		ru:   newRunner(exec),
	}
}

func (h *testHelper) start() ([]*result, error) {
	return h.ru.run(h.ctx)
}

func createFixtureYaml(t *testing.T, dirPath string, remove bool) string {
	fileName := "semgrep.yaml"
	yamlPath := filepath.Join(dirPath, fileName)

	err := os.WriteFile(yamlPath, []byte("foo"), os.ModePerm)
	require.NoError(t, err)

	if remove {
		t.Cleanup(func() {
			err = os.Remove(yamlPath)
			require.NoError(t, err)
		})
	}
	return yamlPath
}

func createFixtureFiles(t *testing.T, dirPath string) {
	fixtureDir, err := os.MkdirTemp(dirPath, "")
	require.NoError(t, err)

	_ = createFixtureYaml(t, fixtureDir, false)
	t.Cleanup(func() {
		err = os.RemoveAll(fixtureDir)
		require.NoError(t, err)
	})
	return
}

func badOutput() []byte {
	return []byte{234}
}

func expectedOutput(t *testing.T, filePath, category string, numResults int) []byte {
	res := make([]*result, 0, numResults)
	for i := 0; i < numResults; i++ {
		msg := fmt.Sprintf("message for issue at path %s", filePath)
		refs := []string{fmt.Sprintf("ref for issue at path %s", filePath)}
		r := newResult(filePath, msg, category, refs)
		res = append(res, r)
	}

	rslt := &results{Results: res}

	data, err := json.Marshal(rslt)
	require.NoError(t, err)
	return data
}
