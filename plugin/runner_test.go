package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cocov-ci/go-plugin-kit/cocov"
	"github.com/cocov-ci/semgrep/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"testing"
)

type testHelper struct {
	l    *zap.Logger
	ctx  *mocks.MockContext
	exec *mocks.Mockexecutor
	ru   *runner
}

func newHelper(ctrl *gomock.Controller) *testHelper {
	exec := mocks.NewMockexecutor(ctrl)
	return &testHelper{
		l:    zap.NewNop(),
		ctx:  mocks.NewMockContext(ctrl),
		exec: exec,
		ru:   newRunner(exec),
	}
}

func (h *testHelper) start() ([]*result, error) {
	return h.ru.run(h.ctx, h.l)
}

func (h *testHelper) createFixtureYaml(t *testing.T, dirPath string) string {
	fileName := "semgrep.yaml"
	fPath := filepath.Join(dirPath, fileName)
	_, err := os.Create(fPath)
	require.NoError(t, err)

	return fPath
}

func TestRun(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	wd = findParentDir(t, wd, "semgrep")

	boom := errors.New("boom")

	t.Run("Fails looking for semgrep configuration files", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		helper := newHelper(ctrl)

		falsePath := path.Join(wd, "false-path")
		helper.ctx.EXPECT().Workdir().Return(falsePath)

		_, err := helper.start()
		require.Error(t, err)
	})

	t.Run("Single file", func(t *testing.T) {
		singlePath := filepath.Join(wd, "mocks")

		t.Run("Fails running semgrep", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			helper := newHelper(ctrl)

			helper.ctx.EXPECT().Workdir().Return(singlePath)

			sOut := []byte("std output")
			sErr := []byte("std err")
			opts := &cocov.ExecOpts{Workdir: singlePath}
			args := newArgs(noYaml, singlePath)
			helper.exec.EXPECT().Exec2(cmd, args, opts).
				Return(sOut, sErr, boom)

			_, err = helper.start()
			require.Error(t, err)

			expectedErr := runErr(singlePath, args, sOut, sErr, boom)
			assert.EqualError(t, err, expectedErr.Error())
		})

		t.Run("Fails decoding output", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			helper := newHelper(ctrl)

			helper.ctx.EXPECT().Workdir().Return(singlePath)

			opts := &cocov.ExecOpts{Workdir: singlePath}
			args := newArgs(noYaml, singlePath)
			helper.exec.EXPECT().Exec2(cmd, args, opts).
				Return(badOutput(), nil, nil)

			_, err = helper.start()
			require.Error(t, err)

			require.ErrorContains(t, err, "json")
		})

		t.Run("Works as expected without a root file", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			helper := newHelper(ctrl)

			helper.ctx.EXPECT().Workdir().Return(singlePath)

			cat := performance
			filePath := path.Join(singlePath, "file.go")
			output := expectedOutput(t, filePath, cat, 3)

			opts := &cocov.ExecOpts{Workdir: singlePath}
			args := newArgs(noYaml, singlePath)
			helper.exec.EXPECT().Exec2(cmd, args, opts).
				Return(output, nil, nil)

			res, err := helper.start()
			require.NoError(t, err)
			require.NotNil(t, res)
		})

		t.Run("Works as expected with a root file", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			helper := newHelper(ctrl)

			helper.ctx.EXPECT().Workdir().Return(singlePath)

			fixtureYaml := helper.createFixtureYaml(t, singlePath)
			t.Cleanup(func() { _ = os.Remove(fixtureYaml) })

			cat := performance
			filePath := path.Join(singlePath, "file.go")
			output := expectedOutput(t, filePath, cat, 3)

			opts := &cocov.ExecOpts{Workdir: singlePath}
			args := newArgs(fixtureYaml, singlePath)
			helper.exec.EXPECT().Exec2(cmd, args, opts).
				Return(output, nil, nil)

			res, err := helper.start()
			require.NoError(t, err)
			require.NotNil(t, res)
		})
	})

	t.Run("Parallel", func(t *testing.T) {

	})
}

func badOutput() []byte {
	return []byte{234}
}

func expectedOutput(t *testing.T, filePath, category string, numResults int) []byte {
	res := make([]*result, 0, numResults)
	for i := 0; i <= numResults; i++ {
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
