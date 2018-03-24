package main

import (
    "container/heap"
    "fmt"
    "math/big"
    "sort"
)

// ******************** create and display sequences ********************** //
type seqT struct {
    l, r, varnum int
    val *big.Int
}

// int-heap (from container/heap examples)
type IntHeap []int
func (h IntHeap) Len() int              { return len(h) }
func (h IntHeap) Less(i, j int) bool    { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)         { h[i], h[j] = h[j], h[i] }
func (h *IntHeap) Push(x interface{})   { *h = append(*h, x.(int)) }
func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func seq_alloc(sequence []seqT) ([]seqT) {
    // heap of available variables
    var avails = make(IntHeap, 0, len(sequence))
    for i := 1; i < len(sequence); i++ {
        avails = append(avails, i)
    }
    heap.Init(&avails)

    // always need output value
    sequence[len(sequence)-1].varnum = 0

    var maxstor = 0
    var trim = 0
    for i := len(sequence) - 1; i > 0; i-- {
        if sequence[i].varnum < 0 {
            // reached an unneeded variable! trim it
            trim += 1
            continue
        }

        // no longer need this varnum allocated
        heap.Push(&avails, sequence[i].varnum)
        var lval = sequence[i].l
        var rval = sequence[i].r

        // allocate variables for predecessors, if necesary
        for _, lr := range [2]int{lval, rval} {
            if sequence[lr].varnum < 0 {
                var nextvar = heap.Pop(&avails).(int)
                sequence[lr].varnum = nextvar
                if nextvar > maxstor {
                    maxstor = nextvar
                }
            }
        }
    }

    sequence[0].r = trim
    sequence[0].l = maxstor
    return sequence
}

func make_sequence(chn []*big.Int) ([]seqT) {
    var tmp = big.NewInt(0)
    var sequence = make([]seqT, 0, len(chn))
    for i, val := range chn {
        if i == 0 {
            sequence = append(sequence, seqT{-1, 0, -1, val})
            continue
        }

        var found = false
        SeqOuter: for j := 0; j < i; j++ {
            for k := 0; k <= j; k++ {
                if val.Cmp(tmp.Add(chn[j], chn[k])) == 0 {
                    sequence = append(sequence, seqT{j, k, -1, val})
                    found = true
                    break SeqOuter
                }
            }
        }
        if !found {
            panic(fmt.Sprintf("Could not find predecessor for value %v", val))
        }
    }

    return seq_alloc(sequence)
}

func seq_len(sequence []seqT) (int) {
    if len(sequence) == 0 {
        return 0
    } else {
        return len(sequence) - sequence[0].r
    }
}

func seq_storage(sequence []seqT) (int) {
    if len(sequence) == 0 {
        return 0
    } else {
        return sequence[0].l + 1
    }
}

// ********************** dictionary / chain management *********************** //
func insert(l []*big.Int, v *big.Int) ([]*big.Int) {
    if v.Sign() == 0 {
        return l
    }

    // maybe 0-length array, so just create a new one
    var ln = len(l)
    if ln == 0 {
        return append(l, v)
    }

    var idx = sort.Search(ln, func(i int) bool { return l[i].Cmp(v) >= 0 })
    // maybe value just goes at end of array
    if idx == ln {
        return append(l, v)
    }

    // make sure we are not inserting a duplicate value
    if v.Cmp(l[idx]) == 0 {
        return l
    }

    // otherwise need to move values out of the way first
    l = append(l, nil)
    for i := ln; i > idx; i-- {
        l[i] = l[i-1]
    }
    l[idx] = v
    return l
}

func merge(l1, l2 []*big.Int) ([]*big.Int) {
    var ret = make([]*big.Int, 0, len(l1) + len(l2))
    var i1, i2, ll1, ll2 = 0, 0, len(l1), len(l2)

    // while we still have elements
    for ; i1 < ll1 || i2 < ll2; {
        if i1 == ll1 {
            // i1 is out
            ret = append(ret, l2[i2])
            i2++
        } else if i2 == ll2 {
            // i2 is out
            ret = append(ret, l1[i1])
            i1++
        } else {
            var cres = l1[i1].Cmp(l2[i2])
            if cres < 0 {
                // i1 is lesser value
                ret = append(ret, l1[i1])
                i1++
            } else if cres == 0 {
                // i1 and i2 have the same value
                ret = append(ret, l1[i1])
                i1++
                i2++
            } else {
                // i2 is lesser value
                ret = append(ret, l2[i2])
                i2++
            }
        }
    }

    return ret
}
