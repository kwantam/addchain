package main

import "math/big"

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
