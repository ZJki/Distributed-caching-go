package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 给定上述哈希函数，这将给出以下带有“哈希”的副本：
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")

	// 测试用例，包含一些键和预期的映射结果
	testCases := map[string]string{
		"2":  "2", // 键 2 映射到节点 2
		"11": "2", // 键 11 映射到节点 2
		"23": "4", // 键 23 映射到节点 4
		"27": "2", // 键 27 映射到节点 2
	}

	// 验证每个测试用例是否符合预期
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("询问键 %s，预期应该映射到节点 %s", k, v)
		}
	}

	// 添加节点 8，哈希环变为：2, 4, 6, 8, 12, 14, 16, 22, 24, 26, 28
	hash.Add("8")

	// 键 27 现在应该映射到节点 8
	testCases["27"] = "8"

	// 验证每个测试用例是否符合预期
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("询问键 %s，预期应该映射到节点 %s", k, v)
		}
	}
}
