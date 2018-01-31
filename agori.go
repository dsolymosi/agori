package agori

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Agori struct {
	bt                        *BinTrie
	lru                       *LRU
	reclaimExemptionThreshold float32
	numNodes                  int
	numAdded                  int
	numSkip                   int
}

func NewAgoriD() *Agori {
	return NewAgori(256, 1.0/32.0)
}

func NewAgori(numNodes int, reclaimExemptionThreshold float32) *Agori {
	var a Agori
	a.bt = new(BinTrie)
	a.numNodes = numNodes
	a.lru = NewLRU(a.numNodes)
	a.reclaimExemptionThreshold = reclaimExemptionThreshold

	return &a
}

func (a *Agori) Print(thresh float32) {
	fmt.Printf("[dst address] %d (100.00%%)\n", a.numAdded)
	a.bt.PrintContents(a.numAdded, thresh)
	//see how much is in lru
	lrutot := 0
	for e := a.lru.l.Front(); e != nil; e = e.Next() {
		b, v := a.bt.Get(e.Value.(uint32), 0)
		if !b {
			panic("??")
		}
		lrutot += int(v)
	}
	fmt.Printf("%%LRU hits: %.2f%% (%d/%d)\n", 100.0*float32(lrutot)/float32(a.numAdded), lrutot, a.numAdded)
}

//insert IP by string representation
//eg. for 127.0.0.1, call InsertS("127.0.0.1")
func (a *Agori) InsertS(s string) error {
	ss := strings.Split(s, ".")
	if len(ss) == 4 {
		var d [4]uint8
		for i := 0; i < 4; i++ {
			d64, err := strconv.ParseUint(ss[i], 10, 8)
			d[i] = uint8(d64)
			if err != nil {
				return err
			}
		}

		a.InsertD(d[0], d[1], d[2], d[3])
		return nil
	}
	return errors.New("Bad IP string format!")
}

//insert IP by dot decimal notation
//eg. for 127.0.0.1, call InsertD(127, 0, 0, 1)
func (a *Agori) InsertD(d1, d2, d3, d4 uint8) {
	a.Insert(uint32(d1)<<24 + uint32(d2)<<16 + uint32(d3)<<8 + uint32(d4))
}

//insert IP by uint32 representation
func (a *Agori) Insert(k uint32) {
	a.numAdded++

	a.bt.Increment(k)
	if !a.lru.IsFull() {
		d, mip := a.lru.Add(k)
		if d {
			//we should never be here
			fmt.Println("Incorrectly deleting", mip)
			a.bt.Delete(mip)
		}
	} else {
		//find a good ip to delete
		endOffset := 0
		underThreshold := false
		var mip uint32
		for !underThreshold {
			mip = a.lru.GetEnd(endOffset)
			if b, v := a.bt.sumParent(mip, 0); b && v > uint(a.reclaimExemptionThreshold*float32(a.numAdded)) {
				endOffset += 1
			} else {
				underThreshold = true
			}
		}
		a.lru.Delete(mip)
		a.bt.Delete(mip)

		//add
		d, mip := a.lru.Add(k)
		if d {
			fmt.Println("Incorrectly deleting", mip)
			a.bt.Delete(mip)
		}
	}
}

//return whether an IP exists in the agori, and if it does, its value
func (a *Agori) Get(k uint32) (bool, uint) {
	return a.bt.Get(k, 0)
}

//return whether an IP exists in the agori, and if it does, its value
func (a *Agori) GetD(d1, d2, d3, d4 uint8) (bool, uint) {
	return a.Get(uint32(d1)<<24 + uint32(d2)<<16 + uint32(d3)<<8 + uint32(d4))
}

//return whether an IP exists in the agori, and if it does, its value
func (a *Agori) GetS(s string) (error, bool, uint) {
	ss := strings.Split(s, ".")
	if len(ss) == 4 {
		var d [4]uint8
		for i := 0; i < 4; i++ {
			d64, err := strconv.ParseUint(ss[i], 10, 8)
			d[i] = uint8(d64)
			if err != nil {
				return err, false, 0
			}
		}
		b, u := a.GetD(d[0], d[1], d[2], d[3])
		return nil, b, u
	}
	return errors.New("Bad IP string format!"), false, 0
}
