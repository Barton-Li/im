package set

// Union 函数用于计算两个切片的并集。
// 它支持 uint、int 和 string 类型的切片。
// 参数 slice1 和 slice2 是两个需要合并的切片。
// 返回值是一个包含两个切片中所有不重复元素的并集切片。
func Union[T uint | int | string](slice1, slice2 []T) []T {
	// 使用 map 来记录元素出现的次数，用于去重。
	m := make(map[T]int)

	// 遍历第一个切片，将元素及其出现次数记录到 map 中。
	for _, v := range slice1 {
		m[v]++
	}

	// 遍历第二个切片，检查元素在 map 中的出现次数。
	// 如果出现次数为 0，则将该元素添加到结果切片中。
	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}

	// 返回合并后的切片。
	return slice1
}

// Intersect 计算两个切片的交集，并返回结果。
// 参数 slice1 和 slice2 是两个需要计算交集的切片。
// 返回值是包含交集元素的新切片。
// 支持的元素类型为 uint, int, string 和 uint32。
func Intersect[T uint | int | string | uint32](slice1, slice2 []T) []T {
	// 使用 map 来统计第一个切片中每个元素的出现次数。
	m := make(map[T]int)
	// nn 用于存储交集结果。
	nn := make([]T, 0)

	// 遍历第一个切片，统计每个元素的出现次数。
	for _, v := range slice1 {
		m[v]++
	}

	// 遍历第二个切片，查找共同元素。
	for _, v := range slice2 {
		// 获取当前元素在第一个切片中的出现次数。
		times, _ := m[v]
		// 如果元素在第一个切片中出现过一次，则将其添加到交集结果中，并减少其在 map 中的计数。
		if times == 1 {
			nn = append(nn, v)
			m[v]--
		}
	}

	// 返回交集结果。
	return nn
}

// Difference 计算两个切片的差集，返回一个新的切片，该切片包含在slice1中但不在slice2中的元素。
// 参数：
//
//	slice1: 第一个切片，可以是uint、int或string类型。
//	slice2: 第二个切片，与第一个切片类型相同。
//
// 返回值：
//
//	[]T: 一个新的切片，包含在slice1中但不在slice2中的元素。
func Difference[T uint | int | string](slice1, slice2 []T) []T {
	// 使用map来记录交集中每个元素出现的次数
	m := make(map[T]int)
	// 用于存储差集的切片
	nn := make([]T, 0)
	// 调用Intersect函数获取两个切片的交集
	inter := Intersect(slice1, slice2)
	// 遍历交集，增加map中每个交集元素的计数
	for _, v := range inter {
		m[v]++
	}
	// 遍历第一个切片，如果某个元素在交集中出现的次数为0，则将其添加到差集切片中
	for _, value := range slice1 {
		times, _ := m[value]
		if times == 0 {
			nn = append(nn, value)
		}
	}
	// 返回差集切片
	return nn
}
