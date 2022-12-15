package types

import "fmt"

// region Permission

type Permission struct {
	Name string
}

// endregion Permission

// region RequestedPermission

type RequestedPermission struct {
	Permission
}

func (r RequestedPermission) String() string {
	return fmt.Sprintf("RequestedPermission{Nane:%s}", r.Name)
}

// endregion RequestedPermission

// region PackagePermission

type PackagePermission struct {
	Permission
	Granted bool
	Flags   []string
}

func (r PackagePermission) String() string {
	return fmt.Sprintf("PackagePermission{Name:%s, Granted:%t, Flags:%s}", r.Name, r.Granted, r.Flags)
}

// endregion PackagePermission
