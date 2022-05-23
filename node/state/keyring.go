package state

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/celestiaorg/celestia-app/app"
	apptypes "github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/celestiaorg/celestia-node/node/key"
	"github.com/celestiaorg/celestia-node/params"
)

func Keyring(cfg key.Config) func(keystore.Keystore, params.Network) (*apptypes.KeyringSigner, error) {
	return func(ks keystore.Keystore, net params.Network) (*apptypes.KeyringSigner, error) {
		// TODO @renaynay: Include option for setting custom `userInput` parameter with
		//  implementation of https://github.com/celestiaorg/celestia-node/issues/415.
		// TODO @renaynay @Wondertan: ensure that keyring backend from config is passed
		//  here instead of hardcoded `BackendTest`: https://github.com/celestiaorg/celestia-node/issues/603.
		ring, err := keyring.New(app.Name, keyring.BackendTest, ks.Path(), os.Stdin)
		if err != nil {
			return nil, err
		}
		signer := apptypes.NewKeyringSigner(ring, cfg.KeyringAccName, string(net))

		var info keyring.Info
		// if custom keyringAccName provided, find key for that name
		if cfg.KeyringAccName != "" {
			keyInfo, err := signer.Key(cfg.KeyringAccName)
			if err != nil {
				return nil, err
			}
			info = keyInfo
		} else {
			// check if key exists for signer
			keys, err := signer.List()
			if err != nil {
				return nil, err
			}
			// if no key was found in keystore path, generate new key for node
			if len(keys) == 0 {
				log.Infow("NO KEY FOUND IN STORE, GENERATING NEW KEY...", "path", ks.Path())
				keyInfo, mn, err := signer.NewMnemonic("my_celes_key", keyring.English, "",
					"", hd.Secp256k1)
				if err != nil {
					return nil, err
				}
				log.Info("NEW KEY GENERATED...")
				fmt.Printf("\nNAME: %s\nADDRESS: %s\nMNEMONIC (save this somewhere safe!!!): \n%s\n\n",
					keyInfo.GetName(), keyInfo.GetAddress().String(), mn)

				info = keyInfo
			} else {
				// if one or more keys are present and no keyringAccName was given, use the first key in list
				info = keys[0]
			}
		}

		log.Infow("constructed keyring signer", "backend", keyring.BackendTest, "path", ks.Path(),
			"key name", info.GetName(), "chain-id", string(net))

		return signer, nil
	}
}
