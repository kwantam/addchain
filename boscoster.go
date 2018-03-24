package main

import (
    "math/big"
    "sort"
)

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

// ********************* Bos-Coster reduction methods ******************* //
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
