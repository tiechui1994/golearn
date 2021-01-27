package qrcode

import (
	"io"
	"os"
	"image/color"
	"image"
	"image/png"
	"bytes"
	"image/jpeg"
	"golang.org/x/image/bmp"
)

type Qrcode struct {
	border      int
	maskpattern uint
	correction  uint

	version   int
	count     int
	modules   [][]Bool
	datalist  []qrdata
	datacache []uint

	white, black color.Color
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

func (b Bool) String() string {
	if b == True {
		return "True"
	} else if b == False {
		return "False"
	} else {
		return "None"
	}
}

func (b Bool) MarshalJSON() ([]byte, error) {
	if b == True {
		return []byte("true"), nil
	} else if b == False {
		return []byte("false"), nil
	} else {
		return []byte("null"), nil
	}
}

func MakeQrcode(version int, correction uint, border int, maskpattern uint) *Qrcode {
	if version < 0 || version > 41 {
		version = 0
	}
	if correction < 0 || correction > ERROR_CORRECT_Q {
		correction = ERROR_CORRECT_M
	}

	if border == 0 || border < 0 {
		border = 4
	}

	return &Qrcode{
		white:       color.RGBA{R: 240, G: 230, B: 140, A: 255},
		black:       color.Black,
		version:     version,
		border:      border,
		correction:  correction,
		maskpattern: maskpattern,
	}
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
	pos := patternposition(q.version)

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

	pos := patternposition(q.version)
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
	bits := BCHTypeInfo(data)
	for i := 0; i < 15; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
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
		mode := !test && (bits>>uint(i))&1 == 1
		if i < 8 {
			q.modules[8][q.count-i-1] = MakeBool(mode)
		} else if i < 9 {
			q.modules[8][15-i-1+1] = MakeBool(mode)
		} else {
			q.modules[8][15-i-1] = MakeBool(mode)
		}
	}

	// fix
	q.modules[q.count-8][8] = MakeBool(!test)
}

func (q *Qrcode) setupTypeNumber(test bool) {
	bits := BCHTypeNumber(uint(q.version))

	for i := 0; i < 18; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
		q.modules[i/3][i%3+q.count-8-3] = MakeBool(mode)
	}

	for i := 0; i < 18; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
		q.modules[i%3+q.count-8-3][i/3] = MakeBool(mode)
	}
}

