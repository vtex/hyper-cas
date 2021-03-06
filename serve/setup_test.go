package serve

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/vtex/hyper-cas/utils"
)

func TestMain(m *testing.M) {
	utils.SetTestStorage()
	viper.Set("file.enableLocks", true)
	viper.Set("file.lockTimeoutMs", 100)
	os.Exit(m.Run())
}
