package vtime

import (
	"github.com/pion/transport/v3/xtime"
)

var _ xtime.TimeManager = &Simulator{}
