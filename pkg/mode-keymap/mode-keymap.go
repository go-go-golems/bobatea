package mode_keymap

import (
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

type ModeKeyMap struct {
	Accept key.Binding `keymap-mode:"browse,new"`
	Help   key.Binding `keymap-mode:"*"`
	Exit   key.Binding `keymap-mode:"browse,new"`
	Cancel key.Binding
}

type Modes map[string]interface{}

func NewModes(modes []string) Modes {
	ret := make(Modes)
	for _, mode := range modes {
		ret[mode] = struct{}{}
	}

	return ret
}

// Contains checks if the given mode exists in the Modes map.
// If the "*" mode exists, it will match any mode.
// Otherwise it checks if the specific mode exists.
func (m Modes) Contains(mode string) bool {
	if _, ok := m["*"]; ok {
		return true
	}
	_, ok := m[mode]
	return ok
}

// forEachKeyBinding recursively iterates through the fields of the given keymap
// struct, calling the provided function f for any key.Binding field it finds.
// It handles both normal struct fields and pointer fields containing a key.Binding.
// The modes parameter tracks which "keymap-mode" tags have been seen, to pass
// to the callback function f.
//
// Nested structs are followed recursively, be it as pointer to structs or a
// struct itself.
func forEachKeyBinding(
	keymap interface{},
	f func(b *key.Binding, modes Modes),
	modes Modes,
) {
	if keymap == nil {
		return
	}
	// check that keymap is a struct or a pointer to a struct
	if reflect.TypeOf(keymap).Kind() == reflect.Ptr {
		if reflect.TypeOf(keymap).Elem().Kind() != reflect.Struct {
			return
		}
	} else if reflect.TypeOf(keymap).Kind() != reflect.Struct {
		return
	}

	v := reflect.ValueOf(keymap)
	if v.IsNil() {
		return
	}

	// if v is a pointer, get the value it points to
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	n := v.NumField()
	for i := 0; i < n; i++ {
		field := v.Field(i)
		kind := field.Kind()

		modes_ := modes
		modeTag := v.Type().Field(i).Tag.Get("keymap-mode")
		if modeTag != "" {
			modes_ = NewModes(strings.Split(modeTag, ","))
		}

		//exhaustive:ignore
		switch kind {
		case reflect.Struct:
			type_ := field.Type()
			name := type_.Name()
			pkgPath := type_.PkgPath()
			if name == "Binding" && pkgPath == "github.com/charmbracelet/bubbles/key" {
				if addr, ok := field.Addr().Interface().(*key.Binding); ok {
					f(addr, modes_)
				}
				continue
			}

			// recurse into the struct
			forEachKeyBinding(field.Addr().Interface(), f, modes_)

		case reflect.Ptr:
			name := field.Type().Elem().Name()
			pkg := field.Type().Elem().PkgPath()
			if name == "Binding" && pkg == "github.com/charmbracelet/bubbles/key" {
				// get the modes
				if addr, ok := field.Interface().(*key.Binding); ok {
					f(addr, modes_)
				}
				continue
			}

			// recurse into the struct
			forEachKeyBinding(field.Interface(), f, modes_)

		default:
		}
	}
}

// ForEachKeyBinding calls the given function f on every key.Binding in the given
// keymap, passing the binding and its associated modes. It recurses into nested
// structs and pointers, extracting all key.Bindings.
func ForEachKeyBinding(keymap interface{}, f func(b *key.Binding, modes Modes)) {
	// panic if keymap is not an addressable value
	if reflect.TypeOf(keymap).Kind() != reflect.Ptr {
		panic("keymap must be a pointer to a struct")
	}

	forEachKeyBinding(keymap, func(b *key.Binding, modes Modes) {
		f(b, modes)
	}, NewModes([]string{"*"}))
}

// EnableMode enables all key bindings in the given keymap that are tagged
// with the provided mode.
func EnableMode(keymap interface{}, mode string) {
	ForEachKeyBinding(keymap, func(b *key.Binding, modes Modes) {
		if modes.Contains(mode) {
			b.SetEnabled(true)
		}
	})
}
