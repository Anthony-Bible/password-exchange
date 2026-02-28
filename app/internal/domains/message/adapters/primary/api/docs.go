// Package api provides REST API endpoints for Password Exchange service
//
// # Password Exchange API
//
// Secure password sharing service that allows users to submit encrypted messages
// and share them through unique, one-time access URLs. The service provides
// optional email notifications and passphrase protection for enhanced security.
//
// Terms Of Service:
//
// Version: 1.0.0
// Host: password.exchange
// BasePath: /api/v1
// Schemes: https, http
//
// Security:
// - BasicAuth
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// swagger:meta
package api

// @title Password Exchange API
// @version 1.0.0
// @description Secure password sharing service that allows users to submit encrypted messages and share them through unique, one-time access URLs.
// @description The service provides optional email notifications and passphrase protection for enhanced security.
// @description
// @description ## Email Reminder System
// @description
// @description The service includes an automated email reminder system that runs via scheduled jobs (CronJob in Kubernetes).
// @description This system automatically sends reminder emails to recipients who haven't viewed their secure messages after
// @description a configurable time period. The reminder system is not exposed through REST API endpoints but operates as
// @description a background service with configurable timing, retry logic, and resilience patterns.

// @contact.name Password Exchange Support
// @contact.url https://github.com/Anthony-Bible/password-exchange

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host password.exchange
// @BasePath /api/v1

// @schemes https http
