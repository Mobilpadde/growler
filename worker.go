package growler

import (
	"log"
	"net/http"
	"time"
)

type worker struct {
	current *Location
	run     bool
	visits  int
	num     int
}

func (worker *worker) work(growl *Growler) {
	growl.RLock()
	for len(growl.queue) < 1 {
		time.Sleep(growl.waitDefault)
	}
	growl.RUnlock()

	for k := range growl.queue {
		growl.RLock()
		worker.current = growl.queue[k]
		growl.RUnlock()

		growl.Lock()
		growl.previous[k] = growl.queue[k]
		delete(growl.queue, k)
		growl.Unlock()
		break
	}

	if worker.run {
		if !growl.isAllowed(worker.current.Host, worker.current.Path) {
			worker.work(growl)
			return
		}

		res, err := http.Get(worker.current.Source)
		checkErr(err)

		defer res.Body.Close()

		locs, isMatch := growl.find(res.Body)
		for i := 0; i < len(locs); i++ {
			growl.RLock()

			if growl.queue[locs[i]] == nil && growl.previous[locs[i]] == nil { // Isn't in queue and haven't been before
				growl.queue[locs[i]] = deconstructURL(locs[i])
			}

			growl.RUnlock()
		}

		if growl.log {
			log.Printf("Visited \"%s\" and enqueued %d new URL(s), which adds up to a total of %d.\n", worker.current.Source, len(locs), len(growl.queue))
		}

		if isMatch {
			growl.Lock()
			growl.Matches[worker.current.Source] = worker.current
			growl.Unlock()
		}

		worker.visits++
		growl.Visits++

		if len(growl.queue) > 0 {
			if growl.wait[worker.current.Host] > time.Duration(0) {
				time.Sleep(growl.wait[worker.current.Host])
				worker.work(growl)
			} else {
				time.Sleep(growl.waitDefault)
				worker.work(growl)
			}
		} else if growl.log {
			log.Printf("Worker #%d has been terminated with %d visit(s).", worker.num, worker.visits)
		}
	} else {
		if growl.log {
			log.Printf("Worker #%d has been terminated with %d visit(s).", worker.num, worker.visits)
		}
	}

	return
}
