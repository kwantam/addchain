package main

import (
    "fmt"
    "math/big"
    "os"
    "sync"
)

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
