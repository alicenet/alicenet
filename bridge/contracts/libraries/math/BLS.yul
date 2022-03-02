object "Deployer" {
    code {
        datacopy(returndatasize(), dataoffset("r1"), datasize("r1"))
        log0(returndatasize(), datasize("r1"))
        return(returndatasize(), datasize("r1"))
    }

    object "r1" {
        code {


            /*
            function baseToG1(ptr, t)->x,y {
                let fp := add(ptr,0xc0)
                let ap1 := mulmod(t, t, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                let alpha := mulmod(ap1, addmod(ap1, 4, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                mstore(add(ptr, 0x60), alpha)
                mstore(add(ptr, 0x80), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd45)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                alpha := mload(fp)
                ap1 := mulmod(ap1, ap1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(ap1, 0xb3c4d79d41a91759a9e4c7e359b6b89eaec68e62effffffd, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(x, alpha, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                x := addmod(x, 0x59e26bcea0d48bacd4f263f1acdb5c4f5763473177fffffe, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                let x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                mstore(add(ptr, 0x60), x_three)
                mstore(add(ptr, 0x80), 0x183227397098d014dc2822db40c0ac2ecbc0b548b438e5469e10460b6c3e7ea3)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                if iszero(eq( mload(fp), 1)) {
                    x  := addmod(x, 1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                    x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    mstore(add(ptr, 0x60), x_three)
                    if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                        revert(0, 0)
                    }
                    if iszero(eq( mload(fp), 1)) {
                        let ap2 := addmod(mulmod(t, t, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 4, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x := mulmod(ap2, ap2, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x := mulmod(x, ap2, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x := mulmod(x, 0x2042def740cbc01bd03583cf0100e593ba56470b9af68708d2c05d6490535385, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x := mulmod(x, alpha, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                        x := addmod(x, 1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    }
                }
                mstore(add(ptr, 0x60), x_three)
                mstore(add(ptr, 0x80), 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                y := mload(fp)
                if gt(t,0x183227397098d014dc2822db40c0ac2ecbc0b548b438e5469e10460b6c3e7ea3) {
                   y := mulmod(y, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd46, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                }
            }
            */

            function bn128_is_on_curve(p0,p1) ->result {
                    let o1 := mulmod(p0, p0, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    o1 := mulmod(p0, o1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    o1 := addmod(o1,3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    let o2 := mulmod(p1,p1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    result := eq(o1,o2)
                }

            function baseToG1(ptr, t)->x,y {
                let fp := add(ptr,0xc0)
                let ap1 := mulmod(t, t, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                let alpha := mulmod(ap1, addmod(ap1, 4, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                mstore(add(ptr, 0x60), alpha)
                mstore(add(ptr, 0x80), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd45)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                alpha := mload(fp)
                ap1 := mulmod(ap1, ap1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(ap1, 0xb3c4d79d41a91759a9e4c7e359b6b89eaec68e62effffffd, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(x, alpha, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                x := addmod(x, 0x59e26bcea0d48bacd4f263f1acdb5c4f5763473177fffffe, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                let x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                mstore(add(ptr, 0x80), 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52)
                mstore(add(ptr, 0x60), x_three)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                let ymul := 1
                if gt(t,0x183227397098d014dc2822db40c0ac2ecbc0b548b438e5469e10460b6c3e7ea3) {
                   ymul := 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd46
                }
                y := mulmod(mload(fp), ymul, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                let y_two := mulmod(y, y, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                if eq(x_three, y_two) {
                    leave
                }
                x  := addmod(x, 1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                mstore(add(ptr, 0x60), x_three)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                y := mulmod(mload(fp), ymul, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                y_two := mulmod(y, y, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                if eq(x_three, y_two) {
                    leave
                }
                ap1 := addmod(mulmod(t, t, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 4, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(ap1, ap1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(x, ap1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(x, 0x2042def740cbc01bd03583cf0100e593ba56470b9af68708d2c05d6490535385, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := mulmod(x, alpha, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                x := addmod(x, 1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                mstore(add(ptr, 0x60), x_three)
                if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                    revert(0, 0)
                }
                y := mulmod(mload(fp), ymul, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
            }

            function HashToG1(ptr, messageptr, messagesize) ->x,y {
                let size := add(messagesize, 1)
                calldatacopy(add(ptr,1), messageptr, messagesize)
                mstore8(ptr, 0x00)
                let h0 := keccak256(ptr, size)
                mstore8(ptr, 0x01)
                let h1 := keccak256(ptr, size)
                mstore8(ptr, 0x02)
                let h2 := keccak256(ptr, size)
                mstore8(ptr, 0x03)
                let h3 := keccak256(ptr, size)
                mstore(ptr, 0x20)
                mstore(add(ptr, 0x20), 0x20)
                mstore(add(ptr, 0x40), 0x20)
                mstore(add(ptr, 0xa0), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                h1 := addmod(h1, mulmod(h0, 0xe0a77c19a07df2f666ea36f7879462c0a78eb28f5c70b3dd35d438dc58f0d9d, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                h2 := addmod(h3, mulmod(h2, 0xe0a77c19a07df2f666ea36f7879462c0a78eb28f5c70b3dd35d438dc58f0d9d, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                let x1,y1 := baseToG1(ptr, h1)
                let x2,y2 := baseToG1(ptr, h2)
                mstore(ptr, x1)
                mstore(add(ptr, 0x20), y1)
                mstore(add(ptr, 0x40), x2)
                mstore(add(ptr, 0x60), y2)
                if iszero(staticcall(gas(), 0x06, ptr, 128, ptr, 64)) {
                    revert(0,0)
                }
                x:=mload(ptr)
                y:=mload(add(ptr, 0x20))
            }


            let x,y := HashToG1(0,0, calldatasize())
            mstore(0,x)
            mstore(0x20,y)
            log0(0,0x20)
            log0(0x20,0x20)
            return(0,0x40)
        }


        object "r2" {
            code {

                function bn128_is_on_curve(p0,p1) ->result {
                    let o1 := mulmod(p0, p0, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    o1 := mulmod(p0, o1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    o1 := addmod(o1,3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    let o2 := mulmod(p1,p1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    result := eq(o1,o2)
                }

                function bn128_multiply(ptr, p0, p1, p2) ->o0, o1 {
                    mstore(ptr, p0)
                    mstore(add(ptr,0x20), p1)
                    mstore(add(ptr,0x40), p2)
                    if iszero(staticcall(gas(), 0x07, ptr, 96, ptr, 64)) {
                        revert(0,0)
                    }
                    o0 := mload(ptr)
                    o1 := mload(add(ptr,0x20))
                }

                function bn128_check_pairing(ptr, paramPtr, x, y) -> result {
                    mstore(add(ptr,0xb0), x)
                    mstore(add(ptr,0xc0), y)
                    calldatacopy(ptr,paramPtr,0xb0)
                    let success := staticcall(gas(), 0x08, ptr, 384, ptr, 32)
                    if iszero(success) {
                            revert(0, 0)
                    }
                    result := mload(ptr)
                }


                function baseToG1(ptr, t)->x,y {
                    let fp := add(ptr,0xc0)
                    let ap1 := mulmod(t, t, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    let alpha := mulmod(ap1, addmod(ap1, 4, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    mstore(add(ptr, 0x60), alpha)
                    mstore(add(ptr, 0x80), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd45)
                    if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                        revert(0, 0)
                    }
                    alpha := mload(fp)
                    ap1 := mulmod(ap1, ap1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x := mulmod(ap1, 0xb3c4d79d41a91759a9e4c7e359b6b89eaec68e62effffffd, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x := mulmod(x, alpha, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                    x := addmod(x, 0x59e26bcea0d48bacd4f263f1acdb5c4f5763473177fffffe, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    let x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    mstore(add(ptr, 0x60), x_three)
                    mstore(add(ptr, 0x80), 0x183227397098d014dc2822db40c0ac2ecbc0b548b438e5469e10460b6c3e7ea3)
                    if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                        revert(0, 0)
                    }
                    if iszero(eq( mload(fp), 1)) {
                        x  := addmod(x, 1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                        x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        mstore(add(ptr, 0x60), x_three)
                        if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                            revert(0, 0)
                        }
                        if iszero(eq( mload(fp), 1)) {
                            let ap2 := addmod(ap1, 4, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x := mulmod(ap2, ap2, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x := mulmod(x, ap2, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x := mulmod(x, 0x2042def740cbc01bd03583cf0100e593ba56470b9af68708d2c05d6490535385, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x := mulmod(x, alpha, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x := sub(0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47, x)
                            x := addmod(x, 1, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x_three := mulmod(x, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x_three := mulmod(x_three, x, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                            x_three := addmod(x_three, 3, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                        }
                    }
                    mstore(add(ptr, 0x60), x_three)
                    mstore(add(ptr, 0x80), 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52)
                    if iszero(staticcall(gas(), 0x05, ptr, 0xc0, fp, 0x20)) {
                        revert(0, 0)
                    }
                    y := mload(fp)
                    if gt(t,0x183227397098d014dc2822db40c0ac2ecbc0b548b438e5469e10460b6c3e7ea3) {
                       y := mulmod(y, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd46, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    }
                }

                function HashToG1(ptr, messageptr, messagesize) ->x,y {
                    let size := add(messagesize, 1)
                    calldatacopy(add(ptr,1), messageptr, messagesize)
                    mstore8(ptr, 0x00)
                    let h0 := keccak256(ptr, size)
                    mstore8(ptr, 0x01)
                    let h1 := keccak256(ptr, size)
                    mstore8(ptr, 0x02)
                    let h2 := keccak256(ptr, size)
                    mstore8(ptr, 0x03)
                    let h3 := keccak256(ptr, size)
                    mstore(ptr, 0x20)
                    mstore(add(ptr, 0x20), 0x20)
                    mstore(add(ptr, 0x40), 0x20)
                    mstore(add(ptr, 0xa0), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    h1 := addmod(h1, mulmod(h0, 0xe0a77c19a07df2f666ea36f7879462c0a78eb28f5c70b3dd35d438dc58f0d9d, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    h2 := addmod(h3, mulmod(h2, 0xe0a77c19a07df2f666ea36f7879462c0a78eb28f5c70b3dd35d438dc58f0d9d, 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47), 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47)
                    let x1,y1 := baseToG1(ptr, h1)
                    let x2,y2 := baseToG1(ptr, h2)
                    mstore(ptr, x1)
                    mstore(add(ptr, 0x20), y1)
                    mstore(add(ptr, 0x40), x2)
                    mstore(add(ptr, 0x60), y2)
                    if iszero(staticcall(gas(), 0x06, ptr, 128, ptr, 64)) {
                        revert(0,0)
                    }
                    x:=mload(ptr)
                    y:=mload(add(ptr, 0x20))
                }

                let x,y := HashToG1(0, 0xb0, sub(calldatasize(),0xb0))
                let ok := bn128_check_pairing(0,0,x,y)
                mstore(0,ok)
                return(0,0x20)
            }
        }

    }
}