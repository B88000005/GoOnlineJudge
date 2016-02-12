package schedule

import (
	"GoOnlineJudge/config"
	"GoOnlineJudge/model"
	"time"
)

type RemoteOJInterface interface {
	Init()
	Host() string
	Ping() error
	GetProblems() error
}

var ROJs = []RemoteOJInterface{&PKUJudger{}}

func init() {
	go func() {
		ojModel := &model.OJModel{}
		status := &model.OJStatus{}
		for {
			for _, oj := range ROJs {
				err := oj.Ping()
				status.Name = oj.Host()
				if err != nil {
					status.Status = config.StatusUnavailable
					ojModel.Update(status)
				} else {
					status.Status = config.StatusOk
					ojModel.Update(status)
				}
			}
			time.Sleep(10 * time.Minute)
		}
	}()
}
