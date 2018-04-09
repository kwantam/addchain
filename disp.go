package main

import (
    "bytes"
    "fmt"
    "math/big"
    "strings"
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
    var buffer bytes.Buffer
    for bitnum := x.BitLen() - 1; bitnum >= 0; bitnum-- {
        if x.Bit(bitnum) > 0 {
            buffer.WriteRune('1')
        } else {
            buffer.WriteRune('0')
        }
    }

    return buffer.String()
}

// print out a graphical representation of an add-sequence
func print_sequence(sequence []seqT) {
    const eqlen = 18
    var i = 0
    var buffer bytes.Buffer

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

        fmt.Fprintf(&buffer, "%*s # %4d : %v\n", -eqlen, pstr, i, seq.val)
        i++
    }

    fmt.Print(buffer.String())
}

// print out a graphical representation of a windowed binary representation
func display_window(x *big.Int, runs []winT) {
    fmt.Println(show_binary(x))

    var linestr bytes.Buffer
    var valstr bytes.Buffer
    for _, run := range runs {
        if run.wval > 0 {
            if run.start == run.end {
                linestr.WriteRune('|')
                valstr.WriteRune('1')
            } else {
                linestr.WriteRune('>')
                linestr.WriteString(strings.Repeat("-", run.start - run.end - 1))
                linestr.WriteRune('<')
                fmt.Fprintf(&valstr, "%*d", -(run.start - run.end + 1), run.wval)
            }
        } else {
            var pstr = strings.Repeat(" ", run.start - run.end + 1)
            linestr.WriteString(pstr)
            valstr.WriteString(pstr)
        }
    }

    fmt.Println(linestr.String())
    fmt.Println(valstr.String())
}

// show the result of a run
func show_run(win *runT, slen, ssto int) {
    var pstr string
    if win.size == 0 {
        pstr = "# Yacobi"
    } else if win.size == 1 {
        pstr = "# Bergeron-Berstel-Brlek-Duboc"
    } else {
        pstr = fmt.Sprintf("# Bos-Coster (win=%d)", win.size)
    }

    fmt.Printf("%-30s : %4d (%2d)\n", pstr, slen, ssto)
}
