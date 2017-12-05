package growler

import (
	"regexp"
	"sync"
	"time"
)

// Growler manages the workers
type Growler struct {
	sync.RWMutex
	log         bool
	start       *Location
	workers     int
	match       *regexp.Regexp
	Matches     map[string]*Location
	queue       map[string]*Location
	previous    map[string]*Location
	Visits      int
	waitDefault time.Duration
	wait        map[string]time.Duration // Location.Host instead of string, maybe?
	ignore      map[string]bool
}

var (
	run     bool
	workers []*worker
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// NewGrowler initializes a new growler
func NewGrowler(start string, log bool, workers int, match *regexp.Regexp, wait time.Duration) *Growler {
	growl := &Growler{
		log:         log,
		start:       deconstructURL(start),
		workers:     workers,
		match:       match,
		waitDefault: wait,
		Matches:     make(map[string]*Location, 0),
		queue:       make(map[string]*Location, 0),
		previous:    make(map[string]*Location, 0),
		wait:        make(map[string]time.Duration, 0),
		ignore:      make(map[string]bool, 0),
	}

	growl.queue[start] = growl.start

	return growl
}

// Start starts the growling
func (growl *Growler) Start() {
	run = true

	// (de)append if workers has been decreased/increased otherwise start all
	for i := 0; i < growl.workers; i++ {
		worker := &worker{
			run: run,
		}

		go worker.work(growl)

		workers = append(workers, worker)
	}
}

// Stop stops the growling
func (growl *Growler) Stop() {
	run = false

	for i := 0; i < len(workers); i++ {
		workers[i].run = false
	}
}

// Present the results found
func (growl *Growler) Present() string {
	str := ""

	str += "Found these matches:"
	for k := range growl.Matches {
		str += growl.Matches[k].Source + "\n"
	}
	str += string(len(growl.Matches))
	str += "\n"
	str += "Total visit(s): "
	str += string(growl.Visits)

	return str
}
