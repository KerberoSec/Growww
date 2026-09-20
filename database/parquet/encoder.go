package parquet

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
)

// DictionaryEncoder encodes repeated string values into 16-bit integer indexes.
type DictionaryEncoder struct {
	IndexMap   map[string]uint16
	Dictionary []string
}

// NewDictionaryEncoder initializes an empty dictionary.
func NewDictionaryEncoder() *DictionaryEncoder {
	return &DictionaryEncoder{
		IndexMap:   make(map[string]uint16),
		Dictionary: make([]string, 0),
	}
}

// Encode converts a string to a dictionary ID.
func (d *DictionaryEncoder) Encode(val string) uint16 {
	if idx, exists := d.IndexMap[val]; exists {
		return idx
	}
	newIdx := uint16(len(d.Dictionary))
	d.Dictionary = append(d.Dictionary, val)
	d.IndexMap[val] = newIdx
	return newIdx
}

// Decode returns the string value for an index.
func (d *DictionaryEncoder) Decode(idx uint16) (string, error) {
	if int(idx) >= len(d.Dictionary) {
		return "", fmt.Errorf("dictionary index %d out of bounds", idx)
	}
	return d.Dictionary[idx], nil
}

// ColumnarBlock represents a serialized, compressed columnar block.
type ColumnarBlock struct {
	RowCount          int
	CompressedBytes   []byte
	UncompressedBytes int64
}

// SerializeColumnarTrades serializes a trade batch with dictionary encoding & Gzip compression.
func SerializeColumnarTrades(batch *ColumnarTradeBatch) (*ColumnarBlock, error) {
	if batch.Len() == 0 {
		return &ColumnarBlock{RowCount: 0, CompressedBytes: []byte{}, UncompressedBytes: 0}, nil
	}

	symDict := NewDictionaryEncoder()
	sideDict := NewDictionaryEncoder()

	var buf bytes.Buffer

	// 1. Write Header: RowCount (uint32)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(batch.Len()))

	// 2. Encode Symbols via Dictionary
	symIndices := make([]uint16, batch.Len())
	for i, s := range batch.Symbols {
		symIndices[i] = symDict.Encode(s)
	}
	// Write symbol dictionary
	_ = binary.Write(&buf, binary.LittleEndian, uint16(len(symDict.Dictionary)))
	for _, dictVal := range symDict.Dictionary {
		_ = binary.Write(&buf, binary.LittleEndian, uint16(len(dictVal)))
		buf.WriteString(dictVal)
	}
	_ = binary.Write(&buf, binary.LittleEndian, symIndices)

	// 3. Encode Sides via Dictionary
	sideIndices := make([]uint16, batch.Len())
	for i, s := range batch.Sides {
		sideIndices[i] = sideDict.Encode(s)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint16(len(sideDict.Dictionary)))
	for _, dictVal := range sideDict.Dictionary {
		_ = binary.Write(&buf, binary.LittleEndian, uint16(len(dictVal)))
		buf.WriteString(dictVal)
	}
	_ = binary.Write(&buf, binary.LittleEndian, sideIndices)

	// 4. Encode Numerical Columns
	_ = binary.Write(&buf, binary.LittleEndian, batch.TimestampsNs)
	_ = binary.Write(&buf, binary.LittleEndian, batch.PricesE8)
	_ = binary.Write(&buf, binary.LittleEndian, batch.QuantitiesE8)
	_ = binary.Write(&buf, binary.LittleEndian, batch.FeesE8)
	_ = binary.Write(&buf, binary.LittleEndian, batch.SettlementBlocks)

	rawBytes := buf.Bytes()
	uncompressedLen := int64(len(rawBytes))

	// Compress
	var compBuf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&compBuf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := gw.Write(rawBytes); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return &ColumnarBlock{
		RowCount:          batch.Len(),
		CompressedBytes:   compBuf.Bytes(),
		UncompressedBytes: uncompressedLen,
	}, nil
}

// DeserializeColumnarTrades reads back a trade batch from a serialized columnar block.
func DeserializeColumnarTrades(compressed []byte) (*ColumnarTradeBatch, error) {
	if len(compressed) == 0 {
		return NewColumnarTradeBatch(0), nil
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	var decompBuf bytes.Buffer
	if _, err := decompBuf.ReadFrom(gr); err != nil {
		return nil, err
	}

	r := bytes.NewReader(decompBuf.Bytes())

	var rowCount uint32
	if err := binary.Read(r, binary.LittleEndian, &rowCount); err != nil {
		return nil, err
	}

	n := int(rowCount)
	batch := NewColumnarTradeBatch(n)

	// 1. Read Symbol Dictionary
	var symDictLen uint16
	_ = binary.Read(r, binary.LittleEndian, &symDictLen)
	symDict := make([]string, symDictLen)
	for i := 0; i < int(symDictLen); i++ {
		var strLen uint16
		_ = binary.Read(r, binary.LittleEndian, &strLen)
		strBytes := make([]byte, strLen)
		_, _ = r.Read(strBytes)
		symDict[i] = string(strBytes)
	}
	symIndices := make([]uint16, n)
	_ = binary.Read(r, binary.LittleEndian, &symIndices)
	batch.Symbols = make([]string, n)
	for i, idx := range symIndices {
		batch.Symbols[i] = symDict[idx]
	}

	// 2. Read Side Dictionary
	var sideDictLen uint16
	_ = binary.Read(r, binary.LittleEndian, &sideDictLen)
	sideDict := make([]string, sideDictLen)
	for i := 0; i < int(sideDictLen); i++ {
		var strLen uint16
		_ = binary.Read(r, binary.LittleEndian, &strLen)
		strBytes := make([]byte, strLen)
		_, _ = r.Read(strBytes)
		sideDict[i] = string(strBytes)
	}
	sideIndices := make([]uint16, n)
	_ = binary.Read(r, binary.LittleEndian, &sideIndices)
	batch.Sides = make([]string, n)
	for i, idx := range sideIndices {
		batch.Sides[i] = sideDict[idx]
	}

	// 3. Read Numerical Columns
	batch.TimestampsNs = make([]int64, n)
	_ = binary.Read(r, binary.LittleEndian, &batch.TimestampsNs)

	batch.PricesE8 = make([]uint64, n)
	_ = binary.Read(r, binary.LittleEndian, &batch.PricesE8)

	batch.QuantitiesE8 = make([]uint64, n)
	_ = binary.Read(r, binary.LittleEndian, &batch.QuantitiesE8)

	batch.FeesE8 = make([]uint64, n)
	_ = binary.Read(r, binary.LittleEndian, &batch.FeesE8)

	batch.SettlementBlocks = make([]uint64, n)
	_ = binary.Read(r, binary.LittleEndian, &batch.SettlementBlocks)

	// Populate placeholder TradeIDs
	batch.TradeIDs = make([]string, n)
	for i := 0; i < n; i++ {
		batch.TradeIDs[i] = fmt.Sprintf("trd_%d", i)
	}

	return batch, nil
}
