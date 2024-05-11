package lab

import (
    workertype "github.com/NCKU-NASA/nasa-judge-lib/enum/worker_type"
)

type worker struct {
    WorkerType workertype.WorkerType `yaml:"workertype" json:"workertype"`
    WorkerPool string `yaml:"workerpool" json:"workerpool"`
}
