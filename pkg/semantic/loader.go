package semantic

// pkg/semantic/loader.go

type PackageLoader interface {
	// Load returns the public scope of a package given its path and the basePath of the importer
	Load(path string, basePath string) (*Scope, error)
}
