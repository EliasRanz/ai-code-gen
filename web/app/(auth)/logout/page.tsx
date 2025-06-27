'use client'

import { useEffect } from 'react'
import { signOut } from 'next-auth/react'

export default function LogoutPage() {
  useEffect(() => {
    // Automatically sign out when this page loads
    signOut({ 
      callbackUrl: '/login?message=logged-out',
      redirect: true 
    })
  }, [])

  return (
    <div className="text-center">
      <div className="max-w-md mx-auto mt-8">
        <div className="bg-blue-50 border border-blue-200 rounded-md p-6">
          <h3 className="text-lg font-medium text-blue-800 mb-2">
            Signing you out...
          </h3>
          <p className="text-blue-700">
            Please wait while we sign you out of your account.
          </p>
        </div>
      </div>
    </div>
  )
}
