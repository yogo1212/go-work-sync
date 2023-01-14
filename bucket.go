package work_sync

// caller chooses value
type BucketId uint64

type WorkFunc func()

type BucketWorkReq struct {
	// each worker will only handle one request at time
	Bucket BucketId
	F WorkFunc
}

type BucketData struct {
	work chan WorkFunc
	jobs *uint
}

// Handle all requests on w until it's closed.
// Creating a custom BucketRunner allows managing extra resources per bucket.
type BucketRunner func (id BucketId, w chan WorkFunc)

func defaultBucketRunner(id BucketId, w chan WorkFunc) {
	for f := range w {
		f()
	}
}

func handleBucketWorkReq(done chan BucketId, buckets map[BucketId]BucketData, runner BucketRunner, req BucketWorkReq) {
	n, ok := buckets[req.Bucket]
	if ! ok {
		n.work = make(chan WorkFunc, 4)
		jobs := uint(0)
		n.jobs = &jobs

		buckets[req.Bucket] = n

		go runner(req.Bucket, n.work)
	}

	*n.jobs = *n.jobs + 1
	n.work <- func () {
		defer func() {
			done <- req.Bucket
		}()
		req.F()
	}
}

func handleDone(buckets map[BucketId]BucketData, bId BucketId) {
	n, ok := buckets[bId]
	if ! ok {
		return
	}

	*n.jobs = *n.jobs - 1

	if *n.jobs > 0 {
		return
	}

	close(n.work)
	delete(buckets, bId)
}

// Handles requests from c until closed.
// Waits for all workers to finish before returning.
func Bucket(c chan BucketWorkReq, runner BucketRunner) {
	buckets := make(map[BucketId]BucketData)
	done := make(chan BucketId)
	defer close(done)

	for {
		select {
		case req, ok := <-c:
			if ! ok {
				goto finish
			}

			handleBucketWorkReq(done, buckets, runner, req)
		case bId := <-done:
			handleDone(buckets, bId)
		}
	}

finish:
	if len(buckets) == 0 {
		return
	}

	for bId := range done {
		handleDone(buckets, bId)

		if len(buckets) == 0 {
			return
		}
	}
}

func DefaultBucket(c chan BucketWorkReq) {
	Bucket(c, defaultBucketRunner)
}
