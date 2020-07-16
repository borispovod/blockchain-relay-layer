package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	vmExport "github.com/dfinance/dnode/x/vm/export"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Read validators and counters
	PermReader perms.Permission = ModuleName + "PermReader"
	// Add/update validators
	PermWriter perms.Permission = ModuleName + "PermWriter"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermReader, PermWriter}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestVMStoragePerms returns module perms used by this module.
func RequestVMStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			vmExport.PermStorageReader,
			vmExport.PermStorageWriter,
		}
		return
	}
}
