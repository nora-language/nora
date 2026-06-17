package semantic

// pkg/semantic/loader.go

type PackageLoader interface {
	// Load returns the public scope of a package given its path
	Load(path string) (*Scope, error)
}
