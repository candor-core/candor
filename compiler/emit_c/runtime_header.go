package emit_c

import _ "embed"

//go:embed _cnd_runtime.h
var runtimeHeaderContent string

// RuntimeHeader returns the content of _cnd_runtime.h, which is written
// alongside generated .c files so that Candor-level emitters (emit_c.cnd)
// can #include it when compiling their own output (Stage 2).
func RuntimeHeader() string {
	return runtimeHeaderContent
}
