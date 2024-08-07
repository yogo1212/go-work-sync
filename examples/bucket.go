package main

import (
	"fmt"

	work_sync "github.com/yogo1212/go-work-sync"
)

// A work_sync.Bucket ensures that each element of charCount is only accessed
// once at a time and prevents a race condition when incrementing.
// This is a showcase, not an efficient way to count bytes :-)

var (
	charCount [256]uint
	c chan work_sync.BucketWorkReq[byte, struct{}]
)


func countBytes(b []byte, done chan bool) {
	for _, i := range b {
		c <- work_sync.BucketWorkReq[byte, struct{}]{
			i,
			func (_s *struct{}) {
				charCount[i] = charCount[i] + 1
			},
		}
	}

	done <- true
}

const (
	loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	finibusBonorum = `At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident, similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga. Et harum quidem rerum facilis est et expedita distinctio. Nam libero tempore, cum soluta nobis est eligendi optio cumque nihil impedit quo minus id quod maxime placeat facere possimus, omnis voluptas assumenda est, omnis dolor repellendus. Temporibus autem quibusdam et aut officiis debitis aut rerum necessitatibus saepe eveniet ut et voluptates repudiandae sint et molestiae non recusandae. Itaque earum rerum hic tenetur a sapiente delectus, ut aut reiciendis voluptatibus maiores alias consequatur aut perferendis doloribus asperiores repellat.`
)

func addWorkAndClose() {
	defer close(c)

	// wait before calling close on c
	done := make(chan bool)
	defer close(done)

	go countBytes([]byte(loremIpsum), done)
	go countBytes([]byte(finibusBonorum), done)

	<-done
	<-done
}

func main() {
	c = make(chan work_sync.BucketWorkReq[byte, struct{}])

	go addWorkAndClose()

	work_sync.DefaultBucket(c)

	for char, count := range charCount {
		fmt.Printf("%q:\t%d\n", char, count)
	}
}
