package entity

import (
	"sync"
	"time"
	"github.com/patrickmn/go-cache"
)

const RegistryURL = "http://localhost:5000/v2"

type ContainerInfo struct {
	Layers int `json:"layers"`
	Size   int `json:"size"`
}

var ContainerCache = cache.New(5*time.Minute, 10*time.Minute)
var IpRequestTimes = make(map[string]time.Time)
var Mu sync.Mutex
