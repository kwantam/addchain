package main

import "math/big"

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



