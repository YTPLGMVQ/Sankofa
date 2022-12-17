// interface for persistent score database, indexed by Oware position ranks
//
// # DESIGN, TACTICS AND HACKS
//
// Definitions:
//   - scores are signed bytes | int8
//   - ranks are signed integers | int64
//   - the file size is 1_399_358_844_974 bytes plus one more for the state and minus the size of level-47
//   - SCOREs ∈[-48, 48]; actually: ∈[-level, level].
//   - -49 is the initial value, meaning "uninitialized/unreachabel"
//   - the database lacks locality. caching/mmap/ramdisk is useless for higher levels.
//
// Score life cycle: uninitialized⇢0⇢retrograde
//
// Level processing life cycle:
//   - upwards from level 0: retrograde analysis until no changes are possible
//   - skip level 47
//
// # CAVEATS
//
// The >1TB database file is huge for the current storage size.
// Creating it may fill up the file system or otherwise crash your system and render it unbootable.
// BUILD A DATABASE AT YOUR OWN RESPONSIBILITY AND ONLY IF YOU KNOW WHAT YOU ARE DOING!
//
// The database is used to implement Romein's retrograde analysis of Awari.
// Awari and Oware are close enough to be playable in an almost identical way.
// When a cycle is found, Awari splits the stones on the table half-half and may yield fractional scores.
// Oware gives each player the stones on his side when a cycle is found.
// The same position might be in the middle or at the end of a cycle, dependent on how the players got there.
// For this reason, a given position might be evaluated to several scores.
// It is either the split-score, or it is computed backwards from where the cycle ends.
// The conclusion might be that the Oware state space is much larger than thought
// and that the game is not solvable with the current strategy. Awari is a very good approximation.
package db

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"sankofa/ow"
	"sync"
)

// singleton, thus global variable
var FileName string
var file *os.File
var isOpen bool

const OFFSET = int8(49)

// synchronization
var mutex sync.RWMutex // data access

func init() {
	// default path to database file
	user, err := user.Current()
	ow.Check(err)
	FileName = path.Clean(path.Join(user.HomeDir, "oware.db"))
}

// panics on error
func Open() {
	mutex.Lock()
	defer mutex.Unlock()

	ow.Log("open database:", FileName)
	var err error

	file, err = os.OpenFile(FileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		// the database is optional for sankofa
		fmt.Println("could not open database file:", FileName, ":", err)
		return
	}
	isOpen = true
	ow.Log("fd:", file.Fd())
}

func Close() {
	mutex.Lock()
	defer mutex.Unlock()

	if isOpen {
		ow.Log("close database")
		ow.Check(file.Close())
		isOpen = false
	} else {
		ow.Log("nothing to close")
	}
}

func IsOpen() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	r := isOpen

	return r
}

// the initial value (zero) maps to -48 and means the start of the marking phase
func SetState(state int8) {
	if !isOpen {
		ow.Panic("cannot write in a closed database")
	}

	if state < 0 || state > 48 || state == 47 {
		ow.Panic("state out of range:", state)
	}

	mutex.Lock()
	defer mutex.Unlock()

	r := make([]byte, 1, 1)
	r[0] = byte(state)
	n, err := file.WriteAt(r, 0)
	ow.Check(err)
	if n != 1 {
		ow.Panic(n)
	}
}

func GetState() int8 {
	if !isOpen {
		ow.Log("database closed: default")
		return 0
	}

	mutex.RLock()
	defer mutex.RUnlock()

	r := make([]byte, 1, 1)
	n, err := file.ReadAt(r, 0)
	if err == io.EOF {
		// read past EOF: return default value
		return 0
	} else {
		ow.Check(err)
		if n != 1 {
			ow.Panic(n)
		}
	}

	return int8(r[0])
}

// the initial value (zero) maps to -47 and means unreachable
func SetScore(rank int64, score int8) {
	if !isOpen {
		ow.Panic("cannot write in a closed database")
	}

	level := ow.Level(rank)
	if level == -47 || level == 47 {
		ow.Panic("level out of range:", level)
	}
	// skip unused level-47
	if rank >= ow.LevelUpperLimits[47] {
		rank = rank - ow.LevelUpperLimits[47] + ow.LevelUpperLimits[46]
	}

	// -OFFSET is the default, initial-value of the database
	if score == -OFFSET {
		ow.Log("skip NaN")
		return
	}
	if score < -level || score > level {
		ow.Panic("score out of range: level:", level, "score:", score)
	}
	score += OFFSET

	mutex.Lock()
	defer mutex.Unlock()

	r := make([]byte, 1, 1)
	r[0] = byte(score)
	n, err := file.WriteAt(r, rank+1)
	ow.Check(err)
	if n != 1 {
		ow.Panic(n)
	}
}

// returns the score and true if initialized
func GetScore(rank int64) (int8, bool) {
	level := ow.Level(rank)
	if level == -47 || level == 47 {
		ow.Panic("level out of range:", level)
	}

	if !isOpen {
		ow.Log("database closed: default")
		return -OFFSET, false
	}

	// skip unused level-47
	if rank >= ow.LevelUpperLimits[47] {
		rank = rank - ow.LevelUpperLimits[47] + ow.LevelUpperLimits[46]
	}

	mutex.RLock()
	defer mutex.RUnlock()

	r := make([]byte, 1, 1)
	n, err := file.ReadAt(r, rank+1)
	if err == io.EOF {
		// read past EOF: return default value
		return -OFFSET, false
	} else {
		ow.Check(err)
		if n != 1 {
			ow.Panic(n)
		}
	}

	score := int8(r[0]) - OFFSET
	if score == -OFFSET {
		// default to zero
		return score, false
	} else {
		return score, true
	}
}
