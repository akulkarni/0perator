/** @type {import('next').NextConfig} */

// Disable SSL certificate validation for Tiger Cloud (self-signed certs)
// This must be set before any database connections are made
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const nextConfig = {}

module.exports = nextConfig
