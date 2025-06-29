'use client'

import { useSearchParams } from 'next/navigation'
import { Suspense } from 'react'

function ErrorContent() {
  const searchParams = useSearchParams()
  const error = searchParams.get('error')

  const getErrorMessage = (error: string | null) => {
    switch (error) {
      case 'CredentialsSignin':
        return 'Invalid email or password. Please try again.'
      case 'AccessDenied':
        return 'Access denied. You do not have permission to sign in.'
      case 'RefreshAccessTokenError':
        return 'Session expired. Please sign in again.'
      default:
        return 'An authentication error occurred. Please try again.'
    }
  }

  return (
    <div className="max-w-md mx-auto mt-8">
      <div className="bg-red-50 border border-red-200 rounded-md p-6">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">
              Authentication Error
            </h3>
            <div className="mt-2 text-sm text-red-700">
              <p>{getErrorMessage(error)}</p>
            </div>
            <div className="mt-4">
              <a
                href="/login"
                className="text-sm font-medium text-red-800 underline hover:text-red-600"
              >
                Try signing in again
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function AuthErrorPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <ErrorContent />
    </Suspense>
  )
}
