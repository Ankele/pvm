package render

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ankele/pvm/internal/model"
)

type Renderer struct {
	Out    io.Writer
	Format model.OutputFormat
}

func (r Renderer) Print(value any) error {
	if r.Out == nil {
		return nil
	}
	switch r.Format {
	case model.OutputJSON:
		enc := json.NewEncoder(r.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(value)
	default:
		_, err := fmt.Fprintln(r.Out, Text(value))
		return err
	}
}

func Text(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case model.ActionResult:
		return v.Message
	case *model.ActionResult:
		if v == nil {
			return ""
		}
		return v.Message
	case *model.GraphicsInfo:
		lines := []string{"VM: " + v.VM}
		for _, entry := range v.Entries {
			lines = append(lines, fmt.Sprintf("- %s listen=%s port=%d tls=%d websocket=%d", entry.Type, entry.Listen, entry.Port, entry.TLSPort, entry.Websocket))
		}
		return strings.Join(lines, "\n")
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return ""
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		return Text(rv.Elem().Interface())
	}
	if rv.Kind() == reflect.Slice {
		lines := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			lines = append(lines, Text(rv.Index(i).Interface()))
		}
		return strings.Join(lines, "\n")
	}
	if rv.Kind() == reflect.Struct {
		typ := rv.Type()
		lines := make([]string, 0, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}
			fv := rv.Field(i)
			if isZeroValue(fv) {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: %v", field.Name, fv.Interface()))
		}
		return strings.Join(lines, "\n")
	}
	return fmt.Sprintf("%v", value)
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
