package main

import (
    "fmt"
    "math/big"
    "strings"
    "strconv"
)

// print the LZ compression tree from Yacobi
func display_lz(root *lzdT) {
    if root == nil {
        return
    }

    fmt.Println(root.nbits, root.val, show_binary(root.val))
    display_lz(root.zeroone)
    display_lz(root.one)
}

// show a binary expansion of a big.Int
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

// print out a graphical representation of an add-sequence
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

// print out a graphical representation of a windowed binary representation
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
