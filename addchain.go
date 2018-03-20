package main

import (
    "fmt"
    "math/big"
    "os"
    "sort"
    "strconv"
    "strings"
)

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

func sequence(chn []*big.Int) {
    var tmp = big.NewInt(0)
    for i, val := range chn {
        fmt.Print(val)
        if i == 0 {
            fmt.Println("    a[0]=x")
            continue
        }

        var found = false
        SeqOuter: for quo := 0; quo < i; quo++ {
            for rem := 0; rem <= quo; rem++ {
                if val.Cmp(tmp.Add(chn[quo], chn[rem])) == 0 {
                    if rem < quo {
                        fmt.Printf("    a[%d]=a[%d]*a[%d]\n", i, rem, quo)
                    } else {
                        fmt.Printf("    a[%d]=square(a[%d])\n", i, rem)
                    }
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
}

type winT struct {
    start, end, wval int
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
    return runs
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
    for bitnum := x.BitLen() - 1; bitnum >= 0; bitnum-- {
        if x.Bit(bitnum) > 0 {
            print("1")
        } else {
            print("0")
        }
    }
    print("\n")
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

func bc_approx(d, chn []*big.Int) ([]*big.Int, []*big.Int) {
    var targ = d[len(d)-1]
    var tmp = big.NewInt(0)
    // TODO should this be a constant rather than log(k) ???
    var eps = big.NewInt(int64(targ.BitLen() - 1))
    //var eps = big.NewInt(3)

    ApxOuter: for i := len(d)-2; i >= 0; i-- {
        for j := i-1; j >= 0; j-- {
            if tmp.Add(d[i], d[j]).Sub(targ, tmp).Cmp(eps) < 0 {
                // found small epsilon
                d = insert(d[:len(d)-1], tmp.Add(d[j], tmp))
                chn = insert(chn, targ)
                break ApxOuter
            }
        }
    }

    return d, chn
}

func bc_halve(d, chn []*big.Int) ([]*big.Int, []*big.Int) {
    var targ = d[len(d)-1]
    var tmp = big.NewInt(0)
    var blen int

    for i := len(d) - 2; i >= 0; i-- {
        tmp.Div(targ, d[i])                             // tmp = targ / fi
        blen = tmp.BitLen() - 1                         // u = log2(targ / fi)
        if blen < 1 {                                   // TODO is 1 the right value here ???
            continue
        }
        tmp.Rsh(targ, uint(blen))                       // targ // 2^blen
        var klst = make([]*big.Int, 0, 2 + blen)        // k 2k 4k ... 2^blen k
        for j := 0; j <= blen; j++ {
            klst = append(klst, big.NewInt(0).Lsh(tmp, uint(j)))
        }
        tmp.Sub(targ, klst[blen])                   // targ - 2^blen k
        klst = insert(klst, tmp)                    // add it to the list

        d = merge(d[:len(d)-1], klst)
        chn = insert(chn, targ)
        break
    }

    return d, chn
}

func bc_cleanup(d, chn []*big.Int) ([]*big.Int, []*big.Int) {
    if len(chn) == 0 {
        return d, chn
    }
    var targ = d[len(d)-1]
    var idx = sort.Search(len(chn), func(i int) bool { return chn[i].Cmp(targ) >= 0 })
    if idx < len(chn) && targ.Cmp(chn[idx]) == 0 {
        d = d[:len(d)-1]
    }
    if d[0].Sign() == 0 {
        d = d[1:]
    }
    if chn[0].Sign() == 0 {
        chn = chn[1:]
    }
    return d, chn
}

func bos_coster(q *big.Int, winsize int) ([]*big.Int) {
    var runs = window(q, winsize)
    //display_window(q, runs)
    var d = runs_to_dict(runs)
    var chn = make([]*big.Int, 0, len(d))

    for ; len(d) > 2; {
        d, chn = bc_halve(d, chn)
        d, chn = bc_approx(d, chn)
        d, chn = bc_cleanup(d, chn)
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

func usage() {
    fmt.Printf("Usage: %s <formula> [-w <winsize>]\n", os.Args[0])
}

func main() {
    var q *big.Int = nil

    // read in arguments
    if len(os.Args) < 2 {
        usage()
        fmt.Println("\nYou must specify q > 4.")
        os.Exit(-1)
    }
    q = convert_formula(os.Args[1])
    if q == nil {
        usage()
        fmt.Printf("\ncannot convert formula '%s'\n", os.Args[1])
        os.Exit(-1)
    }

    // try Bergeron-Berstel-Brlek-Duboc first
    var winner = 1
    var chn = minchain(q)
    // then try Bos-Coster for various window sizes
    for i := 2; i < 10; i++ {
        var tmp = bos_coster(q, i)
        if len(tmp) < len(chn) {
            winner = i
            chn = tmp
        }
    }
    sequence(chn)
    fmt.Printf("(Winner was method %d.)\n", winner)
}