func (q *Qrcode) mapData(data []uint, maskpattern uint) {
	inc := -1
	row := q.count - 1

	bitIndex, byteIndex := 7, 0
	mask := maskfunc(maskpattern)
	datalen := len(data)
	ranges := xrange(q.count-1, 0, -2)
	for _, col := range ranges {
		if col <= 6 {
			col -= 1
		}

		colrange := []int{col, col - 1}
		for {
			for _, c := range colrange {
				if q.modules[row][c] == None {
					dark := false
					if byteIndex < datalen {
						dark = (data[byteIndex]>>uint(bitIndex))&1 == 1
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

	q.setupPositionProbePattern(0, 0)
	q.setupPositionProbePattern(q.count-7, 0)
	q.setupPositionProbePattern(0, q.count-7)
	q.setupPositionAdjustPattern()
	q.setupTimingPattern()
	q.setupTypeInfo(test, maskpattern)

	if q.version >= 7 {
		q.setupTypeNumber(test)
	}

	if len(q.datacache) == 0 {
		q.datacache = createData(q.version, q.correction, q.datalist)
	}

	q.mapData(q.datacache, maskpattern)
}

func (q *Qrcode) make(fit bool) {
	if fit || q.version == 0 {
		q.bestFit(q.version)
	}

	if q.version > 40 {
		panic("the data is too large, the QR code cannot accommodate.")
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
	for i := uint(0); i < 8; i++ {
		q.makeImpl(true, i)

		lostPoint := lostPoint(q.modules)
		if i == 0 || minLostPoint > lostPoint {
			minLostPoint = lostPoint
			pattern = i
		}
	}

	return pattern
}

func (q *Qrcode) bestFit(start int) int {
	if start == 0 {
		start = 1
	}

	modesizes := modeSizesForVersion(start)
	buffer := &BitBuffer{}

	for i := range q.datalist {
		data := &q.datalist[i]
		buffer.put(data.mode, 4)
		buffer.put(data.len(), modesizes[data.mode])
		data.write(buffer)
	}

	needbits := buffer.length
	q.version = bitsectLeft(BIT_LIMIT_TABLE[q.correction], needbits, start, 0)
	newmodesizes := modeSizesForVersion(q.version)

	diff := func(a, b map[int]int) bool {
		if len(a) != len(b) {
			return true
		}

		for k, v := range b {
			val, ok := a[k]
			if !ok || val != v {
				return true
			}
		}

		return false
	}

	if diff(newmodesizes, modesizes) {
		q.bestFit(q.version)
	}

	return q.version
}

func (q *Qrcode) AddData(data interface{}, optimize int) {
	if optimize <= 0 {
		optimize = 20
	}

	var origin []byte
	switch data.(type) {
	case qrdata:
		q.datalist = append(q.datalist, data.(qrdata))
		return
	case *qrdata:
		q.datalist = append(q.datalist, *data.(*qrdata))
		return

	case []byte:
		origin = data.([]byte)
	case string:
		origin = []byte(data.(string))
	case int:
		origin = []byte(string(data.(int)))
	case uint:
		origin = []byte(string(data.(uint)))
	default:
		panic("invalid data type")
	}

	if optimize != 0 {
		q.datalist = append(q.datalist, optimal_data_chunks(origin, optimize)...)
	} else {
		q.datalist = append(q.datalist, *Qrdata(origin, 0, true))
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

	/*
	import "golang.org/x/text/encoding/charmap"
	for i, ch := range []byte{255, 223, 220, 219} {
		codes[i] = charmap.CodePage437.DecodeByte(ch)
	}
	*/

	for i, ch := range []uint{0xa0, 0x2580, 0x2584, 0x2588} {
		codes[i] = rune(ch)
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

func (q *Qrcode) image(size int) image.Image {
	if q.datacache == nil {
		q.make(true)
	}

	white := q.white
	black := q.black

	realSize := 21 + (q.version-1)*4 + 2*q.border

	if size < 0 {
		size = size * (-1) * realSize
	}

	if size < realSize {
		size = realSize
	}

	pixelsPerModule := size / realSize

	offset := (size - realSize*pixelsPerModule) / 2

	react := image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(size, size)}

	p := color.Palette([]color.Color{white, black})
	img := image.NewPaletted(react, p)

	border := q.border
	modcount := q.count
	point := func(x0, y0 int) (x, y int, b Bool) {
		if min(x0, y0) < 0 || max(x0, y0) >= modcount {
			return x0 + border, y0 + border, False
		}

		return x0 + border, y0 + border, q.modules[x0][y0]
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			img.Set(i, j, white)
		}
	}

	for r := -q.border; r < modcount+q.border; r += 1 {
		for c := -q.border; c < modcount+q.border; c += 1 {
			x, y, value := point(r, c)

			if value == True {
				startX := x*pixelsPerModule + offset
				startY := y*pixelsPerModule + offset

				for i := startX; i < startX+pixelsPerModule; i++ {
					for j := startY; j < startY+pixelsPerModule; j++ {
						img.Set(i, j, black)
					}
				}
			}
		}
	}

	//for y, row := range q.modules {
	//	for x, v := range row {
	//		if v == True {
	//			startX := x*pixelsPerModule + offset
	//			startY := y*pixelsPerModule + offset
	//
	//			for i := startX; i < startX+pixelsPerModule; i++ {
	//				for j := startY; j < startY+pixelsPerModule; j++ {
	//					img.Set(i, j, black)
	//				}
	//			}
	//		}
	//	}
	//}

	return img
}

func (q *Qrcode) PNG(size int) ([]byte, error) {
	img := q.image(size)
	encoder := png.Encoder{CompressionLevel: png.BestCompression}

	var b bytes.Buffer
	err := encoder.Encode(&b, img)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (q *Qrcode) JPEG(size int) ([]byte, error) {
	img := q.image(size)

	var b bytes.Buffer
	err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (q *Qrcode) BMP(size int) ([]byte, error) {
	img := q.image(size)

	var b bytes.Buffer
	err := bmp.Encode(&b, img)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
