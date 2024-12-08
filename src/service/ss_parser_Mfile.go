package service

import (
	"nosqlEngine/src/models/bloom_filter"
	"nosqlEngine/src/models/key_value"
)

type SSParserMfile struct {
	mems       []MemValues
	isParsing  bool // flag to check if SS is being written
	fileWriter FileWriter
}

func NewSSParserMfile(fileWriter FileWriter) SSParser {
	return &SSParserMfile{mems: make([]MemValues, 0), isParsing: false, fileWriter: fileWriter}
}

func (ssParser *SSParserMfile) AddMemtable(keyValues []key_value.KeyValue) {
	memValues := MemValues{values: keyValues}
	ssParser.mems = append(ssParser.mems, memValues)
	ssParser.parseNextMem()
}

func (ssParser *SSParserMfile) parseNextMem() {

	/*
		Checks if SS is being written, if not, then it writes the next memtable to SS to avoid collision

		SSTable format:
		1. Data section:8 bytes for key size, key, 8 bytes for size of value, value
		2. Index section: 8 bytes for size of key, key, 8 bytes for offset in data section
		3. Summary section: 8 bytes for size of key, key, 8 bytes for offset in index section
		4. MetaData section: 8 bytes summary size, 8 bytes summary start offset,  8 bytes merkle tree size, merkle tree 8 bytes bloom filter size, bloom filter, 8 byters filter size
		
	*/
	if ssParser.isParsing {
		return
	}
	ssParser.isParsing = true

	data := ssParser.mems[0].values
	ssParser.mems = ssParser.mems[1:]

	key_value.SortByKeys(&data)

	_ = bloom_filter.GetBloomFilterArray(key_value.GetKeys(data))
//	_ = merkle_tree.GetMerkleTree(data)

	dataBytes, keys, keyOffsets := serializeDataGetOffsets(data)
	indexBytes, indexOffsets := serializeIndexGetOffsets(keys, keyOffsets, int64(0))
	summaryBytes := getSummaryBytes(key_value.GetKeys(data), indexOffsets)
	metaDataBytes := getMetaDataBytes(int64(len(summaryBytes)),int64(0), make([]byte, 0), make([]byte, 0), int64(len(data)))

	ssParser.fileWriter.WriteSS(dataBytes, indexBytes, summaryBytes, metaDataBytes)


	if len(ssParser.mems) != 0 {
		ssParser.parseNextMem()
	} else {
		ssParser.isParsing = false
	}
}
