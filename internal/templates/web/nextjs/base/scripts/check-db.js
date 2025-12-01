const { Pool } = require('pg');
const fs = require('fs');
const path = require('path');

// Disable SSL certificate validation for Tiger Cloud
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

// Load .env.local file
const envPath = path.join(__dirname, '..', '.env.local');
if (fs.existsSync(envPath)) {
  const envContent = fs.readFileSync(envPath, 'utf8');
  envContent.split('\n').forEach(line => {
    const match = line.match(/^([^#=]+)=(.*)$/);
    if (match) {
      const key = match[1].trim();
      const value = match[2].trim();
      if (!process.env[key]) {
        process.env[key] = value;
      }
    }
  });
}

async function checkDatabase() {
  if (!process.env.DATABASE_URL) {
    console.log('⚠️  DATABASE_URL not configured in .env.local');
    console.log('   Run "setup_database" to create a PostgreSQL database');
    return;
  }

  const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
    connectionTimeoutMillis: 5000,
  });

  try {
    await pool.query('SELECT 1');
    console.log('✅ Database connected');

    // Check if tables exist
    const result = await pool.query(
      "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'"
    );

    if (result.rows[0].count === '0') {
      console.log('⚠️  No tables found. Run "npm run db:init" to create tables');
    }
  } catch (error) {
    console.log('❌ Database connection failed:', error.message);
    console.log('   Check your DATABASE_URL in .env.local');
  } finally {
    await pool.end();
  }
}

checkDatabase();
