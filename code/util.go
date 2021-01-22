package code

import "bytes"

func BCH_type_info(data uint) uint {
	d := data << 10
	for BCH_digit(d)-BCH_digit(G15) >= 0 {
		d ^= G15 << (BCH_digit(d) - BCH_digit(G15))
	}
	return ((data << 10) | d) ^ G15_MASK
}

func BCH_type_number(data uint) uint {
	d := data << 12
	for BCH_digit(d)-BCH_digit(G18) >= 0 {
		d ^= G18 << (BCH_digit(d) - BCH_digit(G18))
	}
	return (data << 12) | d
}

func BCH_digit(data uint) uint {
	var digit uint
	for data != 0 {
		digit += 1
		data >>= 1
	}

	return digit
}

func pattern_position(version int) []int {
	return PATTERN_POSITION_TABLE[version-1]
}

func mask_func(pattern uint) func(i, j int) bool {
	switch pattern {
	case 0:
		return func(i, j int) bool {
			return (i+j)%2 == 0
		}
	case 1:
		return func(i, j int) bool {
			return i%2 == 0
		}
	case 2:
		return func(i, j int) bool {
			return j%3 == 0
		}
	case 3:
		return func(i, j int) bool {
			return (i+j)%3 == 0
		}
	case 4:
		return func(i, j int) bool {
			return (i+j)%2 == 0
		}
	case 5:
		return func(i, j int) bool {
			return (i*j)%2+(i*j)%3 == 0
		}
	case 6:
		return func(i, j int) bool {
			return ((i*j)%2+(i*j)%3)%2 == 0
		}
	case 7:
		return func(i, j int) bool {
			return ((i*j)%3+(i+j)%2)%2 == 0
		}
	}

	panic("bad pattern")
}

func mode_sizes_for_version(version int) map[int]int {
	if version < 10 {
		return MODE_SIZE_SMALL
	} else if version < 27 {
		return MODE_SIZE_MEDIUM
	} else {
		return MODE_SIZE_LARGE
	}
}

func length_in_bits(mode, version int) int {
	return mode_sizes_for_version(version)[mode]
}

func lost_point(modules [][]int) {

}

//============

type qrdata struct {
	mode int
	data []byte
}

func Qrdata(data []byte, mode int, check bool) *qrdata {
	d := new(qrdata)
	d.mode = mode
	if mode != MODE_NUMBER && mode != MODE_ALPHA_NUM && mode != MODE_8BIT_BYTE {
		panic("Invalid mode")
	}

	d.data = data
	return d
}

func (q *qrdata) len() int {
	return len(q.data)
}

func (q *qrdata) write(buffer BitBuffer) {
	if q.mode == MODE_NUMBER {
		for i := 0; i < len(q.data); i += 3 {
			chars := q.data[i:i+3]
			bitlen := NUMBER_LENGTH[len(chars)]
			buffer.put(int(len(chars)), bitlen) // TODO:
		}
	} else if q.mode == MODE_ALPHA_NUM {
		for i := 0; i < len(q.data); i += 3 {
			chars := q.data[i:i+2]
			if len(chars) > 1 {
				idx0 := bytes.IndexByte(ALPHA_NUM, chars[0])
				idx1 := bytes.IndexByte(ALPHA_NUM, chars[1])
				buffer.put(idx0*45+idx1, 11)
			} else {
				idx0 := bytes.Index(ALPHA_NUM, chars)
				buffer.put(idx0, 6)
			}
		}
	} else {
		data := q.data
		for _, c := range data {
			buffer.put(int(c), 8)
		}
	}
}

type BitBuffer struct {
	buffer []byte
	length int
}

func (bit *BitBuffer) get(index uint) bool {
	idx := index / 8
	return (bit.buffer[idx]>>(7-index%8))&1 == 1
}

func (bit *BitBuffer) put(num int, length int) {
	for i := 0; i < length; i++ {
		b := (num>>(uint(length-i)-1))&1 == 1
		bit.putbit(b)
	}
}

