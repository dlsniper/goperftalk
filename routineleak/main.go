// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0
// Original content at https://github.com/ardanlabs/gotraining/topics/go/profiling/godebug/godebug.go.

package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

type User struct {
	ID          int
	Username    string
	JobTitle    string
	Email       string
	Description string
}

var (
	leakedRoutines      uint64
	leakedSlowRoutines  uint64
	leakedMemRoutines   uint64
	leakedRoutinesGuard = &sync.RWMutex{}

	chars = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz~!@#$%€^&*()_-+={[}]|\:;"'<,>.?/'"`
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/details", details)
	http.HandleFunc("/mem", leakerMem)
	http.HandleFunc("/", leakerSlow)

	log.Println("server starting")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func leakerSlow(w http.ResponseWriter, r *http.Request) {
	message := `{"ID":1,"Username":"dlsniper","jobtitle":"Developer Advocate"}`
	if rand.Intn(5) == 3 {
		leakedRoutinesGuard.Lock()
		leakedRoutines++
		leakedSlowRoutines++
		leakedRoutinesGuard.Unlock()

		message = `{"ID":0,"Username":"leakerSlow","jobtitle":"Resource Leaker"}`
		go func() {
			// Pretend to do some heavy work here, like talking to a database
			for {
				time.Sleep(time.Millisecond * 10)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(message))
}

func genText(maxChar int) string {
	size := rand.Intn(maxChar)
	result := ""
	for i := 0; i < size; i++ {
		result += string(chars[rand.Intn(len(chars))])
	}

	return result
}

func leakerMem(w http.ResponseWriter, r *http.Request) {
	message := `{"ID":1,"Username":"dlsniper","jobtitle":"Developer Advocate"}`
	if rand.Intn(5) == 3 {
		leakedRoutinesGuard.Lock()
		leakedRoutines++
		leakedMemRoutines++
		leakedRoutinesGuard.Unlock()

		message = `{"ID":0,"Username":"leakerMem","jobtitle":"Memory Leaker"}`
		go func() {
			var everGrowing []User
			// Pretend to do some heavy work here, like talking to a database
			for {
				everGrowing = append(everGrowing, User{
					ID:          0,
					Username:    genText(30),
					JobTitle:    genText(50),
					Email:       genText(40),
					Description: genText(400),
				})
				time.Sleep(time.Millisecond * 250)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(message))
}

func details(w http.ResponseWriter, r *http.Request) {
	leakedRoutinesGuard.RLock()
	message := struct {
		LeakedRoutines     uint64 `json:"leakedRoutines"`
		LeakedSlowRoutines uint64 `json:"slowRoutines"`
		LeakedMemRoutines  uint64 `json:"memRoutines"`
	}{
		LeakedRoutines:     leakedRoutines,
		LeakedSlowRoutines: leakedSlowRoutines,
		LeakedMemRoutines:  leakedMemRoutines,
	}
	leakedRoutinesGuard.RUnlock()

	w.WriteHeader(http.StatusOK)
	m, _ := json.Marshal(message)
	_, _ = w.Write(m)
}

//
// Useful to see pressure on heap over time.
// -alloc_space  : All allocations happened since program start  	** default
//                 go tool pprof --alloc_objects leaker.exe http://localhost:8080/debug/pprof/heap
// -alloc_objects: Number of object allocated at the time of profile
//                 go tool pprof --alloc_objects leaker.exe http://localhost:8080/debug/pprof/heap
//
// Useful to see current status of heap.
// -inuse_space  : Allocations live at the time of profile
//                 go tool pprof --inuse_space leaker.exe http://localhost:8080/debug/pprof/heap
// -inuse_objects: Number of bytes allocated at the time of profile
//                 go tool pprof --inuse_objects leaker.exe http://localhost:8080/debug/pprof/heap
//
