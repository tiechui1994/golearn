package code

type Polynomial struct {
	num []byte
}

func MakePolynomial(num []byte, shift int) *Polynomial {
	p := new(Polynomial)
	var offset int
	for offset = 0; offset < len(num); offset++ {
		if num[offset] != 0 {
			goto done
		}
	}
	offset += 1

done:
	p.num = append(num[offset:], make([]byte, shift)...)
	return p
}

func (p *Polynomial) Mul(other *Polynomial) *Polynomial {
	num := make([]byte, len(p.num)+len(other.num)-1)

	for i, ii := range p.num {
		for j, jj := range other.num {
			num[i+j] ^= gexp(glog(ii) + glog(jj))
		}
	}

	return MakePolynomial(num, 0)
}

func (p *Polynomial) Mod(other *Polynomial) *Polynomial {
	diff := len(p.num) - len(other.num)
	if diff < 0 {
		return p
	}

	ratio := glog(p.num[0]) - glog(other.num[0])

	var num []byte
	for _, v := range zip(p.num, other.num) {
		num = append(num, v[0]^gexp(glog(v[1])+ratio))
	}

	if diff != 0 {
		n := len(p.num)
		num = append(num, p.num[n-diff:]...)
	}

	return MakePolynomial(num, 0).Mod(other)
}

func (p *Polynomial) Len() int {
	return len(p.num)
}

func gexp(n byte) byte {
	return EXP_TABLE[n%255]
}

func glog(n byte) byte {
	return LOG_TABLE[n]
}

func zip(a, b []byte) [][2]byte {
	al := len(a)
	bl := len(b)

	var length int
	if al > bl {
		length = bl
		a = a[:length]
	} else if al < bl {
		length = al
		b = b[:length]
	} else {
		length = al
	}

	var ans = make([][2]byte, length)
	for i := 0; i < length; i++ {
		ans[i] = [...]byte{a[i], b[i]}
	}

	return ans
}

//====================================================================

type RSBlock struct {
	totalcount int
	datacount  int
}

func rs_blocks(version, correction int) []RSBlock {
	offset := RS_BLOCK_OFFSET[correction]
	rsblock := RS_BLOCK_TABLE[(version-1)*4+offset]

	var blocks []RSBlock
	for i := 0; i < len(rsblock); i += 3 {
		count, totalcount, datacount := rsblock[i], rsblock[i+1], rsblock[i+2]
		for j := 0; j < count; j++ {
			blocks = append(blocks, RSBlock{totalcount, datacount})
		}
	}

	return blocks
}
