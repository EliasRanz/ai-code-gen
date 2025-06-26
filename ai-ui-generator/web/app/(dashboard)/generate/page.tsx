'use client'

import ChatInterface from '@/components/ChatInterface'
import PreviewPane from '@/components/PreviewPane'

export default function GeneratePage() {
  return (
    <div className="h-full flex">
      {/* Chat Interface */}
      <div className="w-1/2 border-r">
        <ChatInterface
          onPromptSubmit={(prompt: string) => console.log('Submit prompt:', prompt)}
          isGenerating={false}
          messages={[]}
        />
      </div>

      {/* Preview Pane */}
      <div className="w-1/2">
        <PreviewPane
          generatedCode=""
          isLoading={false}
          language="tsx"
        />
      </div>
    </div>
  )
}
