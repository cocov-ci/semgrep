package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/cocov-ci/go-plugin-kit/cocov"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
		t.Run("Works as expected", func(t *testing.T) {
			rootPath := filepath.Join(wd, "plugin", "fixtures")
			rootYaml := filepath.Join(rootPath, "semgrep.yaml")

			configs, err := findYamlRecursive(rootPath)
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			helper := newHelper(ctrl)

			rootArgs := newArgs(rootYaml, rootPath)
			jobs := buildJobs(configs, rootPath, rootArgs)

			helper.ctx.EXPECT().Workdir().Return(rootPath)

			issuesPerProject := 3

			for _, j := range jobs {
				cat := performance
				filePath := filepath.Join(j.path, "file.go")
				output := expectedOutput(t, filePath, cat, issuesPerProject)
				opts := &cocov.ExecOpts{Workdir: j.path}
				helper.exec.EXPECT().Exec2("semgrep", j.args, opts).
					Return(output, nil, nil)
			}

			res, err := helper.start()
			require.NoError(t, err)
			assert.NotNil(t, res)
			expectedResults := len(configs) * issuesPerProject
			ok := len(res) == expectedResults
			assert.Truef(t, ok,
				"Matches the number of expected results.\nCurrent: %d / Expected: %d",
				len(res), expectedResults,
			)
		})
	})
}

func TestBuildJobs(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	wd = findParentDir(t, wd, "semgrep")

	t.Run("Works as expected", func(t *testing.T) {
		rootPath := filepath.Join(wd, "plugin", "fixtures")
		rootYaml := filepath.Join(rootPath, "semgrep.yaml")

		configs, err := findYamlRecursive(rootPath)
		require.NoError(t, err)

		rootArgs := newArgs(rootYaml, rootPath)

		jobs := buildJobs(configs, rootPath, rootArgs)
		// nolint innefassing
		entries, err := os.ReadDir(rootPath)

		// should not count test-no-yaml folder
		assert.Equal(t, len(entries)-1, len(jobs))

		for _, j := range jobs {
			f, err := os.Stat(j.path)
			require.NoError(t, err)
			assert.True(t, f.IsDir())
		}
	})
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
