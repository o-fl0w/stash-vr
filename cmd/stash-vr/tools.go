//go:build tools
// +build tools

//Prevent go mod tidy from removing genqlient module
//https://github.com/golang/go/issues/45552#issuecomment-819545037

package tools

import _ "github.com/Khan/genqlient"
