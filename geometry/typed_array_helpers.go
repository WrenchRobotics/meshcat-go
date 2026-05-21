package geometry

import (
	"encoding/binary"
	"fmt"
	"math"
)

func itemSize2D[T any](x [][]T) (int, error) {
	if len(x) == 0 {
		return 0, fmt.Errorf("array must not be empty")
	}
	cols := len(x[0])
	for i := 1; i < len(x); i++ {
		if len(x[i]) != cols {
			return 0, fmt.Errorf("array rows must have equal length")
		}
	}
	return len(x), nil
}

func transposeFloat32(x [][]float32) [][]float32 {
	if len(x) == 0 {
		return [][]float32{}
	}
	r, c := len(x), len(x[0])
	out := make([][]float32, c)
	for i := 0; i < c; i++ {
		out[i] = make([]float32, r)
		for j := 0; j < r; j++ {
			out[i][j] = x[j][i]
		}
	}
	return out
}

func transposeUint32(x [][]uint32) [][]uint32 {
	if len(x) == 0 {
		return [][]uint32{}
	}
	r, c := len(x), len(x[0])
	out := make([][]uint32, c)
	for i := 0; i < c; i++ {
		out[i] = make([]uint32, r)
		for j := 0; j < r; j++ {
			out[i][j] = x[j][i]
		}
	}
	return out
}

func flattenFloat32ColumnMajor(x [][]float32) []byte {
	if len(x) == 0 {
		return []byte{}
	}
	rows := len(x)
	cols := len(x[0])
	out := make([]byte, 0, rows*cols*4)
	buf := make([]byte, 4)
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
			binary.LittleEndian.PutUint32(buf, mathFloat32bits(x[r][c]))
			out = append(out, buf...)
		}
	}
	return out
}

func flattenUint32ColumnMajor(x [][]uint32) []byte {
	if len(x) == 0 {
		return []byte{}
	}
	rows := len(x)
	cols := len(x[0])
	out := make([]byte, 0, rows*cols*4)
	buf := make([]byte, 4)
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
			binary.LittleEndian.PutUint32(buf, x[r][c])
			out = append(out, buf...)
		}
	}
	return out
}

func packFloat32Array2D(x [][]float32) (map[string]any, error) {
	size, err := itemSize2D(x)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"itemSize":   size,
		"type":       "Float32Array",
		"array":      flattenFloat32ColumnMajor(x),
		"normalized": false,
	}, nil
}

func packUint32Array2D(x [][]uint32) (map[string]any, error) {
	size, err := itemSize2D(x)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"itemSize":   size,
		"type":       "Uint32Array",
		"array":      flattenUint32ColumnMajor(x),
		"normalized": false,
	}, nil
}

func mathFloat32bits(f float32) uint32 {
	return math.Float32bits(f)
}
