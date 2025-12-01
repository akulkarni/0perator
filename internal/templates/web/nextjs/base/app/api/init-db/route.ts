import { NextResponse } from 'next/server';
import { query } from '@/lib/db';

export async function POST() {
  if (!process.env.DATABASE_URL) {
    return NextResponse.json(
      { error: 'Database not configured' },
      { status: 500 }
    );
  }

  try {
    // Check if tables already exist
    const checkResult = await query(
      "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'"
    );

    if (checkResult.rows[0].count > 0) {
      return NextResponse.json({
        message: 'Database already initialized',
        tables: ['users', 'sessions', 'posts']
      });
    }

    // Schema is created by the database setup tool
    // This is just a fallback
    await query(`
      CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email VARCHAR(255) UNIQUE NOT NULL,
        name VARCHAR(255),
        created_at TIMESTAMPTZ DEFAULT NOW()
      )
    `);

    return NextResponse.json({
      message: 'Database initialized successfully',
      tables: ['users']
    });
  } catch (error) {
    console.error('Database init error:', error);
    return NextResponse.json(
      { error: 'Failed to initialize database' },
      { status: 500 }
    );
  }
}
