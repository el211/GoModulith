// Package architecture validates package-level module boundaries by inspecting
// Go source imports and go.mod module paths, without runtime reflection.
//
// Convention: module roots live directly inside a configured directory, such
// as modules/orders and modules/payments. Imports of another module's internal
// packages are forbidden. AllowedDependencies can further restrict imports.
// The analyzer checks static Go imports, not reflective or dynamic interactions.
package architecture
