package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddBrutalistUI adds brutalist/minimalist UI components and design system
func AddBrutalistUI(ctx context.Context, args map[string]string) error {
	componentType := args["component"]
	if componentType == "" {
		componentType = "all"
	}

	fmt.Println("🏗️  Adding Brutalist UI design system...")

	// Create directories
	dirs := []string{
		"components/brutalist",
		"lib/brutalist",
		"styles",
	}

	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// Create the design system configuration
	createDesignSystem()

	// Create components based on request
	switch componentType {
	case "auth":
		createAuthComponents()
	case "forms":
		createFormComponents()
	case "layout":
		createLayoutComponents()
	case "feedback":
		createFeedbackComponents()
	case "all":
		createAuthComponents()
		createFormComponents()
		createLayoutComponents()
		createFeedbackComponents()
		createDesignSystemDoc()
	default:
		createCustomComponent(componentType)
	}

	fmt.Println("✅ Brutalist UI components added!")
	fmt.Println("\nDesign principles applied:")
	fmt.Println("  - Monospace font throughout")
	fmt.Println("  - #ff4500 for all interactive elements")
	fmt.Println("  - No external CSS frameworks")
	fmt.Println("  - Inline styles only")
	fmt.Println("  - Maximum 4 colors total")
	fmt.Println("\nComponents created in: components/brutalist/")

	return nil
}

func createDesignSystem() {
	// Create the core design system
	designSystemContent := `// Brutalist Design System
// No CSS frameworks, inline styles only, monospace everything

export const brutal = {
  // Core spacing units - only these three
  spacing: {
    small: '0.5rem',
    medium: '1rem',
    large: '2rem'
  },

  // Complete color palette - only 4 colors
  colors: {
    text: 'inherit',        // Browser default (usually black)
    background: 'inherit',  // Browser default (usually white)
    action: '#ff4500',      // Orange-red for all interactive elements
    muted: '#f0f0f0'       // Light gray for feedback/messages only
  },

  // Typography - one font for everything
  typography: {
    fontFamily: 'monospace',
    fontSize: 'inherit',
    lineHeight: 'inherit'
  },

  // Component styles
  styles: {
    // Main container for any page
    container: {
      padding: '2rem',
      fontFamily: 'monospace'
    },

    // Standard form layout
    form: {
      display: 'flex',
      flexDirection: 'column',
      gap: '1rem',
      maxWidth: '300px'
    },

    // All input fields
    input: {
      padding: '0.5rem',
      fontFamily: 'monospace'
    },

    // Primary action buttons
    button: {
      padding: '0.5rem',
      cursor: 'pointer',
      fontFamily: 'monospace'
    },

    // Text-style link buttons
    linkButton: {
      background: 'none',
      border: 'none',
      color: '#ff4500',
      cursor: 'pointer',
      textDecoration: 'underline',
      padding: 0,
      font: 'inherit'
    },

    // Feedback/message boxes
    messageBox: {
      marginTop: '1rem',
      padding: '1rem',
      background: '#f0f0f0',
      borderRadius: '4px',
      fontFamily: 'monospace'
    },

    // Error states - just red text
    error: {
      color: 'red'
    },

    // Success states - just the text
    success: {
      color: 'green'
    },

    // Horizontal button group
    buttonGroup: {
      display: 'flex',
      gap: '0.5rem'
    },

    // Stacked content sections
    stack: {
      display: 'flex',
      flexDirection: 'column',
      gap: '1rem'
    }
  }
};

// Utility function to combine styles
export const combine = (...styles) => Object.assign({}, ...styles);

// Pre-made component bases
export const components = {
  page: brutal.styles.container,
  form: brutal.styles.form,
  input: brutal.styles.input,
  button: brutal.styles.button,
  link: brutal.styles.linkButton,
  message: brutal.styles.messageBox
};
`
	os.WriteFile(filepath.Join("lib", "brutalist", "design.js"), []byte(designSystemContent), 0644)
}

