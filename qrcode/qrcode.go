package qrcode

import (
	"io"
	"os"
	"golang.org/x/text/encoding/charmap"
	"fmt"
)

type Qrcode struct {
	boxsize     int
	border      int
	maskpattern uint
	correction  uint

	version   int
	count     int
	modules   [][]Bool
	datalist  []qrdata
	datacache []byte
}

type Bool int

const (
	None  Bool = 0
	True  Bool = 1
	False Bool = 2
)

func MakeBool(b bool) Bool {
	if b {
		return True
	} else {
		return False
	}
}

func (b Bool) uint() uint {
	if b == True {
		return 1
	}

	return 0
}

func MakeQrcode(version int, correction uint, boxsize, border int, maskpattern uint) *Qrcode {
	if correction == 0 {
		correction = ERROR_CORRECT_M
	}
	if boxsize == 0 {
		boxsize = 10
	}
	if border == 0 {
		border = 4
	}

	return &Qrcode{version: version,
		border: border, boxsize: boxsize,
		correction: correction,
		maskpattern: maskpattern,}

}

func (q *Qrcode) setupPositionProbePattern(row, col int) {
	for r := -1; r < 8; r++ {
		if row+r <= -1 || q.count <= row+r {
			continue
		}

		for c := -1; c < 8; c++ {
			if col+c <= -1 || q.count <= col+c {
				continue
			}
			if (0 <= r && r <= 6 && (c == 0 || c == 6 )) ||
				(0 <= c && c <= 6 && (r == 0 || r == 6)) ||
				(2 <= r && r <= 4 && 2 <= c && c <= 4) {
				q.modules[row+r][col+c] = True
			} else {
				q.modules[row+r][col+c] = False
			}

		}
	}
}

func (q *Qrcode) setupPositionAdjustPattern() {
	pos := pattern_position(q.version)

	for i := 0; i < len(pos); i++ {
		for j := 0; j < len(pos); j++ {
			row, col := pos[i], pos[j]
			if q.modules[row][col] != None {
				continue
			}

			for r := -2; r < 3; r++ {
				for c := -2; c < 3; c++ {
					if r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0) {
						q.modules[row+r][col+c] = True
					} else {
						q.modules[row+r][col+c] = False
					}
				}
			}
		}
	}
}

func (q *Qrcode) setupTimingPattern() {
	for r := 8; r < q.count-8; r++ {
		if q.modules[r][6] != None {
			continue
		}
		if r%2 == 0 {
			q.modules[r][6] = True
		} else {
			q.modules[r][6] = False
		}
	}

	for c := 8; c < q.count-8; c++ {
		if q.modules[6][c] != None {
			continue
		}
		if c%2 == 0 {
			q.modules[6][c] = True
		} else {
			q.modules[6][c] = False
		}
	}

	pos := pattern_position(q.version)
	for i := 0; i < len(pos); i++ {
		for j := 0; j < len(pos); j++ {
			row, col := pos[i], pos[j]
			if q.modules[row][col] != None {
				continue
			}

			for r := -2; r < 3; r++ {
				for c := -2; c < 3; c++ {
					if r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0) {
						q.modules[row+r][col+c] = True
					} else {
						q.modules[row+r][col+c] = False
					}
				}
			}
		}
	}

}

func (q *Qrcode) setupTypeInfo(test bool, maskpattern uint) {
	data := (q.correction << 3) | maskpattern
	bits := BCH_type_info(data)
	fmt.Println("++", data, bits, maskpattern)
	for i := 0; i < 15; i++ {
		mode := !test && (bits>>uint(i))&1 == 0
		if i < 6 {
			q.modules[i][8] = MakeBool(mode)
		} else if i < 8 {
			q.modules[i+1][8] = MakeBool(mode)
		} else {
			q.modules[q.count-15+i][8] = MakeBool(mode)
		}
	}

	// 垂直
	for i := 0; i < 15; i++ {
		mode := !test && (bits>>uint(i))&1 == 0
		if i < 8 {
			q.modules[8][q.count-i-1] = MakeBool(mode)
		} else if i < 9 {
			q.modules[8][15-i-1] = MakeBool(mode)
		} else {
			q.modules[8][15-i+1] = MakeBool(mode)
		}
	}

	// fix
	q.modules[q.count-8][8] = MakeBool(!test)
}

func (q *Qrcode) setupTypeNumber(test bool) {
	bits := BCH_type_number(uint(q.version))

	for i := 0; i < 18; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
		q.modules[i/3][i%3+q.count-8-3] = MakeBool(mode)
	}

	for i := 0; i < 18; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
		q.modules[i%3+q.count-8-3][i/3] = MakeBool(mode)
	}
}

