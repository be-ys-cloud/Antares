package structures

type ConfigPaths struct {
	Engine    string   `json:"engine"`
	Paths     []string `json:"paths"`
	FullPaths []string `json:"-"`
}

type MetadataList struct {
	Data MetadataKey `json:"data"`
}

type MetadataKey struct {
	Keys []string `json:"keys"`
}

type DataList struct {
	Data DataData `json:"data"`
}

type DataData struct {
	Data map[string]string `json:"data"`
}

type Configuration struct {
	Export         bool
	Import         bool
	KeePassFile    string
	PathsToExport  []ConfigPaths
	MasterPassword string
	VaultAddress   string
	VaultToken     string
}
