package bantu

import (
	"github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/db"
)

type JobManager struct {
	link      *db.Link
	scheduler *soffa.Scheduler
}
