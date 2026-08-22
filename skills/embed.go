// Package skill embeds the ANAL skill files so `anal install` can write them.
package skill

import "embed"

// FS holds the ANAL skill, rooted at the "anal" directory.
//
//go:embed anal
var FS embed.FS
