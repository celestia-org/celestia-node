package params

// GenesisFor reports a hash of a genesis block for a given network.
// Genesis is strictly defined and can't be modified.
// To run a custom genesis private network use CELESTIA_PRIVATE_GENESIS env var.
func GenesisFor(net Network) (string, error) {
	if err := net.Validate(); err != nil {
		return "", err
	}

	return genesisList[net], nil
}

// NOTE: Every time we add a new long-running network, its genesis hash has to be added here.
var genesisList = map[Network]string{
	DevNet:  "4632277C441CA6155C4374AC56048CF4CFE3CBB2476E07A548644435980D5E17",
	Private: "",
}
