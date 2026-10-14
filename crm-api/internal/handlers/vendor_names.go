package handlers

import "strings"

// Only verified aliases belong here. Short or unfamiliar names must never be
// expanded by guessing which legal supplier the operator intended.
func vendorName(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	switch strings.ToLower(name) {
	case "in":
		return "", invalid("The vendor name 'in' is incomplete; enter the full vendor name")
	case "alfanar electric", "alfanar electric (riyadh)":
		return "alfanar Electric", nil
	}
	return name, nil
}
