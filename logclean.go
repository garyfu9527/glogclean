package glogclean

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

type LogfileRule struct {
	LognamePatter	string
	MaxFiles		int
}

var perGlogSize uint64 = 50;		//unit is MB
var stopTaskChan chan bool = make(chan bool, 0)
func init()  {
	flag.Uint64Var(&perGlogSize, "glogsize", 50, "per glog file size")
}

// 定时清理glog日志
func RunCleanLogTask() chan bool {
	glog.MaxSize = perGlogSize * 1024 * 1024	//每个日志文件最大50MB
	var logDir string
	f := flag.Lookup("log_dir")
	if f != nil {
		logDir = f.Value.String()
	} else {
		logDir = os.TempDir()
	}
	exeName := filepath.Base(os.Args[0])
	rules := []LogfileRule{
		{
			LognamePatter: exeName + "*.INFO.*",
			MaxFiles: 4,
		},
		{
			LognamePatter: exeName + "*.ERROR.*",
			MaxFiles: 2,
		},
		{
			LognamePatter: exeName + "*.WARNING.*",
			MaxFiles: 2,
		},
		{
			LognamePatter: exeName + "*.FATAL.*",
			MaxFiles: 5,
		},
	}

	return startCleanTask(logDir, rules, time.Minute)
}

// 停止清理
func StopTask()  {
	stopTaskChan <- true
}

// 按照规则定时清理日志文件
func startCleanTask(logPath string, rules []LogfileRule, timer time.Duration) chan bool {
	go func() {
		fmt.Println("StartCleanTask Start!")
		t := time.NewTicker(timer)
		for true {
			select {
				case <-t.C:
					runClean(logPath, rules)
				case <-stopTaskChan:
					fmt.Println("stop clean task")
					t.Stop()
					return
			}
		}
		fmt.Println("StartCleanTask End!")
	}()

	return stopTaskChan
}

// 执行真正的清理
func runClean(logPath string, rules []LogfileRule)  {
	files, err := ioutil.ReadDir(logPath)
	if err != nil {
		return
	}

	for _, rule := range rules {
		var matchFiles []os.FileInfo = make([]os.FileInfo, 0, 10)
		//找到匹配的文件名
		for _, item := range files{
			if item.IsDir() {
				continue
			}

			_, fileName := path.Split(item.Name())
			if ok, _ := path.Match(rule.LognamePatter, fileName); ok {
				matchFiles = append(matchFiles, item)
			}
		}

		// 按修改时间倒排序
		sort.Slice(matchFiles, func(i, j int) bool {
			return matchFiles[i].ModTime().Sub(matchFiles[j].ModTime()).Seconds() > 0
		})

		// 删除多余的文件
		for i:=rule.MaxFiles; i<len(matchFiles); i++ {
			os.Remove(path.Join(logPath, matchFiles[i].Name()))
		}
	}

}
