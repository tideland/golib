// Tideland Go Library - Etc
//
// Copyright (C) 2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library etc configuration package provides the reading,
// parsing, and accessing of configuration data. Different readers
// can be passed as sources for the SML formatted input.
//
//     {etc
//         {global
//             {base-directory /var/lib/myserver}
//             {host-address localhost:1234}
//             {max-users 50}
//         }
//         {service-a
//             {url http://[global/host-address]/service-a}
//             {directory [global/base-directory||.]/service-a}
//         }
//     }
//
// After reading this from a file, reader, or string the number of users
// can be retrieved with a default value of 10 by calling
//
//     maxUsers := cfg.ValueAsInt("global/max-users", 10)
//
// The leading "etc" node of the path is set by default.
//
// If values contain templates formatted [path||default] the configuration
// tries to read the value out of the given path and replace the template.
// The default value is optional. It will be used, if the path cannot
// be found. If the path is invalid and has no default the template will
// stay inside the value. So accessing the directory of service-a by
//
//     svcDir := cfg.ValueAsString("service-a/directory", ".")
//
// leads to "/var/lib/myserver/service-a" and if the base directory
// isn't set to "./service-a". If nothing is set the default value is ".".
package etc

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// Version returns the version of the SML package.
func Version() version.Version {
	return version.New(1, 6, 0)
}

// EOF
