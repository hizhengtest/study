package service

import (
	"sync"
)

type baseService struct {
	mutex *sync.Mutex
}
