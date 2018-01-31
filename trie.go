package agori

import (
	"fmt"
	"strconv"
)

type BinTrie struct {
	p, c0, c1 *BinTrie

	//the possibly multi-bit key
	key uint32
	//length of key in bits (key is between 0 and 2^keylen - 1)
	//ie. highest bit is 1<<(keylen-1)
	keylen uint8

	//value stored
	val uint
}

//increments k
func (bt *BinTrie) Increment(k uint32) {

	complen := -1
	for i := 31; i >= 0; i-- {

		//if we have exhausted bt.key (matched all of it)
		if complen <= 0 {
			if k&(1<<uint(i)) == 0 {
				if bt.c0 == nil {
					//create new left child all the way to end
					//this should only happen if we disagree at the root
					bt.c0 = new(BinTrie)
					bt.c0.p = bt
					bt.c0.key = k
					if i < 31 {
						//Inserted onto non-root leaf! - this should not happen
						bt.c0.key %= 1 << (uint(i) + 1)
					}
					bt.c0.keylen = uint8(i) + 1
					bt.c0.val = 1
					return
				}
				bt = bt.c0
				complen = int(bt.keylen)
			} else {
				if bt.c1 == nil {
					//create new right child all the way to end
					//this should only happen if we disagree at the root
					bt.c1 = new(BinTrie)
					bt.c1.p = bt
					bt.c1.key = k
					if i < 31 {
						//Inserted onto non-root leaf! - this should not happen
						bt.c1.key %= 1 << (uint(i) + 1)
					}
					bt.c1.keylen = uint8(i) + 1
					bt.c1.val = 1
					return
				}
				bt = bt.c1
				complen = int(bt.keylen)
			}
		}

		//if we dont match
		if (bt.key&(1<<(uint(complen)-1)) == 0 && k&(1<<uint(i)) != 0) || (bt.key&(1<<(uint(complen)-1)) != 0 && k&(1<<uint(i)) == 0) {
			//create new node between bt.p and bt
			nbt := new(BinTrie)
			nbt.p = bt.p

			//new child of nbt
			nbtc := new(BinTrie)
			nbtc.p = nbt
			//pick the correct side for bt
			if bt.key&(1<<(uint(complen)-1)) == 0 {
				nbt.c0 = bt
				nbt.c1 = nbtc
			} else {
				nbt.c1 = bt
				nbt.c0 = nbtc
			}
			bt.p = nbt

			//update parent of nbt
			if bt.key&(1<<(bt.keylen-1)) == 0 {
				nbt.p.c0 = nbt
			} else {
				nbt.p.c1 = nbt
			}

			//split the string between nbt and bt
			//note that bt.keylen + 1 >= complen guaranteed
			nbt.key = bt.key >> uint(complen)
			nbt.keylen = bt.keylen - uint8(complen)
			//just in case
			if complen < 32 {
				//take right side of key
				bt.key = bt.key % (1 << uint(complen))
			}
			bt.keylen = uint8(complen)

			//add rest of k as key for nbtc
			nbtc.key = k
			if i < 31 {
				nbtc.key = nbtc.key % (1 << (uint(i) + 1))
			}
			nbtc.keylen = uint8(i) + 1

			//increment nbtc
			nbtc.val += 1

			return
		}
		complen -= 1

	}

	//endpoint exists!
	bt.val++
}

func (bt *BinTrie) getBT(k uint32, m uint8) (bool, *BinTrie) {
	complen := -1
	for i := 31; i >= int(m); i-- {

		//if we have exhausted bt.key (matched all of it)
		if complen <= 0 {
			if k&(1<<uint(i)) == 0 {
				if bt.c0 == nil {
					return false, nil
				}
				bt = bt.c0
				complen = int(bt.keylen)
			} else {
				if bt.c1 == nil {
					return false, nil
				}
				bt = bt.c1
				complen = int(bt.keylen)
			}
		}

		//if we are comparing to bt.key
		//if we dont match
		if (bt.key&(1<<(uint(complen)-1)) == 0 && k&(1<<uint(i)) != 0) || (bt.key&(1<<(uint(complen)-1)) != 0 && k&(1<<uint(i)) == 0) {
			return false, nil
		}
		complen -= 1

	}

	//path exists, are we in the middle of a key?
	if complen == 0 {
		return true, bt
	}
	return false, nil
}

