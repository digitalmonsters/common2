package boilerplate

import "fmt"

func GetHardDeleteUpdateQuery(idField string, field string, assignToField string, isNullable bool) string {
	if isNullable {
		return fmt.Sprintf("%v = case %v when null then null else ('deleted-' || %v || '-' || %v) end",
			assignToField, field, idField, field)
	}

	return fmt.Sprintf("%v = 'deleted-' || %v || '-' || %v", assignToField, idField, field)
}
