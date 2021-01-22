package code

import "bytes"

const (
	MODE_NUMBER    = 1 << 0
	MODE_ALPHA_NUM = 1 << 1
	MODE_8BIT_BYTE = 1 << 2
	MODE_KANJI     = 1 << 3
)

var (
	MODE_SIZE_SMALL = map[int]int{
		MODE_NUMBER:    10,
		MODE_ALPHA_NUM: 9,
		MODE_8BIT_BYTE: 8,
		MODE_KANJI:     8,
	}
	MODE_SIZE_MEDIUM = map[int]int{
		MODE_NUMBER:    12,
		MODE_ALPHA_NUM: 11,
		MODE_8BIT_BYTE: 16,
		MODE_KANJI:     10,
	}
	MODE_SIZE_LARGE = map[int]int{
		MODE_NUMBER:    14,
		MODE_ALPHA_NUM: 13,
		MODE_8BIT_BYTE: 16,
		MODE_KANJI:     12,
	}
)

var (
	ALPHA_NUM = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:")

	NUMBER_LENGTH = map[int]int{
		3: 10,
		2: 7,
		1: 4,
	}

	PATTERN_POSITION_TABLE = [][]int{
		{},
		{6, 18},
		{6, 22},
		{6, 26},
		{6, 30},
		{6, 34},
		{6, 22, 38},
		{6, 24, 42},
		{6, 26, 46},
		{6, 28, 50},
		{6, 30, 54},
		{6, 32, 58},
		{6, 34, 62},
		{6, 26, 46, 66},
		{6, 26, 48, 70},
		{6, 26, 50, 74},
		{6, 30, 54, 78},
		{6, 30, 56, 82},
		{6, 30, 58, 86},
		{6, 34, 62, 90},
		{6, 28, 50, 72, 94},
		{6, 26, 50, 74, 98},
		{6, 30, 54, 78, 102},
		{6, 28, 54, 80, 106},
		{6, 32, 58, 84, 110},
		{6, 30, 58, 86, 114},
		{6, 34, 62, 90, 118},
		{6, 26, 50, 74, 98, 122},
		{6, 30, 54, 78, 102, 126},
		{6, 26, 52, 78, 104, 130},
		{6, 30, 56, 82, 108, 134},
		{6, 34, 60, 86, 112, 138},
		{6, 30, 58, 86, 114, 142},
		{6, 34, 62, 90, 118, 146},
		{6, 30, 54, 78, 102, 126, 150},
		{6, 24, 50, 76, 102, 128, 154},
		{6, 28, 54, 80, 106, 132, 158},
		{6, 32, 58, 84, 110, 136, 162},
		{6, 26, 54, 82, 110, 138, 166},
		{6, 30, 58, 86, 114, 142, 170},
	}
)

const (
	G15      = (1 << 10) | (1 << 8) | (1 << 5) | (1 << 4) | (1 << 2) | (1 << 1) | (1 << 0)
	G18      = (1 << 12) | (1 << 11) | (1 << 10) | (1 << 9) | (1 << 8) | (1 << 5) | (1 << 2) | (1 << 0)
	G15_MASK = (1 << 14) | (1 << 12) | (1 << 10) | (1 << 4) | (1 << 1)
)

const (
	PAD0 = 0xEC
	PAD1 = 0x11
)

var ()

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
	dcdata, ecdata := make([][]byte, len(rsblocks)), make([]byte, len(rsblocks))

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

		if val, ok := RSPoly_LUT[ecCount]; ok {
			rsPolly := "xxx"
		} else {
			rsPolly := "vvv"
		}

	}

}

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

type Polynomial struct {
	num []uint
}

func MakePolynomial(num []uint, shift int) *Polynomial {
	p := new(Polynomial)
	var offset int
	for offset = 0; offset < len(num); offset++ {
		if num[offset] != 0 {
			goto done
		}
	}
	offset += 1

done:
	p.num = append(num[offset:], make([]uint, shift)...)
	return p
}

func (p *Polynomial) Range() {

}

func (p *Polynomial) Mul(other Polynomial) *Polynomial {
	num := make([]uint, len(p.num)+len(other.num)-1)

	for i, ii := range p.num {
		for j, jj := range other.num {
			num[i+j] ^= gexp(glog(ii) + glog(jj))
		}
	}

	return MakePolynomial(num, 0)
}

func (p *Polynomial) Mod(other Polynomial) *Polynomial {
	diff := len(p.num) - len(other.num)
	if diff < 0 {
		return p
	}

	ratio := glog(p.num[0]) - glog(other.num[0])

	num := make([]uint, len(p.num)+len(other.num)-1)

	if diff != 0 {
		n := len(p.num)
		num = append(num, p.num[n-diff:]...)
	}
	
	return MakePolynomial(num, 0).Mod(other)
}

func gexp(n uint) uint {
	return EXP_TABLE[n%255]
}

func glog(n uint) uint {
	return LOG_TABLE[n]
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
