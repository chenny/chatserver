package report

import (
	"github.com/gansidui/chatserver/config"
	"testing"
)

func TestReport(t *testing.T) {
	config.ReadIniFile("../config.ini")

	AddCount(TryConnect, 5)
	AddCount(TryConnect, 3)
	AddCount(SuccessConnect, 2)
	AddCount(SuccessConnect, 1)
	AddCount(OnlineMsg, 5)
	AddCount(OfflineMsg, 2)
	AddCount(OnlineUser, 6)
	AddCount(OnlineUser, -2)

	Work()
}
