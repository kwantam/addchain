package main

import (
    "container/heap"
    "fmt"
    "math/big"
    "os"
    "sort"
    "strconv"
    "strings"
    "sync"
)

// ******************* Yacobi L-Z--based *********************** //
type lzdT struct {
    nbits uint
    val *big.Int
    zeroone, one *lzdT
}

func build_lz(x *big.Int) ([]*big.Int, []*lzdT) {
    var maxdepth uint = 1
    var root = &lzdT{1, big.NewInt(1), nil, nil}
    var ord = make([]*lzdT, 0, x.BitLen() / 2)
    var curr *lzdT
    var nzeros uint = 0
    var zero = big.NewInt(0)
    for i := 0; i < x.BitLen(); i++ {
        if curr == nil {
            if x.Bit(i) != 0 {
                // found a 1, which is the root
                curr = root
                if nzeros > 0 {
                    var lzd_zero = &lzdT{nzeros, zero, nil, nil}
                    ord = append(ord, lzd_zero)
                }
            } else {
                nzeros++
            }
        } else if x.Bit(i) == 0 {
            i++
            if i >= x.BitLen() || x.Bit(i) == 0 {
                // two zeros in a row can't be a codeword
                ord = append(ord, curr)
                nzeros, curr = 2, nil
                continue
            }

            if curr.zeroone == nil {
                maxdepth, ord = insert_lz(curr, ord, maxdepth, false)
                nzeros, curr = 0, nil
            } else {
                curr = curr.zeroone
            }
        } else {
            if curr.one == nil {
                maxdepth, ord = insert_lz(curr, ord, maxdepth, true)
                nzeros, curr = 0, nil
            } else {
                curr = curr.one
            }
        }
    }
    if curr != nil {
        ord = append(ord, curr)
    }
    //display_lz(root)

    var chn = make([]*big.Int, 0, len(ord) + int(maxdepth))
    for i := uint(0); i < maxdepth; i++ {
        var tmp = big.NewInt(1)
        tmp.Lsh(tmp, i)
        chn = append(chn, tmp)
    }
    for i := 0; i < len(ord); i++ {
        chn = insert(chn, ord[i].val)
    }

    return chn, ord
}

func insert_lz(curr *lzdT, ord []*lzdT, maxdepth uint, is_one bool) (uint, []*lzdT) {
    var depth = curr.nbits + 2
    if is_one {
        depth = curr.nbits + 1
    }
    var tmp = big.NewInt(1)
    tmp.Lsh(tmp, depth - 1).Add(tmp, curr.val)
    var new_lz = &lzdT{depth, tmp, nil, nil}

    // insert into tree
    if is_one {
        curr.one = new_lz
    } else {
        curr.zeroone = new_lz
    }

    // new maxdepth
    if depth > maxdepth {
        maxdepth = depth
    }

    return maxdepth, append(ord, new_lz)
}

func display_lz(root *lzdT) {
    if root == nil {
        return
    }

    fmt.Println(root.nbits, root.val, show_binary(root.val))
    display_lz(root.zeroone)
    display_lz(root.one)
}

func show_binary(x *big.Int) (string) {
    var buffer = ""
    for bitnum := x.BitLen() - 1; bitnum >= 0; bitnum-- {
        if x.Bit(bitnum) > 0 {
            buffer += "1"
        } else {
            buffer += "0"
        }
    }

    return buffer
}

func yacobi_lz(x *big.Int) ([]*big.Int) {
    var chn, ord = build_lz(x)

    var last = len(ord) - 1
    var curr = big.NewInt(0).Set(ord[last].val)
    for j := last - 1; j >= 0; j-- {
        var cord = ord[j]
        for i := uint(0); i < cord.nbits; i++ {
            curr.Lsh(curr, 1)
            chn = insert(chn, big.NewInt(0).Set(curr))
        }
        curr.Add(curr, cord.val)
        chn = insert(chn, big.NewInt(0).Set(curr))
    }

    return chn
}

