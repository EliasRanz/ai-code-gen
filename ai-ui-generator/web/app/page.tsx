export default function HomePage() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8 text-center">
        <div>
          <h1 className="text-4xl font-bold text-gray-900 mb-4">AI UI Generator</h1>
          <p className="text-lg text-gray-600 mb-8">
            Transform natural language prompts into high-quality, interactive frontend components
          </p>
        </div>
        
        <div className="space-y-4">
          <a
            href="/auth/login"
            className="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            Get Started
          </a>
          
          <a
            href="/dashboard/home"
            className="w-full flex justify-center py-3 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            View Dashboard
          </a>
        </div>
        
        <div className="text-sm text-gray-500">
          <p>Start creating beautiful UI components with AI</p>
        </div>
      </div>
    </div>
  )
}
