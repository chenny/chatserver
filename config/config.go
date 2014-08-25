package config

import (
	"fmt"
	"github.com/robfig/config"
)

var (
	Addr                           string
	NumCpu                         int
	AcceptTimeout                  int
	ReadTimeout                    int
	WriteTimeout                   int
	C2COfflinemsgRedisAddr         string
	C2COfflinemsgRedisKeyExpireDay string
	LogFile                        string
	EmailServerAddr                string
	EmailServerPort                string
	EmailAccount                   string
	EmailPassword                  string
	EmailToList                    string
	EmailDuration                  int
)

func ReadIniFile(inifile string) error {
	conf, err := config.ReadDefault(inifile)
	if err != nil {
		return fmt.Errorf("Read %v error. %v", inifile, err.Error())
	}

	Addr, _ = conf.String("service", "addr")
	NumCpu, _ = conf.Int("service", "num_cpu")
	AcceptTimeout, _ = conf.Int("service", "accept_timeout")
	ReadTimeout, _ = conf.Int("service", "read_timeout")
	WriteTimeout, _ = conf.Int("service", "write_timeout")

	C2COfflinemsgRedisAddr, _ = conf.String("redisdb", "c2c_offlinemsg_redisAddr")
	C2COfflinemsgRedisKeyExpireDay, _ = conf.String("redisdb", "c2c_offlinemsg_redisKeyExpireDay")

	LogFile, _ = conf.String("debug", "logfile")

	EmailServerAddr, _ = conf.String("email", "email_server_addr")
	EmailServerPort, _ = conf.String("email", "email_server_port")
	EmailAccount, _ = conf.String("email", "email_account")
	EmailPassword, _ = conf.String("email", "email_password")
	EmailToList, _ = conf.String("email", "email_tolist")
	EmailDuration, _ = conf.Int("email", "email_duration")

	return nil
}
