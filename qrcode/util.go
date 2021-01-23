package qrcode

import (
	"bytes"
	"fmt"
	"regexp"
	"encoding/hex"
	"math"
)

func BCH_type_info(data uint) uint {
	d := data << 10
	fmt.Println(data)
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

//=============================================================================

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

func (q *qrdata) write(buffer *BitBuffer) {
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

func (q qrdata) String() string {
	switch q.mode {
	case MODE_ALPHA_NUM, MODE_NUMBER:
		return string(q.data)
	default:
		return hex.EncodeToString(q.data)
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

	if len(bit.buffer) <= idx {
		bit.buffer = append(bit.buffer, 0)
	}

	if b {
		bit.buffer[idx] |= 0x80 >> uint(bit.length%8)
	}

	bit.length += 1
}

func create_data(version int, correction uint, qrdatas []qrdata) []byte {
	buffer := &BitBuffer{}
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

func create_bytes(buffer *BitBuffer, rsblocks []RSBlock) []byte {
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

//=============================================================================

func optimal_data_chunks(data []byte, minimum int) []qrdata {
	alpha := "[0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ%/: \\$\\*\\+\\-\\.]"
	num := "\\d"

	if len(data) <= minimum {
		alpha = "^" + alpha + "+$"
		num = "^" + num + "+$"
	} else {
		alpha = fmt.Sprintf("%s{%d,}", alpha, minimum)
		num = fmt.Sprintf("%s{%d,}", num, minimum)
	}

	rnum := regexp.MustCompile(num)
	ralpha := regexp.MustCompile(alpha)

	split := func(data string, pattern *regexp.Regexp) func() (bool, bool, string) {
		hasnext := true
		return func() (is, next bool, match string) {
			if !hasnext || len(data) == 0 {
				return false, false, ""
			}

			if !pattern.MatchString(data) {
				hasnext = false
				return false, true, data
			}

			indexs := pattern.FindAllStringIndex(data, 1)
			start, end := indexs[0][0], indexs[0][1]
			if start != 0 {
				match = data[:start]
				data = data[start:]
				return false, true, match
			}

			match = data[start:end]
			data = data[end:]
			return true, true, match
		}
	}

	var (
		ans                     []qrdata
		mode                    int
		isnum, isalpha, hasnext bool
		chunk, subchunk         string
	)

	origin := string(data)
	numfunc := split(origin, rnum)

num:
	isnum, hasnext, chunk = numfunc()
	if !hasnext {
		goto done
	}
	if isnum {
		ans = append(ans, *Qrdata([]byte(chunk), MODE_NUMBER, false))
	} else {
		alphafunc := split(chunk, ralpha)
	alpha:
		isalpha, hasnext, subchunk = alphafunc()
		if !hasnext {
			goto num
		}
		if isalpha {
			mode = MODE_ALPHA_NUM
		} else {
			mode = MODE_8BIT_BYTE
		}
		ans = append(ans, *Qrdata([]byte(subchunk), mode, false))
		goto alpha
	}

	goto num

done:
	return ans
}

//=============================================================================

func lost_point(modules [][]Bool) int {
	modcount := len(modules)
	lostpoint := 0

	lostpoint += lostPointLevel1(modules, modcount)
	fmt.Println("lostPointLevel1", lostpoint)
	lostpoint += lostPointLevel2(modules, modcount)
	lostpoint += lostPointLevel3(modules, modcount)
	lostpoint += lostPointLevel4(modules, modcount)

	return lostpoint
}

func lostPointLevel1(modules [][]Bool, modcount int) int {
	lostpoint := 0

	container := make([]int, modcount+1)

	for row := 0; row < modcount; row++ {
		thisRow := modules[row]
		previewColor := thisRow[0]

		length := 0

		for col := 0; col < modcount; col++ {
			if thisRow[col] == previewColor {
				length += 1
			} else {
				if length >= 5 {
					container[length] += 1
				}

				length = 1
				previewColor = thisRow[col]
			}
		}

		if length >= 5 {
			container[length] += 1
		}

	}

	for col := 0; col < modcount; col++ {
		previewColor := modules[0][col]
		length := 0

		for row := 0; row < modcount; row++ {
			if modules[row][col] == previewColor {
				length += 1
			} else {
				if length >= 5 {
					container[length] += 1
				}

				length = 1
				previewColor = modules[row][col]
			}
		}

		if length >= 5 {
			container[length] += 1
		}

	}

	var sum int
	for i := 5; i < modcount+1; i++ {
		sum += container[i] * (i - 2)
	}

	lostpoint += int(sum)

	return lostpoint
}

func lostPointLevel2(modules [][]Bool, modcount int) int {
	lostpoint := 0

	for row := 0; row < modcount-1; row++ {
		thisRow := modules[row]
		nextRow := modules[row+1]

		col := 0
		iter := iter(modcount - 1)
		for next(iter, &col) < modcount-1 {
			topright := thisRow[col+1]
			if topright != nextRow[col+1] {
				skip(iter)
			} else if topright != thisRow[col] {
				continue
			} else if topright != nextRow[col] {
				continue
			} else {
				lostpoint += 3
			}
		}
	}

	return lostpoint
}

func lostPointLevel3(modules [][]Bool, modcount int) int {
	lostpoint := 0

	for row := 0; row < modcount; row++ {
		thisRow := modules[row]

		col := 0
		iter := iter(modcount - 10)
		for next(iter, &col) < modcount-10 {
			if thisRow[col+1] != True &&
				thisRow[col+4] == True &&
				thisRow[col+5] != True &&
				thisRow[col+6] == True &&
				thisRow[col+9] != True &&
				(
					thisRow[col+0] == True &&
						thisRow[col+2] == True &&
						thisRow[col+3] == True &&
						thisRow[col+7] != True &&
						thisRow[col+8] != True &&
						thisRow[col+10] != True ||

						thisRow[col+0] != True &&
							thisRow[col+2] != True &&
							thisRow[col+3] != True &&
							thisRow[col+7] == True &&
							thisRow[col+8] == True &&
							thisRow[col+10] == True) {
				lostpoint += 40
			}

			if thisRow[col+10] == True {
				skip(iter)
			}
		}
	}

	for col := 0; col < modcount; col++ {

		row := 0
		iter := iter(modcount - 10)
		for next(iter, &row) < modcount-10 {
			if modules[row+1][col] != True &&
				modules[row+4][col] == True &&
				modules[row+5][col] != True &&
				modules[row+6][col] == True &&
				modules[row+9][col] != True &&
				(
					modules[row+0][col] == True &&
						modules[row+2][col] == True &&
						modules[row+3][col] == True &&
						modules[row+7][col] != True &&
						modules[row+8][col] != True &&
						modules[row+10][col] != True ||

						modules[row+0][col] != True &&
							modules[row+2][col] != True &&
							modules[row+3][col] != True &&
							modules[row+7][col] == True &&
							modules[row+8][col] == True &&
							modules[row+10][col] == True) {
				lostpoint += 40
			}

			if modules[row+10][col] == True {
				skip(iter)
			}
		}
	}

	return lostpoint
}

func lostPointLevel4(modules [][]Bool, modcount int) int {
	darkcount := 0
	for row := 0; row < len(modules); row++ {
		for col := 0; col < len(modules[row]); col++ {
			if modules[row][col] == True {
				darkcount += 1
			}
		}
	}

	percent := float64(darkcount) / float64(modcount*modcount)
	rating := int(math.Abs(percent*100-50) / 5)
	return rating * 10
}

//==================================================================================================

type iteror struct {
	len   int
	visit []int
}

func iter(len int) *iteror {
	if len <= 0 {
		panic("invalid len")
	}

	i := &iteror{len: len}

	visit := make([]int, len)
	for i := 0; i < len; i++ {
		visit[i] = i
	}

	i.visit = visit
	return i
}

func next(i *iteror, idx *int) int {
	if len(i.visit) == 0 {
		return i.len
	}

	*idx = i.visit[0]
	i.visit = i.visit[1:]
	return *idx
}

func skip(i *iteror) {
	if len(i.visit) == 0 {
		return
	}

	if len(i.visit) >= 1 {
		i.visit = i.visit[1:]
	}
}

func bitsectLeft(arr []int, x, lo, hi int) int {
	if hi == 0 {
		hi = len(arr)
	}

	for lo < hi {
		mid := (lo + hi) / 2
		if arr[mid] < x {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	return lo
}
