# AI UI Generator - Frontend

Next.js frontend application for the AI UI Generator system.

## Structure

This is a Next.js 14 application using the App Router architecture.

### Route Structure

```
/                           → Redirects to /dashboard
/dashboard                  → Main dashboard with overview
/dashboard/projects         → Project management page
/dashboard/settings         → User settings and preferences
/dashboard/generate         → AI generation interface
/auth                      → Login page
/auth/register             → Registration page
```

### Key Files

- `app/layout.tsx` - Root layout with global styles
- `app/page.tsx` - Home page (redirects to dashboard)
- `app/(dashboard)/layout.tsx` - Dashboard layout with navigation
- `app/(dashboard)/page.tsx` - Dashboard home page
- `app/(auth)/layout.tsx` - Authentication layout
- `app/(auth)/page.tsx` - Login page
- `app/globals.css` - Global CSS with Tailwind setup

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## Features

- 📱 Responsive design with Tailwind CSS
- 🎨 Modern UI with consistent styling
- 🔐 Authentication flow scaffolding
- 📁 Project management interface
- ⚙️ Settings configuration
- 🤖 AI generation interface placeholder

## Next Steps

- Implement actual authentication logic
- Connect to backend APIs
- Add real UI generation functionality
- Implement project management features
- Add user settings persistence
