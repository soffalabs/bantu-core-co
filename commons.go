package bantu

import (
	"fmt"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/soffa-io/soffa-core-go/db"
	"github.com/soffa-io/soffa-core-go/h"
	"github.com/soffa-io/soffa-core-go/log"
	"gorm.io/gorm"
	"time"
)

type Job struct {
	Id        string    `json:"id"`
	Event     string    `json:"event"`
	Payload   string    `json:"payload"`
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	Retries   int       `json:"retries"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobProcessor interface {
	Process(job *Job)
}

type JobRepo struct {
	link *db.Link
}

func NewJobRepo(link *db.Link) *JobRepo {
	return &JobRepo{link: link}
}

func (r *JobRepo) Transactional(cb func(repo JobRepo)) {
	r.link.Transactional(func(link *db.Link) {
		cb(JobRepo{link: link})
	})
}

func (r *JobRepo) CatchErr(job Job) {
	if err := recover(); err != nil {
		job.Retry((err.(error)).Error())
		r.Update(job)
	}
}

func (r *JobRepo) Update(job Job) {
	job.UpdatedAt = time.Now()
	r.link.Save(job)
}

func (r *JobRepo) Create(event string, payload interface{}, status string) Job {
	job := Job{
		Id:        h.NewUniqueIdP("job_"),
		Event:     event,
		Payload:   h.ToJsonStrSafe(payload),
		Status:    status,
		Retries:   0,
		CreatedAt: time.Time{},
	}
	r.link.Create(&job)
	return job
}

func (j *Job) Success() {
	j.Status = "success"
	j.Error = ""
}

func (j *Job) Fail(msg string) {
	j.Status = "failed"
	j.Error = msg
	log.Default.Error("a job as failed: %s -- %s", j.Id, msg)
}

func (j *Job) Failf(format string, args ...interface{}) {
	j.Failf(fmt.Sprintf(format, args...))
}

func (j *Job) Retry(msg string) {
	j.Status = "pending"
	j.Error = msg
	j.Retries += 1
	if j.Retries >= 10 {
		j.Fail(msg)
		j.Status = "exhausted"
	} else {
		log.Default.Warnf("a job is going to be retried: %s -- %s", j.Id, msg)
	}
}

func (j *Job) GetPayload(dest interface{}) {
	if !h.IsEmpty(j.Payload) {
		if err := h.FromJsonStr(j.Payload, dest); err != nil {
			log.Default.Wrap(err, "error deserializaing job data. Is it a json ?")
		}
	}
}

func (r *JobRepo) Poll() []Job {
	var jobs []Job
	r.link.Find(&jobs, db.Q().Limit(100).W(h.Map{"status": "pending"}).Sort("created_at"))
	return jobs
}

func WithJobMigrations(migrations []*gormigrate.Migration) []*gormigrate.Migration {
	mig := []*gormigrate.Migration{{
		ID: "20210828_CreateJobs",
		Migrate: func(tx *gorm.DB) error {
			type Job struct {
				Id        string `gorm:"primaryKey;not null"`
				Event     string `gorm:"not null"`
				Payload   string
				Status    string `gorm:"not null"`
				Error     *string
				Retries   *int
				CreatedAt time.Time
				UpdatedAt *time.Time
			}
			return tx.Migrator().CreateTable(&Job{})
		},
	}}
	return append(migrations, mig[:]...)
}