// ******************* Bergeron-Berstel-Brlek-Duboc ******************** //
func oplus(v []*big.Int, j *big.Int) ([]*big.Int) {
    var tmp = big.NewInt(0)
    tmp.Add(v[len(v)-1], j)
    return append(v, tmp)
}

func otimes(v, w []*big.Int) ([]*big.Int) {
    vlast := v[len(v)-1]
    for _, wi := range w {
        wi.Mul(wi, vlast)
        if wi.Cmp(vlast) != 0 {
            v = append(v, wi)
        }
    }
    return v
}

func ispow2(x *big.Int) ([]*big.Int) {
    var log = x.BitLen()
    var tmp = big.NewInt(int64(log-1))
    var two = big.NewInt(2)
    tmp.Exp(two, tmp, nil)
    if tmp.Cmp(x) != 0 {
        return nil
    }
    var ret = make([]*big.Int, 0, log)
    for i := 0; i < log; i++ {
        ret = append(ret, big.NewInt(0).Exp(two, big.NewInt(int64(i)), nil))
    }
    return ret
}

func minchain(x *big.Int) ([]*big.Int) {
    if pret := ispow2(x); pret != nil {
        return pret
    }

    var two = big.NewInt(2)
    var three = big.NewInt(3)
    if x.Cmp(three) == 0 {
        return []*big.Int{big.NewInt(1), two, three}
    }

    // don't need to subtract 1 from n because if n were a power of two we would have returned by now
    var hf_lgn int = x.BitLen() / 2
    var k = big.NewInt(0).Div(x, big.NewInt(0).Exp(two, big.NewInt(int64(hf_lgn)), nil))
    return chain(x, k)
}

func chain(x, k *big.Int) ([]*big.Int) {
    var q, r = big.NewInt(0), big.NewInt(0)
    q.DivMod(x, k, r)
    if r.Sign() == 0 {
        return otimes(minchain(k), minchain(q))
    }
    return oplus(otimes(chain(k, r), minchain(q)), r)
}


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
            fmt.Println("*** ERROR *** could not find predecessor for value", val, "*** ERROR ***")
            os.Exit(-1)
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

func print_sequence(sequence []seqT) {
    const eqlen = 18
    var i = 0
    for _, seq := range sequence {
        var pstr string
        if i == 0 {
            pstr = fmt.Sprintf("t%d = input", seq.varnum)
        } else if seq.varnum < 0 {
            continue
        } else if seq.l == seq.r {
            pstr = fmt.Sprintf("t%d = sqr(t%d)", seq.varnum, sequence[seq.l].varnum)
        } else {
            pstr = fmt.Sprintf("t%d = t%d * t%d", seq.varnum, sequence[seq.l].varnum, sequence[seq.r].varnum)
        }

        pstr += strings.Repeat(" ", eqlen - len(pstr))
        fmt.Println(pstr, "#", i, ":", seq.val)
        i++
    }
}

// **************************** Bos-Coster ***************************** //
// **** windowing **** //
type winT struct {
    start, end, wval int
}

// re-balance windows
func win_balance(x *big.Int, runs []winT) ([]winT) {
    for i := 0; i < len(runs) - 1; i++ {
        // consider successive runs that are both non-zero
        if runs[i].wval == 0 || runs[i+1].wval == 0 {
            continue
        }
        var lr1 = runs[i].start - runs[i].end + 1
        var lr2 = runs[i+1].start - runs[i+1].end + 1

        // only rebalance if there's a difference
        if (lr1 - lr2 < 2) && (lr2 - lr1 < 2) {
            continue
        }

        // rebalance windows
        var mid = (runs[i].start + runs[i+1].end) / 2
        runs[i].end = mid + 1
        runs[i+1].start = mid

        // fix up window values
        fix_run(x, &(runs[i]))
        fix_run(x, &(runs[i+1]))

        // if we created a hole, fix it
        if runs[i].end - runs[i+1].start > 1 {
            runs = append(runs, runs[len(runs)-1])
            for j := len(runs)-2; j > i+1; j-- {
                runs[j] = runs[j-1]
            }
            runs[i+1] = winT{runs[i].end - 1, runs[i+2].start + 1, 0}
            i++
        }
        i++
    }

    return runs
}

