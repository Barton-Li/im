package maps

import "reflect"

// MapToStruct 将一个 map 字段的值映射到一个结构体的对应字段中。
// data: 包含键值对的数据源，其键对应于结构体字段的 JSON 标签。
// dst: 需要填充的结构体实例。
func MapToStruct(data map[string]any, dst any) {
	// 获取结构体类型
	t := reflect.TypeOf(dst).Elem()
	// 获取结构体值
	v := reflect.ValueOf(dst).Elem()
	// 遍历结构体的所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		// 如果字段没有 JSON 标签或标记为 "-"，则跳过
		if tag == "" || tag == "-" {
			continue
		}
		// 尝试从数据源中获取对应字段的值
		mapField, ok := data[tag]
		// 如果数据源中没有对应字段的值，则跳过
		if !ok {
			continue
		}
		val := v.Field(i)
		// 如果字段类型为指针
		if field.Type.Kind() == reflect.Ptr {
			switch field.Type.Elem().Kind() {
			// 如果指针的元素类型为字符串
			case reflect.String:
				mapFieldValue := reflect.ValueOf(mapField)
				// 如果数据源中的值类型也为字符串
				if mapFieldValue.Type().Kind() == reflect.String {
					strVal := mapField.(string)
					// 设置结构体字段的值
					val.Set(reflect.ValueOf(&strVal))
				}
			}
		}
	}
}
