import { NextAuthOptions } from 'next-auth'
import CredentialsProvider from 'next-auth/providers/credentials'
import { JWT } from 'next-auth/jwt'

export const authOptions: NextAuthOptions = {
  providers: [
    CredentialsProvider({
      name: 'credentials',
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' }
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          return null
        }

        try {
          // Call our backend authentication endpoint
          const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              email: credentials.email,
              password: credentials.password,
            }),
          })

          if (response.ok) {
            const user = await response.json()
            return {
              id: user.id,
              email: user.email,
              name: user.name,
            }
          }
        } catch (error) {
          console.error('Auth error:', error)
        }

        // For development: allow any user with email/password
        if (process.env.NODE_ENV === 'development') {
          return {
            id: '1',
            email: credentials.email,
            name: credentials.email.split('@')[0],
          }
        }

        return null
      }
    })
  ],
  session: {
    strategy: 'jwt' as const,
    maxAge: 24 * 60 * 60, // 24 hours
  },
  jwt: {
    maxAge: 24 * 60 * 60, // 24 hours
  },
  callbacks: {
    async jwt({ token, user }) {
      // Initial sign in
      if (user) {
        token.userId = user.id
        // In a real implementation, these would come from the backend login response
        token.accessToken = 'access-token-' + Date.now()
        token.refreshToken = 'refresh-token-' + Date.now()
        token.tokenExpires = Date.now() + (60 * 60 * 1000) // 1 hour
      }

      // Check if token needs refresh
      if (Date.now() < (token.tokenExpires as number)) {
        return token
      }

      // Attempt to refresh the token
      return await refreshAccessToken(token)
    },
    async session({ session, token }) {
      if (token) {
        session.user.id = token.userId as string
        session.accessToken = token.accessToken as string
        session.error = token.error as string | undefined
      }
      return session
    },
  },
  pages: {
    signIn: '/login',
    signOut: '/logout',
    error: '/error',
  },
  secret: process.env.NEXTAUTH_SECRET,
}

async function refreshAccessToken(token: JWT) {
  try {
    // Call backend token refresh endpoint
    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refreshToken: token.refreshToken,
      }),
    })

    const refreshedTokens = await response.json()

    if (!response.ok) {
      throw refreshedTokens
    }

    return {
      ...token,
      accessToken: refreshedTokens.accessToken,
      tokenExpires: Date.now() + (60 * 60 * 1000), // 1 hour
      refreshToken: refreshedTokens.refreshToken ?? token.refreshToken,
    }
  } catch (error) {
    console.error('Token refresh error:', error)

    return {
      ...token,
      error: 'RefreshAccessTokenError',
    }
  }
}