func window(x *big.Int, winsize int) ([]winT) {
    var in_window = false
    var win_start = -1
    var last_one = -1
    var thiswinsize = 0
    var bit = -1
    var bitnum = x.BitLen() - 1
    var runs = make([]winT, 0, 1 + bitnum / winsize)
    var thisrun = make([]int, 0, 1 + winsize)
    for ; bitnum >= -1; bitnum-- {
        // dummy value for last iteration
        if bitnum == -1 {
            bit = -1
        } else {
            bit = int(x.Bit(bitnum))
        }
        if in_window {
            // update last-seen 1
            if bit > 0 {
                last_one = bitnum
                thisrun = append(thisrun, 1)
            } else {
                thisrun = append(thisrun, 0)
            }

            if (thiswinsize < winsize) && (bit >= 0) {
                // continue growing the window
                thiswinsize++
            } else {
                // compute the value in this window
                var one_delta = last_one - bitnum
                var thisrunval = 0
                for _, bval := range thisrun[:len(thisrun)-one_delta] {
                    thisrunval <<= 1
                    thisrunval += bval
                }
                // end the window
                runs = append(runs, winT{win_start, last_one, thisrunval})
                thisrun = thisrun[:0]
                in_window = false
                win_start = last_one - 1
                thiswinsize = last_one - bitnum + 1
                last_one = -1
                // special case when we hit the end of the number
                if bitnum == -1 && win_start >= 0 {
                    runs = append(runs, winT{win_start, 0, 0})
                }
            }
        } else {
            if bit != 0 {
                // valid zero-run has non-zero length and a valid start position
                if win_start >= 0  && win_start > bitnum {
                    runs = append(runs, winT{win_start, bitnum + 1, 0})
                }
                thisrun = append(thisrun, 1)
                in_window = true
                win_start = bitnum
                thiswinsize = 1
                last_one = bitnum
            }
        }
    }
    return win_balance(x, runs)
}

// fix-up a run (for rebalancing)
func fix_run(x *big.Int, run *winT) {
    // move start and end until they're on ones
    for ; x.Bit(run.start) == 0; {
        run.start--
    }
    for ; x.Bit(run.end) == 0; {
        run.end++
    }

    var ival = 0
    for i := run.start; i >= run.end; i-- {
        ival <<= 1
        ival += int(x.Bit(i))
    }
    run.wval = ival
}

