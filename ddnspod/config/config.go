package config

type UciConfig struct {
	Section map[string]UciSection
}

type UciSection struct {
	Key    string
	Type   string
	List   map[string]UciList
	Option map[string]UciOption
}
type UciList = []string
type UciOption = string