func createAuthComponents() {
	// Login component
	loginContent := `'use client';

import { useState } from 'react';
import { brutal } from '@/lib/brutalist/design';

export default function BrutalistLogin({ onSubmit, onSignUpClick }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await onSubmit({ email, password });
    } catch (err) {
      setError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main style={brutal.styles.container}>
      <h1>Login</h1>

      <form onSubmit={handleSubmit} style={brutal.styles.form}>
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          disabled={loading}
          style={brutal.styles.input}
        />

        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          disabled={loading}
          style={brutal.styles.input}
        />

        <button
          type="submit"
          disabled={loading}
          style={brutal.styles.button}
        >
          {loading ? 'Loading...' : 'Login'}
        </button>
      </form>

      {error && (
        <div style={brutal.styles.messageBox}>
          Error: {error}
        </div>
      )}

      <p>
        Don't have an account?{' '}
        <button
          onClick={onSignUpClick}
          style={brutal.styles.linkButton}
        >
          Sign Up
        </button>
      </p>
    </main>
  );
}
`
	os.WriteFile(filepath.Join("components", "brutalist", "Login.jsx"), []byte(loginContent), 0644)

	// Register component
	registerContent := `'use client';

import { useState } from 'react';
import { brutal } from '@/lib/brutalist/design';

export default function BrutalistRegister({ onSubmit, onLoginClick }) {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    name: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (formData.password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setLoading(true);
    try {
      await onSubmit({
        email: formData.email,
        password: formData.password,
        name: formData.name
      });
    } catch (err) {
      setError(err.message || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const updateField = (field, value) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  return (
    <main style={brutal.styles.container}>
      <h1>Create Account</h1>

      <form onSubmit={handleSubmit} style={brutal.styles.form}>
        <input
          type="text"
          placeholder="Name (optional)"
          value={formData.name}
          onChange={(e) => updateField('name', e.target.value)}
          disabled={loading}
          style={brutal.styles.input}
        />

        <input
          type="email"
          placeholder="Email"
          value={formData.email}
          onChange={(e) => updateField('email', e.target.value)}
          required
          disabled={loading}
          style={brutal.styles.input}
        />

        <input
          type="password"
          placeholder="Password (min 8 characters)"
          value={formData.password}
          onChange={(e) => updateField('password', e.target.value)}
          required
          disabled={loading}
          style={brutal.styles.input}
        />

        <input
          type="password"
          placeholder="Confirm Password"
          value={formData.confirmPassword}
          onChange={(e) => updateField('confirmPassword', e.target.value)}
          required
          disabled={loading}
          style={brutal.styles.input}
        />

        <button
          type="submit"
          disabled={loading}
          style={brutal.styles.button}
        >
          {loading ? 'Creating...' : 'Create Account'}
        </button>
      </form>

      {error && (
        <div style={brutal.styles.messageBox}>
          Error: {error}
        </div>
      )}

      <p>
        Already have an account?{' '}
        <button
          onClick={onLoginClick}
          style={brutal.styles.linkButton}
        >
          Login
        </button>
      </p>
    </main>
  );
}
`
	os.WriteFile(filepath.Join("components", "brutalist", "Register.jsx"), []byte(registerContent), 0644)
}

