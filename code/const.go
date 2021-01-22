package code

const (
	ERROR_CORRECT_L = 1
	ERROR_CORRECT_M = 0
	ERROR_CORRECT_Q = 3
	ERROR_CORRECT_H = 2
)

var (
	RS_BLOCK_OFFSET = map[int]int{
		ERROR_CORRECT_L: 0,
		ERROR_CORRECT_M: 1,
		ERROR_CORRECT_Q: 2,
		ERROR_CORRECT_H: 3,
	}

	RS_BLOCK_TABLE = [][]int{
		// L
		// M
		// Q
		// H

		// 1
		{1, 26, 19},
		{1, 26, 16},
		{1, 26, 13},
		{1, 26, 9},

		// 2
		{1, 44, 34},
		{1, 44, 28},
		{1, 44, 22},
		{1, 44, 16},

		// 3
		{1, 70, 55},
		{1, 70, 44},
		{2, 35, 17},
		{2, 35, 13},

		// 4
		{1, 100, 80},
		{2, 50, 32},
		{2, 50, 24},
		{4, 25, 9},

		// 5
		{1, 134, 108},
		{2, 67, 43},
		{2, 33, 15, 2, 34, 16},
		{2, 33, 11, 2, 34, 12},

		// 6
		{2, 86, 68},
		{4, 43, 27},
		{4, 43, 19},
		{4, 43, 15},

		// 7
		{2, 98, 78},
		{4, 49, 31},
		{2, 32, 14, 4, 33, 15},
		{4, 39, 13, 1, 40, 14},

		// 8
		{2, 121, 97},
		{2, 60, 38, 2, 61, 39},
		{4, 40, 18, 2, 41, 19},
		{4, 40, 14, 2, 41, 15},

		// 9
		{2, 146, 116},
		{3, 58, 36, 2, 59, 37},
		{4, 36, 16, 4, 37, 17},
		{4, 36, 12, 4, 37, 13},

		// 10
		{2, 86, 68, 2, 87, 69},
		{4, 69, 43, 1, 70, 44},
		{6, 43, 19, 2, 44, 20},
		{6, 43, 15, 2, 44, 16},

		// 11
		{4, 101, 81},
		{1, 80, 50, 4, 81, 51},
		{4, 50, 22, 4, 51, 23},
		{3, 36, 12, 8, 37, 13},

		// 12
		{2, 116, 92, 2, 117, 93},
		{6, 58, 36, 2, 59, 37},
		{4, 46, 20, 6, 47, 21},
		{7, 42, 14, 4, 43, 15},

		// 13
		{4, 133, 107},
		{8, 59, 37, 1, 60, 38},
		{8, 44, 20, 4, 45, 21},
		{12, 33, 11, 4, 34, 12},

		// 14
		{3, 145, 115, 1, 146, 116},
		{4, 64, 40, 5, 65, 41},
		{11, 36, 16, 5, 37, 17},
		{11, 36, 12, 5, 37, 13},

		// 15
		{5, 109, 87, 1, 110, 88},
		{5, 65, 41, 5, 66, 42},
		{5, 54, 24, 7, 55, 25},
		{11, 36, 12, 7, 37, 13},

		// 16
		{5, 122, 98, 1, 123, 99},
		{7, 73, 45, 3, 74, 46},
		{15, 43, 19, 2, 44, 20},
		{3, 45, 15, 13, 46, 16},

		// 17
		{1, 135, 107, 5, 136, 108},
		{10, 74, 46, 1, 75, 47},
		{1, 50, 22, 15, 51, 23},
		{2, 42, 14, 17, 43, 15},

		// 18
		{5, 150, 120, 1, 151, 121},
		{9, 69, 43, 4, 70, 44},
		{17, 50, 22, 1, 51, 23},
		{2, 42, 14, 19, 43, 15},

		// 19
		{3, 141, 113, 4, 142, 114},
		{3, 70, 44, 11, 71, 45},
		{17, 47, 21, 4, 48, 22},
		{9, 39, 13, 16, 40, 14},

		// 20
		{3, 135, 107, 5, 136, 108},
		{3, 67, 41, 13, 68, 42},
		{15, 54, 24, 5, 55, 25},
		{15, 43, 15, 10, 44, 16},

		// 21
		{4, 144, 116, 4, 145, 117},
		{17, 68, 42},
		{17, 50, 22, 6, 51, 23},
		{19, 46, 16, 6, 47, 17},

		// 22
		{2, 139, 111, 7, 140, 112},
		{17, 74, 46},
		{7, 54, 24, 16, 55, 25},
		{34, 37, 13},

		// 23
		{4, 151, 121, 5, 152, 122},
		{4, 75, 47, 14, 76, 48},
		{11, 54, 24, 14, 55, 25},
		{16, 45, 15, 14, 46, 16},

		// 24
		{6, 147, 117, 4, 148, 118},
		{6, 73, 45, 14, 74, 46},
		{11, 54, 24, 16, 55, 25},
		{30, 46, 16, 2, 47, 17},

		// 25
		{8, 132, 106, 4, 133, 107},
		{8, 75, 47, 13, 76, 48},
		{7, 54, 24, 22, 55, 25},
		{22, 45, 15, 13, 46, 16},

		// 26
		{10, 142, 114, 2, 143, 115},
		{19, 74, 46, 4, 75, 47},
		{28, 50, 22, 6, 51, 23},
		{33, 46, 16, 4, 47, 17},

		// 27
		{8, 152, 122, 4, 153, 123},
		{22, 73, 45, 3, 74, 46},
		{8, 53, 23, 26, 54, 24},
		{12, 45, 15, 28, 46, 16},

		// 28
		{3, 147, 117, 10, 148, 118},
		{3, 73, 45, 23, 74, 46},
		{4, 54, 24, 31, 55, 25},
		{11, 45, 15, 31, 46, 16},

		// 29
		{7, 146, 116, 7, 147, 117},
		{21, 73, 45, 7, 74, 46},
		{1, 53, 23, 37, 54, 24},
		{19, 45, 15, 26, 46, 16},

		// 30
		{5, 145, 115, 10, 146, 116},
		{19, 75, 47, 10, 76, 48},
		{15, 54, 24, 25, 55, 25},
		{23, 45, 15, 25, 46, 16},

		// 31
		{13, 145, 115, 3, 146, 116},
		{2, 74, 46, 29, 75, 47},
		{42, 54, 24, 1, 55, 25},
		{23, 45, 15, 28, 46, 16},

		// 32
		{17, 145, 115},
		{10, 74, 46, 23, 75, 47},
		{10, 54, 24, 35, 55, 25},
		{19, 45, 15, 35, 46, 16},

		// 33
		{17, 145, 115, 1, 146, 116},
		{14, 74, 46, 21, 75, 47},
		{29, 54, 24, 19, 55, 25},
		{11, 45, 15, 46, 46, 16},

		// 34
		{13, 145, 115, 6, 146, 116},
		{14, 74, 46, 23, 75, 47},
		{44, 54, 24, 7, 55, 25},
		{59, 46, 16, 1, 47, 17},

		// 35
		{12, 151, 121, 7, 152, 122},
		{12, 75, 47, 26, 76, 48},
		{39, 54, 24, 14, 55, 25},
		{22, 45, 15, 41, 46, 16},

		// 36
		{6, 151, 121, 14, 152, 122},
		{6, 75, 47, 34, 76, 48},
		{46, 54, 24, 10, 55, 25},
		{2, 45, 15, 64, 46, 16},

		// 37
		{17, 152, 122, 4, 153, 123},
		{29, 74, 46, 14, 75, 47},
		{49, 54, 24, 10, 55, 25},
		{24, 45, 15, 46, 46, 16},

		// 38
		{4, 152, 122, 18, 153, 123},
		{13, 74, 46, 32, 75, 47},
		{48, 54, 24, 14, 55, 25},
		{42, 45, 15, 32, 46, 16},

		// 39
		{20, 147, 117, 4, 148, 118},
		{40, 75, 47, 7, 76, 48},
		{43, 54, 24, 22, 55, 25},
		{10, 45, 15, 67, 46, 16},

		// 40
		{19, 148, 118, 6, 149, 119},
		{18, 75, 47, 31, 76, 48},
		{34, 54, 24, 34, 55, 25},
		{20, 45, 15, 61, 46, 16},
	}
)

