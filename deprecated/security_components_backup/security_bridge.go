// security_bridge.go
// This file provides integration between the EntityDBServer and security components.
// It is meant to be compiled with the server_db.go file.

package main

import (
	"log"
)

// CreateEntityDBServerSecurityManager creates a security manager specifically for EntityDBServer
// This function must be used instead of NewSecurityManager for EntityDBServer
func CreateEntityDBServerSecurityManager(server *EntityDBServer) *SecurityManager {
	if server == nil {
		log.Printf("EntityDB Server: Cannot initialize security manager with nil server")
		return nil
	}

	// Create a new security manager with the server
	securityManager := &SecurityManager{
		server: server,
	}

	// Initialize input validator
	securityManager.validator = NewInputValidator()
	log.Printf("EntityDB Server: Input validator initialized")

	// Initialize audit logger
	auditLogger, err := NewAuditLogger("/opt/entitydb/var/log/audit", server.entities)
	if err != nil {
		log.Printf("EntityDB Server: Warning - Failed to initialize audit logger: %v", err)
	} else {
		securityManager.auditLogger = auditLogger
		log.Printf("EntityDB Server: Audit logger initialized")
	}

	return securityManager
}

// CreateEntityDBServerSecureMiddleware creates a secure middleware that works with the EntityDBServer
func CreateEntityDBServerSecureMiddleware(securityManager *SecurityManager, server *EntityDBServer) *SecureMiddleware {
	return &SecureMiddleware{
		sm:     securityManager,
		server: server,
	}
}