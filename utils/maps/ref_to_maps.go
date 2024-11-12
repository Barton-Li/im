package maps

import "reflect"

// RefToMap 将结构体或指向结构体的指针转换为一个映射（map），
// 映射的键是结构体字段上的指定标签值，值是字段的值。
// 如果字段是嵌套结构体，会递归转换为嵌套的映射。
// data: 需要转换的结构体或指向结构体的指针。
// tag: 用于作为映射键的结构体字段标签。
func RefToMap(data any, tag string) map[string]any {
	// 初始化一个空的映射，用于存放转换结果。
	maps := map[string]any{}
	// 获取data的反射类型。
	t := reflect.TypeOf(data)
	// 获取data的反射值。
	v := reflect.ValueOf(data)
	// 遍历结构体的所有字段。
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 从字段的标签中获取指定名称的标签值。
		getTag, ok := field.Tag.Lookup(tag)
		// 如果字段没有指定名称的标签，跳过当前字段。
		if !ok {
			continue
		}
		val := v.Field(i)
		// 如果字段值是零值，跳过当前字段。
		if val.IsZero() {
			continue
		}
		// 如果字段类型是结构体，递归转换该字段为映射，并将其添加到结果映射中。
		if field.Type.Kind() == reflect.Struct {
			newMaps := RefToMap(val.Interface(), tag)
			maps[getTag] = newMaps
			continue
		}
		// 如果字段类型是指向结构体的指针，判断指针是否为空。
		// 如果不为空，递归转换指针指向的结构体为映射，并将其添加到结果映射中。
		if field.Type.Kind() == reflect.Ptr {
			if field.Type.Elem().Kind() == reflect.Struct {
				newMaps := RefToMap(val.Elem().Interface(), tag)
				maps[getTag] = newMaps
				continue
			}
			// 如果字段是指向非结构体类型的指针，直接将指针的值转换为接口类型，并添加到结果映射中。
			maps[getTag] = val.Elem().Interface()
			continue
		}
		// 如果字段不是结构体也不是指针，直接将字段的值转换为接口类型，并添加到结果映射中。
		maps[getTag] = val.Interface()
	}
	// 返回转换后的映射。
	return maps
}
