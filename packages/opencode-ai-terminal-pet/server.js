// The package entry point. Hand-written on purpose: it is the one file that has
// to combine a generated reporter with the install step, and generation should
// not own packaging decisions.
//
// Order matters. The reporter's first delivery calls
// protocol.launchApplication(), which starts <data dir>/AI-Terminal-Pet.exe and
// gives up silently when it is not there - so the panel has to be in place
// before the plugin subscribes to anything.
//
// installPanel() never throws and never awaits: it is a few hundred bytes of
// file copy on a machine where the panel is missing, and a no-op afterwards.
// Nothing here may delay or break the host (AGENTS rule 1).

import { installPanel } from './bootstrap.mjs'
import { plugin } from './reporter.js'

installPanel()

export { plugin }
export default plugin
