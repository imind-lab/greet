package cron

import (
	"fmt"
	"time"
)

func (c Cron) EchoTime() {
	fmt.Println(time.Now())
}
