package types

type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type UpgradesPresent map[string]bool

func (u UpgradesPresent) HasUpgrade(upgrade string) bool {
	value, ok := u[upgrade]
	if !ok {
		return false
	}

	return value
}