var (
	RSPoly_LUT = map[int][]byte{
		7:  {1, 127, 122, 154, 164, 11, 68, 117},
		10: {1, 216, 194, 159, 111, 199, 94, 95, 113, 157, 193},
		13: {1, 137, 73, 227, 17, 177, 17, 52, 13, 46, 43, 83, 132, 120},
		15: {1, 29, 196, 111, 163, 112, 74, 10, 105, 105, 139, 132, 151,
			32, 134, 26},
		16: {1, 59, 13, 104, 189, 68, 209, 30, 8, 163, 65, 41, 229, 98, 50, 36, 59},
		17: {1, 119, 66, 83, 120, 119, 22, 197, 83, 249, 41, 143, 134, 85, 53, 125,
			99, 79},
		18: {1, 239, 251, 183, 113, 149, 175, 199, 215, 240, 220, 73, 82, 173, 75,
			32, 67, 217, 146},
		20: {1, 152, 185, 240, 5, 111, 99, 6, 220, 112, 150, 69, 36, 187, 22, 228,
			198, 121, 121, 165, 174},
		22: {1, 89, 179, 131, 176, 182, 244, 19, 189, 69, 40, 28, 137, 29, 123, 67,
			253, 86, 218, 230, 26, 145, 245},
		24: {1, 122, 118, 169, 70, 178, 237, 216, 102, 115, 150, 229, 73, 130, 72,
			61, 43, 206, 1, 237, 247, 127, 217, 144, 117},
		26: {1, 246, 51, 183, 4, 136, 98, 199, 152, 77, 56, 206, 24, 145, 40, 209,
			117, 233, 42, 135, 68, 70, 144, 146, 77, 43, 94},
		28: {1, 252, 9, 28, 13, 18, 251, 208, 150, 103, 174, 100, 41, 167, 12, 247,
			56, 117, 119, 233, 127, 181, 100, 121, 147, 176, 74, 58, 197},
		30: {1, 212, 246, 77, 73, 195, 192, 75, 98, 5, 70, 103, 177, 22, 217, 138,
			51, 181, 246, 72, 25, 18, 46, 228, 74, 216, 195, 11, 106, 130, 150},
	}
)

var (
	EXP_TABLE [256]uint
	LOG_TABLE [256]uint
)

func init() {
	for i := 0; i < 256; i++ {
		EXP_TABLE[i] = uint(i)
		LOG_TABLE[i] = uint(i)
	}

	for i := 0; i < 8; i++ {
		EXP_TABLE[i] = 1 << uint(i)
	}

	for i := 8; i < 256; i++ {
		EXP_TABLE[i] = EXP_TABLE[i-4] ^ EXP_TABLE[i-5] ^ EXP_TABLE[i-6] ^ EXP_TABLE[i-8]
	}

	for i:=0; i<256;i++{
		LOG_TABLE[EXP_TABLE[i]] = uint(i)
	}
}
