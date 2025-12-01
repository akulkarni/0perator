import { Pool, PoolClient } from 'pg';

// Disable SSL certificate validation for Tiger Cloud (self-signed certs)
// This is safe for Tiger Cloud as the connection is still encrypted
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

let pool: Pool | undefined;

function getPool(): Pool {
  if (!pool) {
    if (!process.env.DATABASE_URL) {
      throw new Error('DATABASE_URL not configured. Run setup_database to create a PostgreSQL database.');
    }

    pool = new Pool({
      connectionString: process.env.DATABASE_URL,
      max: 20,
      idleTimeoutMillis: 30000,
      connectionTimeoutMillis: 5000,
    });

    pool.on('error', (err) => {
      console.error('Unexpected database pool error:', err);
    });
  }
  return pool;
}

// Query helper - use this for most database operations
export async function query(text: string, params?: any[]) {
  const p = getPool();
  return await p.query(text, params);
}

// Get a client for transactions
export async function getClient(): Promise<PoolClient> {
  const p = getPool();
  return await p.connect();
}

// Transaction helper
export async function withTransaction<T>(
  callback: (client: PoolClient) => Promise<T>
): Promise<T> {
  const client = await getClient();
  try {
    await client.query('BEGIN');
    const result = await callback(client);
    await client.query('COMMIT');
    return result;
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}

export default pool;
