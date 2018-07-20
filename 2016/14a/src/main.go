package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"

	"logger"
	"pad"
)

var (
	salt    = flag.String("salt", "", "salt")
	verbose = flag.Bool("verbose", false, "verbose")
)

func main() {
	flag.Parse()
	logger.Init(*verbose)

	if *salt == "" {
		log.Fatal("--salt is required")
	}

	hashQueue := pad.NewQueue()

	finishIdx := math.MaxInt32
	keyIndexes := []int{}
	for i := 0; i < finishIdx; i++ {
		if i != 0 && i%1000000 == 0 {
			fmt.Println(i)
		}

		h := pad.MakeHash(*salt, i)

		if reps := pad.HasRepeats(h, 3); len(reps) > 0 {
			// We only consider the first one
			logger.LogF("%d: adding 3-rep to hash queue: %x, exp %v\n", i, reps[0], i+1000)
			hashQueue.Add(i, reps[0], i+1000)
		}

		// Off in the future, we want to verify whether the 3-reps have
		// corresponding 5-reps. We use ActiveBefore because we want to
		// exclude any added this iteration.
		if reps := pad.HasRepeats(h[:], 5); len(reps) > 0 {
			activeElems := hashQueue.ActiveBefore(i)
			logger.LogF("%d: found 5-rep %x active %v\n", i, h, activeElems)
			for _, activeElem := range activeElems {
				for _, rep := range reps {
					if activeElem.Repeater == rep {
						keyIndexes = append(keyIndexes, activeElem.Index)
						logger.LogF("adding key index %v for %v, now %d keys\n",
							activeElem.Index, rep, len(keyIndexes))
						hashQueue.Delete(activeElem)
					}
				}
			}
		}

		if finishIdx == math.MaxInt32 && len(keyIndexes) >= 64 {
			finishIdx = i + 1000
			fmt.Printf("resetting finishidx to %v\n", finishIdx)
		}

		hashQueue.ExpireTo(i)
	}

	sort.Ints(keyIndexes)

	for i, keyIndex := range keyIndexes {
		fmt.Printf("%3d: %v\n", i+1, keyIndex)
	}
}
