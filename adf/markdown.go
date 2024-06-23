package adf

import "fmt"

func ToMarkdown(v any) string {
	return fmt.Sprintf("%+v", v)
}