func createFormComponents() {
	// Generic form builder
	formBuilderContent := `'use client';

import { useState } from 'react';
import { brutal } from '@/lib/brutalist/design';

export function BrutalistForm({ fields, onSubmit, submitText = 'Submit' }) {
  const [values, setValues] = useState({});
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setErrors({});
    setLoading(true);

    try {
      await onSubmit(values);
    } catch (err) {
      if (err.fieldErrors) {
        setErrors(err.fieldErrors);
      } else {
        setErrors({ _form: err.message });
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={brutal.styles.form}>
      {fields.map(field => (
        <div key={field.name}>
          <input
            type={field.type || 'text'}
            name={field.name}
            placeholder={field.placeholder}
            value={values[field.name] || ''}
            onChange={(e) => setValues({
              ...values,
              [field.name]: e.target.value
            })}
            required={field.required}
            disabled={loading}
            style={brutal.styles.input}
          />
          {errors[field.name] && (
            <div style={{ color: 'red', fontSize: '0.9em', marginTop: '0.25rem' }}>
              {errors[field.name]}
            </div>
          )}
        </div>
      ))}

      <button type="submit" disabled={loading} style={brutal.styles.button}>
        {loading ? 'Loading...' : submitText}
      </button>

      {errors._form && (
        <div style={brutal.styles.messageBox}>
          Error: {errors._form}
        </div>
      )}
    </form>
  );
}

export function BrutalistInput({ label, error, ...props }) {
  return (
    <div>
      {label && <label style={{ display: 'block', marginBottom: '0.25rem' }}>{label}</label>}
      <input style={brutal.styles.input} {...props} />
      {error && <div style={{ color: 'red', fontSize: '0.9em', marginTop: '0.25rem' }}>{error}</div>}
    </div>
  );
}

export function BrutalistSelect({ label, options, error, ...props }) {
  return (
    <div>
      {label && <label style={{ display: 'block', marginBottom: '0.25rem' }}>{label}</label>}
      <select style={brutal.styles.input} {...props}>
        {options.map(opt => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
      {error && <div style={{ color: 'red', fontSize: '0.9em', marginTop: '0.25rem' }}>{error}</div>}
    </div>
  );
}

export function BrutalistTextarea({ label, error, ...props }) {
  return (
    <div>
      {label && <label style={{ display: 'block', marginBottom: '0.25rem' }}>{label}</label>}
      <textarea style={{ ...brutal.styles.input, minHeight: '100px', fontFamily: 'monospace' }} {...props} />
      {error && <div style={{ color: 'red', fontSize: '0.9em', marginTop: '0.25rem' }}>{error}</div>}
    </div>
  );
}
`
	os.WriteFile(filepath.Join("components", "brutalist", "Forms.jsx"), []byte(formBuilderContent), 0644)
}

func createLayoutComponents() {
	layoutContent := `'use client';

import { brutal } from '@/lib/brutalist/design';

export function BrutalistPage({ children, title }) {
  return (
    <main style={brutal.styles.container}>
      {title && <h1>{title}</h1>}
      {children}
    </main>
  );
}

export function BrutalistSection({ children, title }) {
  return (
    <section style={{ marginTop: '2rem' }}>
      {title && <h2>{title}</h2>}
      {children}
    </section>
  );
}

export function BrutalistNav({ items, currentPath }) {
  return (
    <nav style={{ marginBottom: '2rem' }}>
      {items.map((item, index) => (
        <span key={item.path}>
          {index > 0 && ' | '}
          {currentPath === item.path ? (
            <span>{item.label}</span>
          ) : (
            <a
              href={item.path}
              style={{ color: '#ff4500' }}
            >
              {item.label}
            </a>
          )}
        </span>
      ))}
    </nav>
  );
}

export function BrutalistTable({ headers, rows }) {
  return (
    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
      <thead>
        <tr>
          {headers.map(header => (
            <th key={header} style={{
              textAlign: 'left',
              padding: '0.5rem',
              borderBottom: '1px solid #000'
            }}>
              {header}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row, i) => (
          <tr key={i}>
            {row.map((cell, j) => (
              <td key={j} style={{
                padding: '0.5rem',
                borderBottom: '1px solid #ccc'
              }}>
                {cell}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function BrutalistCard({ children }) {
  return (
    <div style={{
      border: '1px solid #000',
      padding: '1rem',
      marginBottom: '1rem'
    }}>
      {children}
    </div>
  );
}

export function BrutalistModal({ isOpen, onClose, children }) {
  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      background: 'rgba(255, 255, 255, 0.95)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: 'monospace'
    }}>
      <div style={{
        background: 'white',
        border: '2px solid #000',
        padding: '2rem',
        maxWidth: '500px',
        width: '90%'
      }}>
        <button
          onClick={onClose}
          style={{
            float: 'right',
            background: 'none',
            border: 'none',
            fontSize: '1.5rem',
            cursor: 'pointer',
            padding: 0,
            marginTop: '-1rem',
            marginRight: '-1rem'
          }}
        >
          ×
        </button>
        {children}
      </div>
    </div>
  );
}
`
	os.WriteFile(filepath.Join("components", "brutalist", "Layout.jsx"), []byte(layoutContent), 0644)
}

