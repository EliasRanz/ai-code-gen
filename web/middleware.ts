import { withAuth } from 'next-auth/middleware'

export default withAuth(
  function middleware(req) {
    // Add any additional middleware logic here
    console.log('Middleware running for:', req.nextUrl.pathname)
  },
  {
    callbacks: {
      authorized: ({ token, req }) => {
        // Define which routes require authentication
        const { pathname } = req.nextUrl
        
        // Public routes that don't require authentication
        const publicRoutes = ['/login', '/register', '/error']
        const isPublicRoute = publicRoutes.some(route => pathname.startsWith(route))
        
        if (isPublicRoute) {
          return true
        }
        
        // Dashboard routes require authentication
        if (pathname.startsWith('/home') || 
            pathname.startsWith('/generate') || 
            pathname.startsWith('/projects') || 
            pathname.startsWith('/settings')) {
          return !!token
        }
        
        // Default: allow access
        return true
      },
    },
  }
)

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}
