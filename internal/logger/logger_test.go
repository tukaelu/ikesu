package logger

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	t.Run("returns logger when file is empty", func(t *testing.T) {
		logger, err := NewLogger("", "info", false)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("returns logger when dryrun is true", func(t *testing.T) {
		logger, err := NewLogger("tempfile.log", "info", true)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("returns error when file is a directory", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		assert.NoError(t, err)
		defer os.RemoveAll(dir)

		_, err = NewLogger(dir, "info", false)
		assert.Error(t, err)
		assert.Equal(t, "file is directory", err.Error())
	})

	t.Run("creates new file when file does not exist", func(t *testing.T) {
		file := "tempfile.log"
		defer os.Remove(file)

		logger, err := NewLogger(file, "info", false)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		_, err = os.Stat(file)
		assert.NoError(t, err)
	})

	t.Run("opens existing file for appending when file already exists", func(t *testing.T) {
		file := "tempfile.log"
		defer os.Remove(file)

		// Create the file first
		f, err := os.Create(file)
		assert.NoError(t, err)
		f.Close()

		logger, err := NewLogger(file, "info", false)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})
}
