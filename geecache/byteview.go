package geecache

// ByteView 保存不可变的字节视图。
type ByteView struct {
	b []byte // 存储字节数据的切片
}

// Len 返回视图的长度。
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回数据的字节切片的副本。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b) // 返回数据的副本，以防止外部修改数据
}

// String 返回数据的字符串表示，如果需要则进行复制。
func (v ByteView) String() string {
	return string(v.b) // 将字节切片转换为字符串
}

// cloneBytes 复制字节切片。
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
