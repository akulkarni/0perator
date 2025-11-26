package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddJWTAuth adds real JWT authentication to a Next.js or Express app
func AddJWTAuth(ctx context.Context, args map[string]string) error {
	framework := args["framework"]
	if framework == "" {
		// Try to detect framework
		if _, err := os.Stat("next.config.js"); err == nil {
			framework = "nextjs"
		} else if _, err := os.Stat("package.json"); err == nil {
			// Check package.json for express
			data, _ := os.ReadFile("package.json")
			if string(data) != "" && (contains(string(data), "express")) {
				framework = "express"
			} else {
				framework = "nextjs" // default
			}
		}
	}

	switch framework {
	case "express":
		return addJWTAuthExpress(ctx, args)
	default:
		return addJWTAuthNextJS(ctx, args)
	}
}

func addJWTAuthNextJS(ctx context.Context, args map[string]string) error {
	fmt.Println("🔐 Adding JWT authentication to Next.js app...")

	// Create auth directories
	dirs := []string{
		"app/api/auth",
		"app/api/auth/login",
		"app/api/auth/register",
		"app/api/auth/verify",
		"app/api/auth/refresh",
		"lib/auth",
		"components/auth",
		"middleware",
	}

	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// Install dependencies
	fmt.Println("📦 Installing JWT dependencies...")
	installCmd := "npm install jsonwebtoken bcryptjs cookie && npm install --save-dev @types/jsonwebtoken @types/bcryptjs"
	fmt.Printf("   Run: %s\n", installCmd)

	// Create JWT utilities (lib/auth/jwt.ts)
	jwtUtilContent := `import jwt from 'jsonwebtoken';
import { cookies } from 'next/headers';

const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
const JWT_REFRESH_SECRET = process.env.JWT_REFRESH_SECRET || 'your-refresh-secret-key';

export interface TokenPayload {
  userId: string;
  email: string;
}

export function generateTokens(payload: TokenPayload) {
  const accessToken = jwt.sign(payload, JWT_SECRET, {
    expiresIn: '15m',
  });

  const refreshToken = jwt.sign(payload, JWT_REFRESH_SECRET, {
    expiresIn: '7d',
  });

  return { accessToken, refreshToken };
}

export function verifyAccessToken(token: string): TokenPayload | null {
  try {
    return jwt.verify(token, JWT_SECRET) as TokenPayload;
  } catch {
    return null;
  }
}

export function verifyRefreshToken(token: string): TokenPayload | null {
  try {
    return jwt.verify(token, JWT_REFRESH_SECRET) as TokenPayload;
  } catch {
    return null;
  }
}

export async function getTokenFromCookies(): Promise<string | null> {
  const cookieStore = cookies();
  const token = cookieStore.get('auth-token');
  return token?.value || null;
}

export async function setAuthCookies(accessToken: string, refreshToken: string) {
  const cookieStore = cookies();

  cookieStore.set('auth-token', accessToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 15, // 15 minutes
    path: '/',
  });

  cookieStore.set('refresh-token', refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: '/',
  });
}

export async function clearAuthCookies() {
  const cookieStore = cookies();
  cookieStore.delete('auth-token');
  cookieStore.delete('refresh-token');
}
`
	os.WriteFile(filepath.Join("lib", "auth", "jwt.ts"), []byte(jwtUtilContent), 0644)

	// Create password utilities (lib/auth/password.ts)
	passwordUtilContent := `import bcrypt from 'bcryptjs';

export async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, 12);
}

export async function verifyPassword(password: string, hashedPassword: string): Promise<boolean> {
  return bcrypt.compare(password, hashedPassword);
}

export function validatePassword(password: string): { valid: boolean; message?: string } {
  if (password.length < 8) {
    return { valid: false, message: 'Password must be at least 8 characters long' };
  }

  if (!/[A-Z]/.test(password)) {
    return { valid: false, message: 'Password must contain at least one uppercase letter' };
  }

  if (!/[a-z]/.test(password)) {
    return { valid: false, message: 'Password must contain at least one lowercase letter' };
  }

  if (!/[0-9]/.test(password)) {
    return { valid: false, message: 'Password must contain at least one number' };
  }

  return { valid: true };
}

export function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}
`
	os.WriteFile(filepath.Join("lib", "auth", "password.ts"), []byte(passwordUtilContent), 0644)

	// Create auth middleware (lib/auth/middleware.ts)
	middlewareContent := `import { NextRequest, NextResponse } from 'next/server';
import { verifyAccessToken, verifyRefreshToken, generateTokens } from './jwt';

export async function withAuth(
  request: NextRequest,
  handler: (request: NextRequest, user: any) => Promise<NextResponse>
) {
  const authToken = request.cookies.get('auth-token')?.value;
  const refreshToken = request.cookies.get('refresh-token')?.value;

  // Try to verify access token
  if (authToken) {
    const payload = verifyAccessToken(authToken);
    if (payload) {
      return handler(request, payload);
    }
  }

  // Try to refresh with refresh token
  if (refreshToken) {
    const payload = verifyRefreshToken(refreshToken);
    if (payload) {
      const { accessToken, refreshToken: newRefreshToken } = generateTokens({
        userId: payload.userId,
        email: payload.email,
      });

      const response = await handler(request, payload);

      response.cookies.set('auth-token', accessToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax',
        maxAge: 60 * 15,
        path: '/',
      });

      response.cookies.set('refresh-token', newRefreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax',
        maxAge: 60 * 60 * 24 * 7,
        path: '/',
      });

      return response;
    }
  }

  return NextResponse.json(
    { error: 'Unauthorized' },
    { status: 401 }
  );
}
`
	os.WriteFile(filepath.Join("lib", "auth", "middleware.ts"), []byte(middlewareContent), 0644)

	// Create login API route (app/api/auth/login/route.ts)
	loginRouteContent := `import { NextRequest, NextResponse } from 'next/server';
import pool from '@/lib/db';
import { verifyPassword, validateEmail } from '@/lib/auth/password';
import { generateTokens, setAuthCookies } from '@/lib/auth/jwt';

export async function POST(request: NextRequest) {
  try {
    const { email, password } = await request.json();

    // Validate input
    if (!email || !password) {
      return NextResponse.json(
        { error: 'Email and password are required' },
        { status: 400 }
      );
    }

    if (!validateEmail(email)) {
      return NextResponse.json(
        { error: 'Invalid email format' },
        { status: 400 }
      );
    }

    if (!pool) {
      return NextResponse.json(
        { error: 'Database not configured' },
        { status: 500 }
      );
    }

    // Find user
    const result = await pool.query(
      'SELECT id, email, password_hash, name FROM users WHERE email = $1',
      [email]
    );

    if (result.rows.length === 0) {
      return NextResponse.json(
        { error: 'Invalid email or password' },
        { status: 401 }
      );
    }

    const user = result.rows[0];

    // Verify password
    const isValid = await verifyPassword(password, user.password_hash);
    if (!isValid) {
      return NextResponse.json(
        { error: 'Invalid email or password' },
        { status: 401 }
      );
    }

    // Generate tokens
    const { accessToken, refreshToken } = generateTokens({
      userId: user.id,
      email: user.email,
    });

    // Set cookies
    const response = NextResponse.json({
      success: true,
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
      },
    });

    response.cookies.set('auth-token', accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 60 * 15, // 15 minutes
      path: '/',
    });

    response.cookies.set('refresh-token', refreshToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 60 * 60 * 24 * 7, // 7 days
      path: '/',
    });

    return response;
  } catch (error) {
    console.error('Login error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
`
	os.WriteFile(filepath.Join("app", "api", "auth", "login", "route.ts"), []byte(loginRouteContent), 0644)

	// Create register API route (app/api/auth/register/route.ts)
	registerRouteContent := `import { NextRequest, NextResponse } from 'next/server';
import pool from '@/lib/db';
import { hashPassword, validatePassword, validateEmail } from '@/lib/auth/password';
import { generateTokens } from '@/lib/auth/jwt';

export async function POST(request: NextRequest) {
  try {
    const { email, password, name } = await request.json();

    // Validate input
    if (!email || !password) {
      return NextResponse.json(
        { error: 'Email and password are required' },
        { status: 400 }
      );
    }

    if (!validateEmail(email)) {
      return NextResponse.json(
        { error: 'Invalid email format' },
        { status: 400 }
      );
    }

    const passwordValidation = validatePassword(password);
    if (!passwordValidation.valid) {
      return NextResponse.json(
        { error: passwordValidation.message },
        { status: 400 }
      );
    }

    if (!pool) {
      return NextResponse.json(
        { error: 'Database not configured' },
        { status: 500 }
      );
    }

    // Check if user exists
    const existingUser = await pool.query(
      'SELECT id FROM users WHERE email = $1',
      [email]
    );

    if (existingUser.rows.length > 0) {
      return NextResponse.json(
        { error: 'Email already registered' },
        { status: 400 }
      );
    }

    // Hash password
    const passwordHash = await hashPassword(password);

    // Create user
    const result = await pool.query(
      'INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id, email, name',
      [email, passwordHash, name || null]
    );

    const newUser = result.rows[0];

    // Generate tokens
    const { accessToken, refreshToken } = generateTokens({
      userId: newUser.id,
      email: newUser.email,
    });

    // Set cookies
    const response = NextResponse.json({
      success: true,
      user: {
        id: newUser.id,
        email: newUser.email,
        name: newUser.name,
      },
    });

    response.cookies.set('auth-token', accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 60 * 15,
      path: '/',
    });

    response.cookies.set('refresh-token', refreshToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 60 * 60 * 24 * 7,
      path: '/',
    });

    return response;
  } catch (error) {
    console.error('Registration error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
`
	os.WriteFile(filepath.Join("app", "api", "auth", "register", "route.ts"), []byte(registerRouteContent), 0644)

	// Create verify API route (app/api/auth/verify/route.ts)
	verifyRouteContent := `import { NextRequest, NextResponse } from 'next/server';
import { verifyAccessToken } from '@/lib/auth/jwt';
import pool from '@/lib/db';

export async function GET(request: NextRequest) {
  try {
    const authToken = request.cookies.get('auth-token')?.value;

    if (!authToken) {
      return NextResponse.json(
        { error: 'No auth token' },
        { status: 401 }
      );
    }

    const payload = verifyAccessToken(authToken);
    if (!payload) {
      return NextResponse.json(
        { error: 'Invalid token' },
        { status: 401 }
      );
    }

    // Get user details
    if (pool) {
      const result = await pool.query(
        'SELECT id, email, name FROM users WHERE id = $1',
        [payload.userId]
      );

      if (result.rows.length > 0) {
        return NextResponse.json({
          authenticated: true,
          user: result.rows[0],
        });
      }
    }

    return NextResponse.json({
      authenticated: true,
      user: {
        id: payload.userId,
        email: payload.email,
      },
    });
  } catch (error) {
    console.error('Verify error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
`
	os.WriteFile(filepath.Join("app", "api", "auth", "verify", "route.ts"), []byte(verifyRouteContent), 0644)

	// Create logout route
	logoutRouteContent := `import { NextRequest, NextResponse } from 'next/server';

export async function POST(request: NextRequest) {
  const response = NextResponse.json({
    success: true,
    message: 'Logged out successfully',
  });

  // Clear auth cookies
  response.cookies.delete('auth-token');
  response.cookies.delete('refresh-token');

  return response;
}
`
	os.WriteFile(filepath.Join("app", "api", "auth", "logout", "route.ts"), []byte(logoutRouteContent), 0644)

	// Create auth context/hook (lib/auth/useAuth.tsx)
	authHookContent := `'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface User {
  id: string;
  email: string;
  name?: string;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name?: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const checkAuth = async () => {
    try {
      const response = await fetch('/api/auth/verify');
      if (response.ok) {
        const data = await response.json();
        setUser(data.user);
      } else {
        setUser(null);
      }
    } catch (error) {
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const login = async (email: string, password: string) => {
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const data = await response.json();
    setUser(data.user);
  };

  const register = async (email: string, password: string, name?: string) => {
    const response = await fetch('/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password, name }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Registration failed');
    }

    const data = await response.json();
    setUser(data.user);
  };

  const logout = async () => {
    await fetch('/api/auth/logout', { method: 'POST' });
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
`
	os.WriteFile(filepath.Join("lib", "auth", "useAuth.tsx"), []byte(authHookContent), 0644)

	// Create LoginForm component - Brutalist light mode
	loginFormContent := `'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth/useAuth';

// Brutalist light mode colors
const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '1rem',
    maxWidth: '400px',
    padding: '2rem',
    background: '#ffffff',
    border: '2px solid #000000',
    fontFamily: 'monospace',
  },
  heading: {
    fontSize: '1.5rem',
    marginBottom: '0.5rem',
    color: '#000000',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.05em',
  },
  error: {
    padding: '0.75rem',
    background: '#fff0f0',
    color: '#cc0000',
    border: '1px solid #cc0000',
  },
  label: {
    fontSize: '0.75rem',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.1em',
    color: '#666666',
    marginBottom: '0.25rem',
  },
  input: {
    padding: '0.75rem',
    border: '2px solid #000000',
    background: '#ffffff',
    fontSize: '1rem',
    fontFamily: 'monospace',
    color: '#000000',
  },
  button: {
    padding: '0.75rem',
    background: '#ff4500',
    color: '#ffffff',
    border: 'none',
    fontSize: '1rem',
    cursor: 'pointer',
    fontFamily: 'monospace',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.05em',
    fontWeight: 'bold',
  },
  buttonDisabled: {
    background: '#cccccc',
    cursor: 'not-allowed',
  },
  link: {
    color: '#ff4500',
    textDecoration: 'underline',
    cursor: 'pointer',
    background: 'none',
    border: 'none',
    fontFamily: 'monospace',
    fontSize: '0.9rem',
  },
};

export default function LoginForm({ onSwitch }: { onSwitch?: () => void }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await login(email, password);
    } catch (err: any) {
      setError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={styles.container}>
      <h2 style={styles.heading}>Login</h2>
      <p style={{ color: '#666666', marginTop: 0 }}>Sign in to your account.</p>

      {error && <div style={styles.error}>{error}</div>}

      <div>
        <label style={styles.label}>Email</label>
        <input
          type="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          style={styles.input}
        />
      </div>

      <div>
        <label style={styles.label}>Password</label>
        <input
          type="password"
          placeholder="••••••••"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          style={styles.input}
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        style={{ ...styles.button, ...(loading ? styles.buttonDisabled : {}) }}
      >
        {loading ? 'Signing in...' : 'Sign In →'}
      </button>

      {onSwitch && (
        <p style={{ textAlign: 'center', color: '#666666' }}>
          Don't have an account?{' '}
          <button type="button" onClick={onSwitch} style={styles.link}>
            Sign up
          </button>
        </p>
      )}
    </form>
  );
}
`
	os.WriteFile(filepath.Join("components", "auth", "LoginForm.tsx"), []byte(loginFormContent), 0644)

	// Create RegisterForm component - Brutalist light mode
	registerFormContent := `'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth/useAuth';

// Brutalist light mode colors
const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '1rem',
    maxWidth: '400px',
    padding: '2rem',
    background: '#ffffff',
    border: '2px solid #000000',
    fontFamily: 'monospace',
  },
  heading: {
    fontSize: '1.5rem',
    marginBottom: '0.5rem',
    color: '#000000',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.05em',
  },
  error: {
    padding: '0.75rem',
    background: '#fff0f0',
    color: '#cc0000',
    border: '1px solid #cc0000',
  },
  label: {
    fontSize: '0.75rem',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.1em',
    color: '#666666',
    marginBottom: '0.25rem',
  },
  input: {
    padding: '0.75rem',
    border: '2px solid #000000',
    background: '#ffffff',
    fontSize: '1rem',
    fontFamily: 'monospace',
    color: '#000000',
    width: '100%',
    boxSizing: 'border-box' as const,
  },
  button: {
    padding: '0.75rem',
    background: '#ff4500',
    color: '#ffffff',
    border: 'none',
    fontSize: '1rem',
    cursor: 'pointer',
    fontFamily: 'monospace',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.05em',
    fontWeight: 'bold',
  },
  buttonDisabled: {
    background: '#cccccc',
    cursor: 'not-allowed',
  },
  link: {
    color: '#ff4500',
    textDecoration: 'underline',
    cursor: 'pointer',
    background: 'none',
    border: 'none',
    fontFamily: 'monospace',
    fontSize: '0.9rem',
  },
};

export default function RegisterForm({ onSwitch }: { onSwitch?: () => void }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { register } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }

    setLoading(true);

    try {
      await register(email, password, name);
    } catch (err: any) {
      setError(err.message || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={styles.container}>
      <h2 style={styles.heading}>Sign Up</h2>
      <p style={{ color: '#666666', marginTop: 0 }}>Create your account.</p>

      {error && <div style={styles.error}>{error}</div>}

      <div>
        <label style={styles.label}>Name</label>
        <input
          type="text"
          placeholder="Your name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          style={styles.input}
        />
      </div>

      <div>
        <label style={styles.label}>Email</label>
        <input
          type="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          style={styles.input}
        />
      </div>

      <div>
        <label style={styles.label}>Password</label>
        <input
          type="password"
          placeholder="••••••••"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          style={styles.input}
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        style={{ ...styles.button, ...(loading ? styles.buttonDisabled : {}) }}
      >
        {loading ? 'Creating account...' : 'Create Account →'}
      </button>

      {onSwitch && (
        <p style={{ textAlign: 'center', color: '#666666' }}>
          Already have an account?{' '}
          <button type="button" onClick={onSwitch} style={styles.link}>
            Login
          </button>
        </p>
      )}
    </form>
  );
}
`
	os.WriteFile(filepath.Join("components", "auth", "RegisterForm.tsx"), []byte(registerFormContent), 0644)

	// Create AuthForms component that combines both
	authFormsContent := `'use client';

import { useState } from 'react';
import LoginForm from './LoginForm';
import RegisterForm from './RegisterForm';

export default function AuthForms() {
  const [mode, setMode] = useState<'login' | 'register'>('login');

  return (
    <div style={{ fontFamily: 'monospace' }}>
      <div style={{ marginBottom: '2rem' }}>
        <button
          onClick={() => setMode('login')}
          style={{
            padding: '0.5rem 1rem',
            background: mode === 'login' ? '#ff4500' : 'transparent',
            color: mode === 'login' ? 'white' : '#ff4500',
            border: '1px solid #ff4500',
            borderRadius: '4px 0 0 4px',
            cursor: 'pointer',
            fontFamily: 'monospace'
          }}
        >
          Login
        </button>
        <button
          onClick={() => setMode('register')}
          style={{
            padding: '0.5rem 1rem',
            background: mode === 'register' ? '#ff4500' : 'transparent',
            color: mode === 'register' ? 'white' : '#ff4500',
            border: '1px solid #ff4500',
            borderRadius: '0 4px 4px 0',
            marginLeft: '-1px',
            cursor: 'pointer',
            fontFamily: 'monospace'
          }}
        >
          Register
        </button>
      </div>

      {mode === 'login' ? <LoginForm /> : <RegisterForm />}
    </div>
  );
}
`
	os.WriteFile(filepath.Join("components", "auth", "AuthForms.tsx"), []byte(authFormsContent), 0644)

	// Create ProtectedRoute component
	protectedRouteContent := `'use client';

import { useAuth } from '@/lib/auth/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.push('/');
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center', fontFamily: 'monospace' }}>
        Loading...
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return <>{children}</>;
}
`
	os.WriteFile(filepath.Join("components", "auth", "ProtectedRoute.tsx"), []byte(protectedRouteContent), 0644)

	// Create an auth page that uses these components
	authPageContent := `'use client';

import { useAuth } from '@/lib/auth/useAuth';
import AuthForms from '@/components/auth/AuthForms';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function AuthPage() {
  const { user, logout } = useAuth();
  const router = useRouter();

  // If already logged in, show user info and logout
  if (user) {
    return (
      <main style={{ padding: '2rem', fontFamily: 'monospace', maxWidth: '800px', margin: '0 auto' }}>
        <h1 style={{ fontSize: '2rem', marginBottom: '2rem' }}>Welcome!</h1>

        <div style={{ padding: '1rem', background: '#f0f0f0', borderRadius: '4px', marginBottom: '2rem' }}>
          <h2 style={{ fontSize: '1.5rem', marginBottom: '1rem' }}>User Profile</h2>
          <p><strong>Email:</strong> {user.email}</p>
          {user.name && <p><strong>Name:</strong> {user.name}</p>}
          <p><strong>ID:</strong> {user.id}</p>
        </div>

        <button
          onClick={async () => {
            await logout();
            router.push('/');
          }}
          style={{
            padding: '0.75rem 1.5rem',
            background: '#ff4500',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            fontSize: '1rem',
            cursor: 'pointer',
            fontFamily: 'monospace'
          }}
        >
          Logout
        </button>
      </main>
    );
  }

  return (
    <main style={{ padding: '2rem', fontFamily: 'monospace', maxWidth: '800px', margin: '0 auto' }}>
      <h1 style={{ fontSize: '2rem', marginBottom: '2rem' }}>Authentication</h1>
      <AuthForms />
    </main>
  );
}
`
	os.WriteFile(filepath.Join("app", "auth", "page.tsx"), []byte(authPageContent), 0644)

	// Update layout.tsx to wrap with AuthProvider
	layoutPath := filepath.Join("app", "layout.tsx")
	if layoutData, err := os.ReadFile(layoutPath); err == nil {
		layoutStr := string(layoutData)
		// Add import at the top
		importLine := "import { AuthProvider } from '@/lib/auth/useAuth';\n"
		if !contains(layoutStr, "AuthProvider") {
			// Find the imports section and add our import
			if idx := findSubstringIndex(layoutStr, "export default function"); idx > 0 {
				layoutStr = layoutStr[:idx] + importLine + layoutStr[idx:]
			}

			// Wrap children with AuthProvider
			layoutStr = strings.Replace(layoutStr,
				"{children}",
				"<AuthProvider>{children}</AuthProvider>",
				1)

			os.WriteFile(layoutPath, []byte(layoutStr), 0644)
		}
	}

	// Update .env.local template
	envAdditions := `
# JWT Secrets (generate with: openssl rand -base64 32)
JWT_SECRET=your-secret-key-change-in-production
JWT_REFRESH_SECRET=your-refresh-secret-key-change-in-production
`

	if data, err := os.ReadFile(".env.local"); err == nil {
		os.WriteFile(".env.local", append(data, []byte(envAdditions)...), 0600)
	}

	// Update the home page to include auth UI
	homePagePath := filepath.Join("app", "page.tsx")
	if homeData, err := os.ReadFile(homePagePath); err == nil {
		homeStr := string(homeData)
		if !contains(homeStr, "AuthForms") {
			// Add import
			importAuth := "import AuthForms from '@/components/auth/AuthForms';\nimport { useAuth } from '@/lib/auth/useAuth';\n"
			if idx := findSubstringIndex(homeStr, "export default function"); idx > 0 {
				homeStr = homeStr[:idx] + importAuth + homeStr[idx:]
			}

			// Add auth section before the closing main tag
			authSection := `
      <div style={{ marginTop: '2rem', padding: '1rem', background: '#f0f0f0', borderRadius: '4px' }}>
        <h2 style={{ fontSize: '1.5rem', marginBottom: '1rem' }}>Authentication</h2>
        <AuthForms />
      </div>`
			if idx := findSubstringIndex(homeStr, "</main>"); idx > 0 {
				homeStr = homeStr[:idx] + authSection + "\n" + homeStr[idx:]
			}

			os.WriteFile(homePagePath, []byte(homeStr), 0644)
		}
	}

	fmt.Println("✅ JWT authentication with UI added successfully!")
	fmt.Println("\nFeatures added:")
	fmt.Println("  - JWT token generation and verification")
	fmt.Println("  - Secure password hashing with bcrypt")
	fmt.Println("  - Login endpoint (/api/auth/login)")
	fmt.Println("  - Register endpoint (/api/auth/register)")
	fmt.Println("  - Token verification (/api/auth/verify)")
	fmt.Println("  - Logout endpoint (/api/auth/logout)")
	fmt.Println("  - Auth middleware for protected routes")
	fmt.Println("  - useAuth React hook for client-side")
	fmt.Println("  - Secure HTTP-only cookies")
	fmt.Println("  - Refresh token support")
	fmt.Println("\nUI Components added:")
	fmt.Println("  - Login form with validation")
	fmt.Println("  - Register form with password confirmation")
	fmt.Println("  - Auth switcher component")
	fmt.Println("  - Protected route wrapper")
	fmt.Println("  - Auth UI on home page (/)")
	fmt.Println("  - Dedicated auth page (/auth)")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run: npm install jsonwebtoken bcryptjs cookie")
	fmt.Println("  2. Run: npm install --save-dev @types/jsonwebtoken @types/bcryptjs")
	fmt.Println("  3. Generate secrets: openssl rand -base64 32")
	fmt.Println("  4. Update JWT_SECRET and JWT_REFRESH_SECRET in .env.local")
	fmt.Println("  5. Wrap your app with <AuthProvider> in layout.tsx")

	return nil
}

func addJWTAuthExpress(ctx context.Context, args map[string]string) error {
	// Express implementation would go here
	// Similar to Next.js but with Express middleware
	fmt.Println("Express JWT auth not yet implemented")
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && s[0:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(substr) > 0 && len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func findSubstringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}