//returns whether k with mask m exists in the tree,
//if it does, its value is returned
//if it does not, the value returned is 0
func (bt *BinTrie) Get(k uint32, m uint8) (bool, uint) {
	b, v := bt.getBT(k, m)
	if b {
		return true, v.val
	}
	return false, 0
}

//returns whether k with mask m exists in the tree,
//if it does, its value plus the value of its parent is returned
//if it does not, the value returned is 0
func (bt *BinTrie) sumParent(k uint32, m uint8) (bool, uint) {
	b, v := bt.getBT(k, m)
	if b {
		return true, v.val + v.p.val
	}
	return false, 0
}

func (bt *BinTrie) sumChildren() uint {
	if bt != nil {
		return bt.val + bt.c0.sumChildren() + bt.c1.sumChildren()
	}
	return 0
}

//deletes a node with key k
func (bt *BinTrie) Delete(k uint32) {

	//in deletion, we first kill the specified child node,
	//then its parent absorbs the other child

	//first find
	b, v := bt.getBT(k, 0)
	//make sure it's an existing child node
	if !b || v.c0 != nil || v.c1 != nil {
		return
	}
	//parent
	p := v.p
	var toAbsorb *BinTrie

	//we can assume p has two children.. unless it's the root!
	//check which one we need to kill & which to absorb
	if p.c0 != v {
		p.val += p.c1.val
		p.c1 = nil
		if p.c0 != nil {
			toAbsorb = p.c0
		}
	} else if p.c1 != v {
		p.val += p.c0.val
		p.c0 = nil
		if p.c1 != nil {
			toAbsorb = p.c1
		}
	}

	if toAbsorb != nil {
		//absorb a child
		p.val += toAbsorb.val
		p.c0 = toAbsorb.c0
		p.c1 = toAbsorb.c1
		if p.c0 != nil {
			p.c0.p = p
		}
		if p.c1 != nil {
			p.c1.p = p
		}
		//combine keys
		p.key = (p.key << toAbsorb.keylen) + toAbsorb.key
		p.keylen += toAbsorb.keylen
	}
}

func (bt *BinTrie) PrintDebug() {
	fmt.Printf("Key: %32.32b, Keylen: %d, Value: %d, 0-child:%t, 1-child:%t\n", bt.key, bt.keylen, bt.val, bt.c0 != nil, bt.c1 != nil)
	if bt.c0 != nil {
		bt.c0.PrintDebug()
	}
	if bt.c1 != nil {
		bt.c1.PrintDebug()
	}
}

func (bt *BinTrie) PrintContents(max int, thresh float32) {
	bt.printWithPrefix(0, 32, max, thresh)
}

func (bt *BinTrie) printWithPrefix(n uint32, m uint8, max int, thresh float32) {
	//fmt.Println(m, bt.keylen, bt.key)
	n += bt.key << (m - bt.keylen)

	if bt.keylen == 0 || bt.val > uint(float32(max)*thresh) {
		for i := 0; i < 32-int(m)+int(bt.keylen); i++ {
			fmt.Printf(" ")
		}
		if m == bt.keylen {
			fmt.Printf("%s %d (%.2f%%)\n", uint32ToIP(n), bt.val, 100.0*float32(bt.val)/float32(max))
		} else {
			fmt.Printf("%s/%d %d (%.2f%%/%.2f%%)\n", uint32ToIP(n), 32-int(m)+int(bt.keylen), bt.val, 100.0*float32(bt.val)/float32(max), 100.0*float32(bt.sumChildren())/float32(max))
		}
	}

	if bt.c0 != nil {
		bt.c0.printWithPrefix(n, m-bt.keylen, max, thresh)
	}
	if bt.c1 != nil {
		bt.c1.printWithPrefix(n, m-bt.keylen, max, thresh)
	}
}

func uint32ToIP(ip uint32) string {
	return strconv.FormatUint(uint64(ip>>24), 10) + "." + strconv.FormatUint(uint64((ip>>16)%(1<<8)), 10) + "." + strconv.FormatUint(uint64((ip>>8)%(1<<8)), 10) + "." + strconv.FormatUint(uint64(ip%(1<<8)), 10)
}
