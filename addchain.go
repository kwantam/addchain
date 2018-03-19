package main

import (
    "fmt"
    "math/big"
    "os"
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

/*
func bits(x *big.Int) {
    var bitnum = x.BitLen() - 1
    for ; bitnum >= 0; bitnum-- {
        if x.Bit(bitnum) > 0 {
            print("1 ")
        } else {
            print("0 ")
        }
    }
    print("\n")
}

func window(x *big.Int, len int) {
    var in_window = false
    var bitnum = x.BitLen()
}
*/

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

    //bits(q)
    sequence(minchain(q))
}
