import { getServerSession } from 'next-auth'
import { getSession as getClientSession } from 'next-auth/react'
import { authOptions } from './auth'

// Server-side session helper
export async function getServerAuthSession() {
  return await getServerSession(authOptions)
}

// Client-side session helper
export async function getClientAuthSession() {
  return await getClientSession()
}

// JWT token utilities (stubs for now)
export const tokenUtils = {
  // Validate JWT token format (stub)
  validateToken: (token: string): boolean => {
    if (!token) return false
    
    // Basic format check for JWT (header.payload.signature)
    const parts = token.split('.')
    return parts.length === 3
  },

  // Extract token expiration (stub)
  getTokenExpiry: (token: string): number | null => {
    try {
      if (!tokenUtils.validateToken(token)) return null
      
      const payload = JSON.parse(atob(token.split('.')[1]))
      return payload.exp ? payload.exp * 1000 : null
    } catch {
      return null
    }
  },

  // Check if token is expired (stub)
  isTokenExpired: (token: string): boolean => {
    const expiry = tokenUtils.getTokenExpiry(token)
    if (!expiry) return true
    
    return Date.now() >= expiry
  },

  // Refresh token logic
  refreshToken: async (refreshToken: string): Promise<{ accessToken: string; refreshToken: string } | null> => {
    try {
      // Call backend refresh endpoint
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refreshToken }),
      })

      if (response.ok) {
        return await response.json()
      }
    } catch (error) {
      console.error('Token refresh failed:', error)
    }
    
    return null
  }
}

// Auth state management utilities
export const authStateUtils = {
  // Local storage keys
  ACCESS_TOKEN_KEY: 'ai-ui-gen-access-token',
  REFRESH_TOKEN_KEY: 'ai-ui-gen-refresh-token',

  // Store tokens in localStorage (client-side only)
  storeTokens: (accessToken: string, refreshToken: string) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(authStateUtils.ACCESS_TOKEN_KEY, accessToken)
      localStorage.setItem(authStateUtils.REFRESH_TOKEN_KEY, refreshToken)
    }
  },

  // Get tokens from localStorage (client-side only)
  getStoredTokens: () => {
    if (typeof window !== 'undefined') {
      return {
        accessToken: localStorage.getItem(authStateUtils.ACCESS_TOKEN_KEY),
        refreshToken: localStorage.getItem(authStateUtils.REFRESH_TOKEN_KEY),
      }
    }
    return { accessToken: null, refreshToken: null }
  },

  // Clear stored tokens
  clearStoredTokens: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem(authStateUtils.ACCESS_TOKEN_KEY)
      localStorage.removeItem(authStateUtils.REFRESH_TOKEN_KEY)
    }
  }
}
