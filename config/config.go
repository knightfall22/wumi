package config

var Host string
var Port int
var KeyLimit int
var AOFFILE string

// will evict 40% of keys when eviction runs
var EvictionRatio float64 = 0.4

var EvictionStrategy string = "allkeys-random"
