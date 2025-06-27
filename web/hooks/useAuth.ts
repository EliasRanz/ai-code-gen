'use client'

import { useSession } from 'next-auth/react'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'

export interface UseAuthReturn {
  user: any
  isLoading: boolean
  isAuthenticated: boolean
  signOut: () => void
}

export function useAuth(requireAuth: boolean = false): UseAuthReturn {
  const { data: session, status } = useSession()
  const router = useRouter()

  useEffect(() => {
    if (requireAuth && status !== 'loading' && !session) {
      router.push('/login')
    }
  }, [session, status, requireAuth, router])

  const handleSignOut = () => {
    // Clear stored tokens
    if (typeof window !== 'undefined') {
      localStorage.removeItem('ai-ui-gen-access-token')
      localStorage.removeItem('ai-ui-gen-refresh-token')
    }
    
    // Sign out with NextAuth
    import('next-auth/react').then(({ signOut }) => {
      signOut({ callbackUrl: '/login' })
    })
  }

  return {
    user: session?.user || null,
    isLoading: status === 'loading',
    isAuthenticated: !!session,
    signOut: handleSignOut,
  }
}

// Hook for protected pages
export function useRequireAuth() {
  return useAuth(true)
}
