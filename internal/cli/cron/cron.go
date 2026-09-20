package cron

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

const Flag = "cron"

func Execute(ctx context.Context, schedule string, f func() error) error {
	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	errPrt := new(error)
	_, err = s.NewJob(
		gocron.CronJob(trim(schedule), false),
		gocron.NewTask(func() { *errPrt = f() }),
		gocron.WithSingletonMode(gocron.LimitModeReschedule))
	if err != nil {
		return err
	}
	s.Start()
	<-ctx.Done()
	s.Shutdown()
	return *errPrt
}

func Validate(schedule string) error {
	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	_, err = s.NewJob(
		gocron.CronJob(trim(schedule), false),
		gocron.NewTask(func() {}))
	if err != nil {
		return err
	}
	s.Shutdown()
	return nil
}

func trim(cron string) string {
	if len(cron) > 1 {
		switch byte((cron)[0]) {
		case '\'':
			return subTrim(cron, '\'')
		case '"':
			return subTrim(cron, '"')
		case '`':
			return subTrim(cron, '`')
		}
	}
	return cron
}

func subTrim(cron string, b byte) string {
	if byte((cron)[len(cron)-1]) == b {
		return (cron)[1 : len(cron)-1]
	}
	return cron
}