func (bit *BitBuffer) putbit(b bool) {
	idx := bit.length / 8

	if bit.length <= idx {
		bit.buffer = append(bit.buffer, 0)
	}

	if b {
		bit.buffer[idx] |= 0x80 >> uint(bit.length%8)
	}

	bit.length += 1
}

func create_data(version int, correction uint, qrdatas []qrdata) []byte {
	buffer := BitBuffer{}
	for i := range qrdatas {
		data := &qrdatas[i]
		buffer.put(data.mode, 4)
		buffer.put(data.len(), length_in_bits(data.mode, version))
		data.write(buffer)
	}

	rsblocks := rs_blocks(version, int(correction))
	bitlimit := 0

	for _, v := range rsblocks {
		bitlimit += v.datacount * 8
	}

	if buffer.length > bitlimit {
		panic("data overflow")
	}

	n := min(bitlimit-buffer.length, 4)
	for i := 0; i < n; i++ {
		buffer.putbit(false)
	}

	delimit := buffer.length % 8
	if delimit != 0 {
		for i := 0; i < 8-delimit; i++ {
			buffer.putbit(false)
		}
	}

	bytetofill := (bitlimit - buffer.length) / 8
	for i := 0; i < bytetofill; i++ {
		if i%2 == 0 {
			buffer.put(PAD0, 8)
		} else {
			buffer.put(PAD1, 8)
		}
	}

	return create_bytes(buffer, rsblocks)
}

func create_bytes(buffer BitBuffer, rsblocks []RSBlock) []byte {
	offset := 0

	maxDcCount, maxEcCount := 0, 0
	dcdata, ecdata := make([][]byte, len(rsblocks)), make([][]byte, len(rsblocks))

	for r := 0; r < len(rsblocks); r++ {
		dcCount := rsblocks[r].datacount
		ecCount := rsblocks[r].totalcount - dcCount

		maxDcCount = max(maxDcCount, dcCount)
		maxEcCount = max(maxEcCount, ecCount)

		dcdata[r] = make([]byte, dcCount)

		for i := 0; i < len(dcdata[r]); i++ {
			dcdata[r][i] = 0xff & buffer.buffer[i+offset]
		}

		offset += dcCount

		var rsPolly *Polynomial
		if val, ok := RSPoly_LUT[ecCount]; ok {
			rsPolly = MakePolynomial(val, 0)
		} else {
			rsPolly = MakePolynomial([]byte{1}, 0)

			for i := byte(0); i < byte(ecCount); i++ {
				rsPolly = rsPolly.Mul(MakePolynomial([]byte{1, gexp(i)}, 0))
			}
		}

		rawPoly := MakePolynomial(dcdata[r], rsPolly.Len()-1)
		modPoly := rawPoly.Mod(rsPolly)

		ecdata[r] = make([]byte, rsPolly.Len()-1)
		for i := 0; i < len(ecdata[r]); i++ {
			modIndex := i + modPoly.Len() - len(ecdata[r])
			if modIndex >= 0 {
				ecdata[r][i] = modPoly.num[modIndex]
			} else {
				ecdata[r][i] = 0
			}
		}
	}

	totalCodeCount := 0
	for _, rsblock := range rsblocks {
		totalCodeCount += rsblock.totalcount
	}

	data := make([]byte, totalCodeCount)
	index := 0

	for i := 0; i < maxDcCount; i++ {
		for r := 0; r < len(rsblocks); r++ {
			if i < len(dcdata[r]) {
				data[index] = dcdata[r][i]
				index += 1
			}
		}
	}

	for i := 0; i < maxEcCount; i++ {
		for r := 0; r < len(rsblocks); r++ {
			if i < len(dcdata[r]) {
				data[index] = ecdata[r][i]
				index += 1
			}
		}
	}

	return data
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

//====================

func optimal_data_chunks(data []byte, minimum int) qrdata {
	
	return qrdata{}
}