func createFeedbackComponents() {
	feedbackContent := `'use client';

import { brutal } from '@/lib/brutalist/design';

export function BrutalistMessage({ type = 'info', children }) {
  const backgrounds = {
    info: '#f0f0f0',
    error: '#ffeeee',
    success: '#eeffee',
    warning: '#ffffee'
  };

  const prefixes = {
    info: 'Info:',
    error: 'Error:',
    success: 'Success:',
    warning: 'Warning:'
  };

  return (
    <div style={{
      ...brutal.styles.messageBox,
      background: backgrounds[type] || backgrounds.info
    }}>
      {prefixes[type]} {children}
    </div>
  );
}

export function BrutalistAlert({ children, onDismiss }) {
  return (
    <div style={{
      ...brutal.styles.messageBox,
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center'
    }}>
      <span>{children}</span>
      {onDismiss && (
        <button
          onClick={onDismiss}
          style={{
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            fontSize: '1.2rem',
            padding: '0 0.5rem'
          }}
        >
          ×
        </button>
      )}
    </div>
  );
}

export function BrutalistLoading({ message = 'Loading...' }) {
  return (
    <div style={{ padding: '2rem', fontFamily: 'monospace' }}>
      {message}
    </div>
  );
}

export function BrutalistProgress({ value, max = 100 }) {
  const percentage = Math.min(100, Math.max(0, (value / max) * 100));

  return (
    <div style={{ fontFamily: 'monospace' }}>
      <div style={{
        border: '1px solid #000',
        height: '20px',
        position: 'relative'
      }}>
        <div style={{
          background: '#ff4500',
          height: '100%',
          width: percentage + '%'
        }} />
      </div>
      <div style={{ marginTop: '0.5rem' }}>
        {value}/{max} ({Math.round(percentage)}%)
      </div>
    </div>
  );
}

export function BrutalistTooltip({ children, text }) {
  return (
    <span style={{ position: 'relative', borderBottom: '1px dotted #000' }}>
      {children}
      <span style={{
        position: 'absolute',
        bottom: '100%',
        left: '50%',
        transform: 'translateX(-50%)',
        background: '#000',
        color: '#fff',
        padding: '0.25rem 0.5rem',
        whiteSpace: 'nowrap',
        fontSize: '0.9em',
        display: 'none'
      }}>
        {text}
      </span>
    </span>
  );
}

export function BrutalistBadge({ children, color = '#ff4500' }) {
  return (
    <span style={{
      display: 'inline-block',
      padding: '0.125rem 0.5rem',
      background: color,
      color: 'white',
      fontSize: '0.875em'
    }}>
      {children}
    </span>
  );
}
`
	os.WriteFile(filepath.Join("components", "brutalist", "Feedback.jsx"), []byte(feedbackContent), 0644)
}

