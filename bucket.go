package work_sync

type WorkFunc[Id comparable, State any] func(state *State)

type BucketWorkReq[Id comparable, State any] struct {
	// each worker will only handle one request at time
	Bucket Id
	F WorkFunc[Id, State]
}

type BucketData[Id comparable, State any] struct {
	work chan WorkFunc[Id, State]
	jobs *uint
}

// Handle all requests on w until it's closed.
// Creating a custom BucketRunner allows managing extra resources per bucket.
type BucketRunner[Id comparable, State any] func (id Id, w chan WorkFunc[Id, State])

func defaultBucketRunner[Id comparable, State any](id Id, w chan WorkFunc[Id, State]) {
	var state State

	for f := range w {
		f(&state)
	}
}

func handleBucketWorkReq[Id comparable, State any](done chan Id, buckets map[Id]BucketData[Id, State], runner BucketRunner[Id, State], req BucketWorkReq[Id, State]) {
	n, ok := buckets[req.Bucket]
	if ! ok {
		n.work = make(chan WorkFunc[Id, State], 4)
		jobs := uint(0)
		n.jobs = &jobs

		buckets[req.Bucket] = n

		go runner(req.Bucket, n.work)
	}

	*n.jobs = *n.jobs + 1
	n.work <- func (state *State) {
		defer func() {
			done <- req.Bucket
		}()
		req.F(state)
	}
}

func handleDone[Id comparable, State any](buckets map[Id]BucketData[Id, State], bId Id) {
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
func Bucket[Id comparable, State any](c <-chan BucketWorkReq[Id, State], runner BucketRunner[Id, State]) {
	buckets := make(map[Id]BucketData[Id, State])
	done := make(chan Id)
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

func DefaultBucket[Id comparable, State any](c <-chan BucketWorkReq[Id, State]) {
	Bucket(c, defaultBucketRunner[Id, State])
}

func SpawnDefaultBucket[Id comparable, State any]() (chan<- BucketWorkReq[Id, State]) {
	c := make(chan BucketWorkReq[Id, State])
	go DefaultBucket(c)
	return c
}
