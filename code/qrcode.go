package code

type Qrcode struct {
	error_correction uint
	version          int
	modules_count    int
	modules          [][]Bool
	data_cache       []byte
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

func (q *Qrcode) setup_position_probe_pattern(row, col int) {
	for r := -1; r < 8; r++ {
		if row+r <= -1 || q.modules_count <= row+r {
			continue
		}

		for c := -1; c < 8; c++ {
			if col+c <= -1 || q.modules_count <= col+c {
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

func (q *Qrcode) setup_position_adjust_pattern() {
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

func (q *Qrcode) setup_timing_pattern() {
	for r := 8; r < q.modules_count-8; r++ {
		if q.modules[r][6] != None {
			continue
		}
		if r%2 == 0 {
			q.modules[r][6] = True
		} else {
			q.modules[r][6] = False
		}
	}

	for c := 8; c < q.modules_count-8; c++ {
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

func (q *Qrcode) setup_type_info(test bool, maskpattern uint) {
	data := (q.error_correction << 3) | maskpattern
	bits := BCH_type_info(data)

	for i := 0; i < 15; i++ {
		mode := !test && (bits>>uint(i))&1 == 0
		if i < 6 {
			q.modules[i][8] = MakeBool(mode)
		} else if i < 8 {
			q.modules[i+1][8] = MakeBool(mode)
		} else {
			q.modules[q.modules_count-15+i][8] = MakeBool(mode)
		}
	}

	// 垂直
	for i := 0; i < 15; i++ {
		mode := !test && (bits>>uint(i))&1 == 0
		if i < 8 {
			q.modules[8][q.modules_count-i-1] = MakeBool(mode)
		} else if i < 9 {
			q.modules[8][15-i-1] = MakeBool(mode)
		} else {
			q.modules[8][15-i+1] = MakeBool(mode)
		}
	}

	// fix
	q.modules[q.modules_count-8][8] = MakeBool(!test)
}

func (q *Qrcode) setup_type_number(test bool) {
	bits := BCH_type_number(uint(q.version))

	for i := 0; i < 18; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
		q.modules[i/3][i%3+q.modules_count-8-3] = MakeBool(mode)
	}

	for i := 0; i < 18; i++ {
		mode := !test && (bits>>uint(i))&1 == 1
		q.modules[i%3+q.modules_count-8-3][i/3] = MakeBool(mode)
	}
}

func (q *Qrcode) map_data(data []byte, maskpattern uint) {
	inc := -1
	row := q.modules_count - 1

	bitIndex, byteIndex := 7, 0
	mask := mask_func(maskpattern)
	datalen := len(data)

	for col := q.modules_count - 1; col > 0; col -= 2 {
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
			if row < 0 || q.modules_count <= row {
				row -= inc
				inc = -inc
				break
			}
		}
	}
}

func (q *Qrcode) makeImpl(test bool, maskpattern uint) {
	q.modules_count = q.version*4 + 17
	q.modules = make([][]Bool, q.modules_count)

	for row := 0; row < q.modules_count; row++ {
		q.modules[row] = make([]Bool, q.modules_count)

		for col := 0; col < q.modules_count; col++ {
			q.modules[row][col] = None
		}
	}

	q.setup_position_probe_pattern(0, 0)
	q.setup_position_probe_pattern(q.modules_count-7, 0)
	q.setup_position_probe_pattern(0, q.modules_count-7)
	q.setup_position_adjust_pattern()
	q.setup_timing_pattern()
	q.setup_type_info(test, maskpattern)

	if q.version >= 7 {
		q.setup_type_number(test)
	}

	if len(q.data_cache) == 0 {
		q.data_cache = nil
	}
}