func createDesignSystemDoc() {
	docContent := `# Brutalist UI Components

## Design Philosophy

This design system implements a brutalist/minimalist aesthetic with:
- **No CSS frameworks** - Pure inline styles only
- **Monospace everything** - Single font throughout
- **4 colors maximum** - Black, white, #ff4500 (orange-red), #f0f0f0 (gray)
- **No decorations** - No shadows, gradients, or animations
- **Functional focus** - Every element has a purpose

## Usage

Import the design system and components:

` + "```javascript" + `
import { brutal } from '@/lib/brutalist/design';
import { BrutalistLogin } from '@/components/brutalist/Login';
import { BrutalistForm } from '@/components/brutalist/Forms';
import { BrutalistPage } from '@/components/brutalist/Layout';
import { BrutalistMessage } from '@/components/brutalist/Feedback';
` + "```" + `

## Components

### Authentication
- ` + "`" + `<BrutalistLogin />` + "`" + ` - Login form
- ` + "`" + `<BrutalistRegister />` + "`" + ` - Registration form

### Forms
- ` + "`" + `<BrutalistForm />` + "`" + ` - Dynamic form builder
- ` + "`" + `<BrutalistInput />` + "`" + ` - Styled input field
- ` + "`" + `<BrutalistSelect />` + "`" + ` - Dropdown select
- ` + "`" + `<BrutalistTextarea />` + "`" + ` - Text area

### Layout
- ` + "`" + `<BrutalistPage />` + "`" + ` - Main page container
- ` + "`" + `<BrutalistSection />` + "`" + ` - Content section
- ` + "`" + `<BrutalistNav />` + "`" + ` - Navigation bar
- ` + "`" + `<BrutalistTable />` + "`" + ` - Data table
- ` + "`" + `<BrutalistCard />` + "`" + ` - Content card
- ` + "`" + `<BrutalistModal />` + "`" + ` - Modal dialog

### Feedback
- ` + "`" + `<BrutalistMessage />` + "`" + ` - Info/error/success messages
- ` + "`" + `<BrutalistAlert />` + "`" + ` - Dismissible alerts
- ` + "`" + `<BrutalistLoading />` + "`" + ` - Loading indicator
- ` + "`" + `<BrutalistProgress />` + "`" + ` - Progress bar
- ` + "`" + `<BrutalistBadge />` + "`" + ` - Status badges

## Design Tokens

Use the ` + "`" + `brutal` + "`" + ` object for consistent styling:

` + "```" + `javascript
// Spacing
brutal.spacing.small   // 0.5rem
brutal.spacing.medium  // 1rem
brutal.spacing.large   // 2rem

// Colors
brutal.colors.text       // inherit (black)
brutal.colors.background // inherit (white)
brutal.colors.action     // #ff4500 (orange-red)
brutal.colors.muted      // #f0f0f0 (light gray)

// Pre-made styles
brutal.styles.container   // Page container
brutal.styles.form        // Form layout
brutal.styles.input       // Input field
brutal.styles.button      // Action button
brutal.styles.linkButton  // Link-style button
brutal.styles.messageBox  // Message/feedback box
` + "```" + `

## Example Implementation

` + "```" + `javascript
'use client';

import { brutal } from '@/lib/brutalist/design';
import { BrutalistPage, BrutalistSection } from '@/components/brutalist/Layout';
import { BrutalistForm } from '@/components/brutalist/Forms';
import { BrutalistMessage } from '@/components/brutalist/Feedback';

export default function ContactPage() {
  const handleSubmit = async (values) => {
    // Handle form submission
    console.log('Form submitted:', values);
  };

  return (
    <BrutalistPage title="Contact">
      <BrutalistSection>
        <p>Fill out the form below to get in touch.</p>

        <BrutalistForm
          fields={[
            { name: 'name', placeholder: 'Your Name', required: true },
            { name: 'email', type: 'email', placeholder: 'Email', required: true },
            { name: 'message', placeholder: 'Message', required: true }
          ]}
          onSubmit={handleSubmit}
          submitText="Send Message"
        />

        <BrutalistMessage type="info">
          We'll respond within 24 hours.
        </BrutalistMessage>
      </BrutalistSection>
    </BrutalistPage>
  );
}
` + "```" + `

## Philosophy

This aesthetic is inspired by:
- Early 1990s web forms
- Terminal UIs in HTML
- Brutalist architecture
- The "View Source" era of web development

The goal is interfaces that are:
1. **Immediately understandable** - No learning curve
2. **Completely transparent** - Users see exactly what's happening
3. **Easily modifiable** - No design system to fight
4. **Universally accessible** - Works everywhere
5. **Honestly minimal** - Actually minimal, not "minimal design"

## Customization

While the system is intentionally rigid, you can:
- Adjust the action color by changing ` + "`" + `#ff4500` + "`" + `
- Modify spacing units in ` + "`" + `brutal.spacing` + "`" + `
- Extend components while maintaining the aesthetic

Remember: This is not "ugly on purpose" but "honest about what it is" - functional interfaces with zero pretense.
`
	os.WriteFile(filepath.Join("components", "brutalist", "README.md"), []byte(docContent), 0644)
}

func createCustomComponent(name string) {
	// Generate a custom component based on the name
	componentName := strings.Title(strings.ToLower(name))

	customContent := fmt.Sprintf(`'use client';

import { brutal } from '@/lib/brutalist/design';

export default function Brutalist%s({ children, ...props }) {
  return (
    <div style={brutal.styles.container} {...props}>
      <h1>%s</h1>
      {children}
    </div>
  );
}
`, componentName, componentName)

	filename := fmt.Sprintf("Brutalist%s.jsx", componentName)
	os.WriteFile(filepath.Join("components", "brutalist", filename), []byte(customContent), 0644)

	fmt.Printf("Created custom component: %s\n", filename)
}