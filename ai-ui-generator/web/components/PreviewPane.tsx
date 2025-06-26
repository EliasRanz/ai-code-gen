'use client'

import { useState } from 'react'

interface PreviewPaneProps {
  generatedCode?: string
  isLoading?: boolean
  language?: string
}

export default function PreviewPane({ 
  generatedCode = '', 
  isLoading = false, 
  language = 'tsx' 
}: PreviewPaneProps) {
  const [activeTab, setActiveTab] = useState<'code' | 'preview'>('code')
  const [isCopied, setIsCopied] = useState(false)

  const handleCopy = async () => {
    if (generatedCode) {
      await navigator.clipboard.writeText(generatedCode)
      setIsCopied(true)
      setTimeout(() => setIsCopied(false), 2000)
    }
  }

  const handleExport = () => {
    if (generatedCode) {
      const blob = new Blob([generatedCode], { type: 'text/plain' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `component.${language}`
      a.click()
      URL.revokeObjectURL(url)
    }
  }

  return (
    <div className="flex flex-col h-full bg-white">
      {/* Preview Header */}
      <div className="flex items-center justify-between p-4 border-b bg-gray-50">
        <div className="flex items-center space-x-4">
          <h2 className="text-lg font-semibold text-gray-900">Preview</h2>
          <div className="flex bg-gray-200 rounded-md p-1">
            <button
              onClick={() => setActiveTab('code')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                activeTab === 'code'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Code
            </button>
            <button
              onClick={() => setActiveTab('preview')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                activeTab === 'preview'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Preview
            </button>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          {generatedCode && (
            <>
              <button
                onClick={handleCopy}
                className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {isCopied ? 'Copied!' : 'Copy Code'}
              </button>
              <button
                onClick={handleExport}
                className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                Export
              </button>
            </>
          )}
        </div>
      </div>
      
      {/* Preview Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'code' ? (
          /* Code View */
          <div className="h-full flex flex-col">
            <div className="px-4 py-2 bg-gray-100 border-b">
              <span className="text-sm font-medium text-gray-700">
                Generated {language.toUpperCase()} Code
              </span>
            </div>
            <div className="flex-1 overflow-auto p-4">
              {isLoading ? (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-500">Generating code...</p>
                  </div>
                </div>
              ) : generatedCode ? (
                <pre className="text-sm font-mono bg-gray-50 p-4 rounded border overflow-auto">
                  <code className="language-tsx">{generatedCode}</code>
                </pre>
              ) : (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center text-gray-500">
                    <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                      <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                      </svg>
                    </div>
                    <h3 className="text-lg font-medium text-gray-900 mb-2">No code generated yet</h3>
                    <p className="text-gray-500">
                      Start a conversation to generate your first component
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          /* Live Preview */
          <div className="h-full flex flex-col">
            <div className="px-4 py-2 bg-gray-100 border-b">
              <span className="text-sm font-medium text-gray-700">Live Preview</span>
            </div>
            <div className="flex-1 overflow-auto">
              {isLoading ? (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-500">Rendering preview...</p>
                  </div>
                </div>
              ) : generatedCode ? (
                <div className="p-4">
                  <div className="bg-white border rounded-lg p-6 shadow-sm">
                    <div className="text-center text-gray-500">
                      <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                        <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                        </svg>
                      </div>
                      <p className="text-lg font-medium text-gray-900 mb-2">Preview Placeholder</p>
                      <p className="text-gray-500">
                        Live component preview will be rendered here
                      </p>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center text-gray-500">
                    <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                      <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </div>
                    <h3 className="text-lg font-medium text-gray-900 mb-2">No preview available</h3>
                    <p className="text-gray-500">
                      Generate a component to see the live preview
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