// from https://github.com/cznic/sortutil/
type BigIntSlice []*big.Int
func (s BigIntSlice) Len() int           { return len(s) }
func (s BigIntSlice) Less(i, j int) bool { return s[i].Cmp(s[j]) < 0 }
func (s BigIntSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// uniqued, sorted list of window values
func runs_to_dict(runs []winT) ([]*big.Int) {
    var dict = make(map[int]bool)
    dict[1] = true
    dict[2] = true
    for _, run := range runs {
        if run.wval > 0 {
            dict[run.wval] = true
        }
    }

    var ret = make(BigIntSlice, 0, len(dict))
    for k, _ := range dict {
        ret = append(ret, big.NewInt(int64(k)))
    }
    sort.Sort(ret)
    return ret
}

// print out a graphical representation of the window
func display_window(x *big.Int, runs []winT) {
    fmt.Println(show_binary(x))

    var linestr = ""
    var valstr = ""
    for _, run := range runs {
        if run.wval > 0 {
            if run.start == run.end {
                linestr += "|"
                valstr += "1"
            } else {
                linestr += ">" + strings.Repeat("-", run.start - run.end - 1) + "<"
                var tvstr = strconv.Itoa(run.wval)
                valstr += tvstr + strings.Repeat(" ", run.start - run.end + 1 - len(tvstr))
            }
        } else {
            var pstr = strings.Repeat(" ", run.start - run.end + 1)
            linestr += pstr
            valstr += pstr
        }
    }
    println(linestr)
    println(valstr)
}

// **** dictionary / chain management **** //
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

// **** Bos-Coster reduction methods **** //
func bc_approx_test(d []*big.Int) (int) {
    var targ = d[len(d)-1]
    var tmp = big.NewInt(0)
    var eps = big.NewInt(int64(targ.BitLen() - 1))

    var aIdx = -1
    ApxOuter: for i := len(d)-2; i >= 0; i-- {
        for j := i-1; j >= 0; j-- {
            if tmp.Add(d[i], d[j]).Sub(targ, tmp).Cmp(eps) < 0 {
                // found small epsilon
                aIdx = i
                break ApxOuter
            }
        }
    }

    return aIdx
}

func bc_approx(d, chn []*big.Int, aIdx int) ([]*big.Int, []*big.Int) {
    var targ = d[len(d)-1]
    var tmp = big.NewInt(0).Sub(targ, d[aIdx])
    d = insert(d[:len(d)-1], tmp)
    chn = insert(chn, targ)
    return d, chn
}

func bc_halve_test(d []*big.Int) (int, int) {
    var targ = d[len(d)-1]
    var tmp = big.NewInt(0)
    var blen, best = 0, -1

    for i := 0; i < len(d) - 1; i++ {
        tmp.Sub(targ, d[i])
        var j = 0
        for ; tmp.Bit(j) == 0; j++ {}           // how many times does 2 divide?
        if j > blen {
            blen = j
            best = i
        }
    }

    return blen, best
}

func bc_halve(d, chn []*big.Int, blen, best int) ([]*big.Int, []*big.Int) {
    var targ = d[len(d)-1]
    var tmp = big.NewInt(0)

    tmp.Sub(targ, d[best])
    var klst = make([]*big.Int, 0, blen)
    for j := blen - 1; j >= 0; j-- {
        klst = append(klst, big.NewInt(0).Rsh(tmp, uint(j)))
    }
    d = merge(d[:len(d)-1], klst)
    chn = insert(chn, targ)

    return d, chn
}

func bc_divide_test(targ *big.Int) (*big.Int) {
    var smallvals = [...]int{19, 17, 13, 11, 7 , 5, 3}
    var tmp = big.NewInt(0)

    for _, val := range smallvals {
        if tmp.SetInt64(int64(val)).Mod(targ, tmp).Sign() == 0 {
            return tmp.SetInt64(int64(val))
        }
    }

    return nil
}

func bc_divide(d, chn, dchain []*big.Int, div *big.Int) ([]*big.Int, []*big.Int) {
    var targ = d[len(d)-1]
    div.Div(targ, div)
    for _, dch := range dchain {
        dch.Mul(dch, div)
    }
    d = merge(d[:len(d)-1], dchain)
    chn = insert(chn, targ)

    return d, chn
}

// **** Bos-Coster dispatch **** //
func bos_coster(q *big.Int, winsize int) ([]*big.Int) {
    var runs = window(q, winsize)
    var d = runs_to_dict(runs)
    var chn = make([]*big.Int, 0, len(d))

    for ; len(d) > 2; {
        if aIdx := bc_approx_test(d); aIdx >= 0 {
            d, chn = bc_approx(d, chn, aIdx)
            continue
        }

        // choose between division and halving
        var blen, best = bc_halve_test(d)
        if div := bc_divide_test(d[len(d)-1]); div != nil {
            var dchain = minchain(div)
            dchain = dchain[:len(dchain)-1]

            if best < 0 || len(dchain) < blen {
                d, chn = bc_divide(d, chn, dchain, div)
                continue
            }
        }

        if best >= 0 {
            d, chn = bc_halve(d, chn, blen, best)
        } else {
            panic("Cannot make progress.")
        }
    }
    chn = merge(d, chn)

    var curr = big.NewInt(int64(runs[0].wval))
    for _, run := range runs[1:] {
        var ln = run.start - run.end + 1
        for i := 0; i < ln; i++ {
            curr.Lsh(curr, 1)
            chn = insert(chn, big.NewInt(0).Set(curr))
        }
        curr.Add(curr, big.NewInt(int64(run.wval)))
        chn = insert(chn, big.NewInt(0).Set(curr))
    }

    return chn
}

// ********************** cmdline UI functions ********************* //
func convert_next_value(current string) (string, string, bool, bool) {
    var i int
    for i=0; len(current) > i && current[i] >= '0' && current[i] <= '9'; i++ {}
    nextval := current[0:i]
    var newcurr string
    var do_add bool
    if i == len(current) {
        newcurr = ""
        do_add = true
    } else {
        if (current[i] != '+') && (current[i] != '-') {
            return "", "", true, true
        } else {
            newcurr = current[i+1:]
            do_add = true
            if current[i] == '-' {
                do_add = false
            }
        }
    }
    return nextval, newcurr, do_add, false
}

func convert_formula(formula string) (*big.Int) {
    var current, next string = formula, ""
    var do_add, do_add_next, err bool = true, true, false
    var two, tmp, q = big.NewInt(2), big.NewInt(0), big.NewInt(0)

    for ; len(current) > 0; {
        tmp.SetUint64(0)
        if (len(current) > 1) && (current[0:2] == "2^") {
            current = current[2:]
            if len(current) == 0 {
                return nil
            }
            if current, next, do_add_next, err = convert_next_value(current); err {
                return nil
            }
            tmp.SetString(current, 10)
            tmp.Exp(two, tmp, nil)
        } else {
            if current, next, do_add_next, err = convert_next_value(current); err {
                return nil
            }
            tmp.SetString(current, 10)
        }

        if do_add {
            q.Add(q, tmp)
        } else {
            q.Sub(q, tmp)
        }

        current = next
        do_add = do_add_next
    }

    return q
}

func usage() {
    fmt.Printf("Usage: %s <formula>\n\n<formula> can be a decimal number or a formula like 2^255-19\n", os.Args[0])
}

type runT struct {
    size int
    seq []seqT
}

func show_run(win *runT, slen, ssto int) {
    if win.size == 0 {
        fmt.Print("# Yacobi")
    } else if win.size == 1 {
        fmt.Print("# Bergeron-Berstel-Brlek-Duboc")
    } else {
        fmt.Printf("# Bos-Coster (win=%d)", win.size)
    }
    fmt.Printf(": %d (%d)\n", slen, ssto)
}

func main() {
    // read in arguments
    if len(os.Args) < 2 {
        usage()
        fmt.Println("\nYou must specify q > 4.")
        os.Exit(-1)
    }
    var q = convert_formula(os.Args[1])
    if q == nil {
        usage()
        fmt.Printf("\ncannot convert formula '%s'\n", os.Args[1])
        os.Exit(-1)
    }

    // search in parallel
    const swin = 2
    const ewin = 11
    var ch = make(chan runT, ewin - swin + 2)
    var wg = sync.WaitGroup{}

    // Yacobi LZ method
    wg.Add(1)
    go func (wg *sync.WaitGroup) {
        ch <- runT{0, make_sequence(yacobi_lz(q))}
        wg.Done()
    }(&wg)

    // Bergeron-Berstel-Brlek-Duboc
    wg.Add(1)
    go func (wg *sync.WaitGroup) {
        ch <- runT{1, make_sequence(minchain(q))}
        wg.Done()
    }(&wg)

    // Bos-Coster for various window sizes
    for i := swin; i < ewin; i++ {
        wg.Add(1)
        go func (wg *sync.WaitGroup, i int) {
            ch <- runT{i, make_sequence(bos_coster(q, i))}
            wg.Done()
        }(&wg, i)
    }

    // wait for all to be done
    wg.Wait()
    close(ch)

    // find the best result
    var win runT
    var wlen, wsto = (1 << 30), (1 << 30)
    for wx := range ch {
        var xlen, xsto = seq_len(wx.seq), seq_storage(wx.seq)
        show_run(&wx, xlen, xsto)

        if xlen < wlen || (xlen == wlen && xsto < wsto) {
            wlen, wsto = xlen, xsto
            win = wx
        }
    }

    // show the winner
    print_sequence(win.seq)
    show_run(&win, wlen, wsto)
}
