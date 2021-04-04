package util

func BytesToUint64(in []byte) uint64 {
	return uint64(in[0])<<56 + uint64(in[1])<<48 + uint64(in[2])<<40 + uint64(in[3])<<32 + uint64(in[4])<<24 + uint64(in[5])<<16 + uint64(in[6])<<8 + uint64(in[7])
}

func BytesToUint32(in []byte) uint32 {
	return uint32(in[0])<<24 + uint32(in[1])<<16 + uint32(in[2])<<8 + uint32(in[3])
}

func BytesToUint24(in []byte) uint32 {
	return uint32(in[0])<<16 + uint32(in[1])<<8 + uint32(in[2])
}

func BytesToUint16(in []byte) uint16 {
	return uint16(in[0])<<8 + uint16(in[1])
}
