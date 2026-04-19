// Package termfolio exports embedded assets for the ronit.sh SSH portfolio.
// The //go:embed directive must live at the module root because content/ is
// a sibling directory -- embed paths cannot use "..".
package termfolio

import "embed"

// ContentFS holds all embedded markdown files from content/.
//
//go:embed content
var ContentFS embed.FS
