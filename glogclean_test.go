package glogclean

import (
	"flag"
	"github.com/golang/glog"
	"testing"
	"time"
)

func TestGlogClean(t *testing.T)  {
	flag.Parse()
	flag.Set("log_dir", "/tmp")

	// 定时清理日志
	RunCleanLogTask()
	defer StopTask()

	for  {
		glog.Infof("glog clean test %d", time.Now().Unix())
		time.Sleep(time.Microsecond)
	}
}