func (q *Qrcode) mapData(data []byte, maskpattern uint) {
	inc := -1
	row := q.count - 1

	bitIndex, byteIndex := 7, 0
	mask := mask_func(maskpattern)
	datalen := len(data)

	for col := q.count - 1; col > 0; col -= 2 {
		if col <= 6 {
			col -= 1
		}

		colrange := []int{col, col - 1}
		for {
			for _, c := range colrange {
				if q.modules[row][c] != None {
					dark := false
					if byteIndex < datalen {
						dark = (data[byteIndex]>>uint(byteIndex))&1 == 1
					}

					if mask(row, c) {
						dark = !dark
					}

					q.modules[row][c] = MakeBool(dark)
					bitIndex -= 1

					if bitIndex == -1 {
						byteIndex += 1
						bitIndex = 7
					}
				}
			}

			row += inc
			if row < 0 || q.count <= row {
				row -= inc
				inc = -inc
				break
			}
		}
	}
}

func (q *Qrcode) makeImpl(test bool, maskpattern uint) {
	q.count = q.version*4 + 17
	q.modules = make([][]Bool, q.count)

	for row := 0; row < q.count; row++ {
		q.modules[row] = make([]Bool, q.count)

		for col := 0; col < q.count; col++ {
			q.modules[row][col] = None
		}
	}
	fmt.Println("modules init success")

	q.setupPositionProbePattern(0, 0)
	q.setupPositionProbePattern(q.count-7, 0)
	q.setupPositionProbePattern(0, q.count-7)

	fmt.Println("setupPositionProbePattern success")
	q.setupPositionAdjustPattern()
	fmt.Println("setupPositionAdjustPattern success")
	q.setupTimingPattern()
	fmt.Println("setupTimingPattern success")
	q.setupTypeInfo(test, maskpattern)
	fmt.Println("setupTypeInfo success")

	if q.version >= 7 {
		q.setupTypeNumber(test)
	}

	if len(q.datacache) == 0 {
		q.datacache = create_data(q.version, q.correction, q.datalist)
	}

	q.mapData(q.datacache, maskpattern)
}

func (q *Qrcode) make(fit bool) {
	if fit || q.version == 0 {
		q.bestFit(q.version)
	}

	if q.maskpattern == 0 {
		q.makeImpl(false, q.bestMaskPattern())
	} else {
		q.makeImpl(false, q.maskpattern)
	}
}

func (q *Qrcode) bestMaskPattern() uint {
	minLostPoint := 0
	pattern := uint(0)
	fmt.Println("pattern", pattern)
	for i := uint(0); i < 8; i++ {
		q.makeImpl(true, i)
		fmt.Println("make")
		lostPoint := lost_point(q.modules)
		if i == 0 || minLostPoint > lostPoint {
			pattern = i
		}
	}

	fmt.Println("pattern", pattern)
	return pattern
}

func (q *Qrcode) bestFit(start int) int {
	if start == 0 {
		start = 1
	}

	modesizes := mode_sizes_for_version(start)
	buffer := &BitBuffer{}

	for i := range q.datalist {
		data := &q.datalist[i]
		buffer.put(data.mode, 4)
		buffer.put(data.len(), modesizes[data.mode])
		data.write(buffer)
	}

	needbits := buffer.length
	q.version = bitsectLeft(BIT_LIMIT_TABLE[q.correction], needbits, start, 0)

	newmodesizes := mode_sizes_for_version(q.version)
	if len(newmodesizes) != len(modesizes) {
		return q.bestFit(q.version)
	}

	for k, v := range newmodesizes {
		if modesizes[k] != v {
			return q.bestFit(q.version)
		}
	}

	return q.version
}

func (q *Qrcode) AddData(data interface{}, optimize int) {
	if val, ok := data.(qrdata); ok {
		q.datalist = append(q.datalist, val)
	} else {
		if optimize != 0 {
			q.datalist = append(q.datalist, optimal_data_chunks(data.([]byte), optimize)...)
		} else {
			q.datalist = append(q.datalist, *Qrdata(data.([]byte), 0, true))
		}
	}

	q.datacache = nil
}

func (q *Qrcode) PrintAscii(out io.Writer, invert bool) {
	if out == nil {
		out = os.Stdout
	}

	if q.datacache == nil {
		q.make(true)
	}

	modcount := q.count
	codes := make([]rune, 4)
	for i, ch := range []byte{255, 223, 220, 219} {
		codes[i] = charmap.CodePage437.DecodeByte(ch)
	}

	if invert {
		for l, r := 0, len(codes)-1; l <= r; {
			codes[l], codes[r] = codes[r], codes[l]
			l++
			r--
		}
	}

	getmodule := func(x, y int) uint {
		if invert && q.border != 0 && max(x, y) >= modcount+q.border {
			return 1
		}

		if min(x, y) < 0 || max(x, y) >= modcount {
			return 0
		}

		return q.modules[x][y].uint()
	}

	for r := -q.border; r < modcount+q.border; r += 2 {
		for c := -q.border; c < modcount+q.border; c += 1 {
			pos := getmodule(r, c) + (getmodule(r+1, c) << 1)
			str := string([]rune{codes[pos]})
			out.Write([]byte(str))
		}

		out.Write([]byte("\n"))
	}

}
