// This is a bogus package to help with dependencies woes...
package ignore

import (
	_ "github.com/google/btree"
)

// github.com/google/btree added for reasons unknown. Building locally removes this dependency,
// but building in travis adds it back in. So let's just force this import for now.
