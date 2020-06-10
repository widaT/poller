package interest

type Interest uint8

const READABLE Interest = 0b0001
const WRITABLE Interest = 0b0010

func (i Interest) IsReadable() bool {
	return (i & READABLE) != 0
}

func (i Interest) IsWritable() bool {
	return (i & WRITABLE) != 0
}
func (i Interest) Add(other Interest) Interest {
	return i | other
}
