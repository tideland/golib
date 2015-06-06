// Tideland Go Library - Web - Options
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web

//--------------------
// IMPORTS
//--------------------

import ()

//--------------------
// OPTIONS
//--------------------

const (
	defaultBasePath        = "/"
	defaultDefaultDomain   = "default"
	defaultDefaultResource = "default"
)

// Option defines a function setting an option for a system
// like the server or the multiplexer.
type Option func(mux Multiplexer) error

// BasePath sets the path thats used as prefix before
// domain and resource.
func BasePath(basePath string) Option {
	return func(mux Multiplexer) error {
		if basePath == "" {
			basePath = defaultBasePath
		}
		if basePath[len(basePath)-1] != '/' {
			basePath += "/"
		}
		m := mux.(*multiplexer)
		m.basePath = basePath
		return nil
	}
}

// DefaultDomainResource sets the default domain and resource.
func DefaultDomainResource(defaultDomain, defaultResource string) Option {
	return func(mux Multiplexer) error {
		if defaultDomain == "" {
			defaultDomain = defaultDefaultDomain
		}
		if defaultResource == "" {
			defaultResource = defaultDefaultResource
		}
		m := mux.(*multiplexer)
		m.defaultDomain = defaultDomain
		m.defaultResource = defaultResource
		return nil
	}
}

// Scenes sets the scene manager of the multiplexer.
func Scenes(sm SceneManager) Option {
	return func(mux Multiplexer) error {
		m := mux.(*multiplexer)
		m.sceneManager = sm
		return nil
	}
}

// Handlers allows to register multiple handlers direct at creation
// of the server.
func Handlers(registrations Registrations) Option {
	return func(mux Multiplexer) error {
		return mux.RegisterAll(registrations)
	}
}

// EOF
