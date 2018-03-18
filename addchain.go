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
    var two, q, tmp *big.Int = big.NewInt(2), big.NewInt(0), big.NewInt(0)

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
    fmt.Printf("%s (prime=%t)\n", q, q.ProbablyPrime(128))
}
