package chainlaunchdeploy

import (
	"errors"
)

// ValidateEVMParams checks that all required EVM deployment parameters are present and valid.
func ValidateEVMParams(params EVMParams) error {
	if params.SolidityCode == "" {
		return errors.New("SolidityCode is required")
	}
	// Optionally, validate constructor params, network, etc.
	return nil
}

// ValidateFabricParams checks that all required Fabric deployment parameters are present and valid.
func ValidateFabricParams(params FabricChaincodeDeployParams) error {
	// Validate install params
	if err := ValidateFabricChaincodeInstallParams(params.InstallParams); err != nil {
		return err
	}

	// Validate approve params
	if err := ValidateFabricChaincodeApproveParams(params.ApproveParams); err != nil {
		return err
	}

	// Validate commit params
	if err := ValidateFabricChaincodeCommitParams(params.CommitParams); err != nil {
		return err
	}

	return nil
}
