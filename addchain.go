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
    var two, q, tmp = big.NewInt(2), big.NewInt(0), big.NewInt(0)

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

        // TODO make this two nested loops
        for j := 0; j < i * i; j++ {
            var rem = j % i
            var quo = j / i
            if rem <= quo {
                var cmp = val.Cmp(tmp.Add(chn[quo], chn[rem]))
                if cmp == 0 {
                    if rem < quo {
                        fmt.Printf("    a[%d]=a[%d]*a[%d]\n", i, rem, quo)
                    } else {
                        fmt.Printf("    a[%d]=square(a[%d])\n", i, rem)
                    }
                    break
                }
            }
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

func runs_to_dict(runs []winT) ([]int) {
    var dict = make(map[int]bool)
    dict[1] = true
    dict[2] = true
    for _, run := range runs {
        if run.wval > 0 {
            dict[run.wval] = true
        }
    }

    var ret = make([]int, 0, len(dict))
    for k, _ := range dict {
        ret = append(ret, k)
    }
    sort.Ints(ret)
    return ret
}

func display_window(x *big.Int, runs []winT) {
    // print out a graphical representation of the window
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

func usage() {
    fmt.Printf("Usage: %s <formula>\n", os.Args[0])
}

func main() {
    if len(os.Args) < 2 {
        usage()
        os.Exit(-1)
    }

    q := convert_formula(os.Args[1])
    if q == nil {
        usage()
        fmt.Printf("\ncannot convert formula '%s'\n", os.Args[1])
        os.Exit(-1)
    }

    var runs = window(q, 10)
    display_window(q, runs)
    fmt.Println(runs_to_dict(runs))
    //sequence(minchain(q))
}
