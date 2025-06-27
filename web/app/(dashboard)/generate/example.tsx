// Example usage of SSE client with the ChatInterface component
'use client'

import { useState } from 'react'
import ChatInterface from '@/components/ChatInterface'
import PreviewPane from '@/components/PreviewPane'
import { createMockStreamingClient, SSEOptions } from '@/lib/sse'

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

export default function ExampleGeneratePage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [generatedCode, setGeneratedCode] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)

  const handlePromptSubmit = (prompt: string) => {
    // Add user message
    const userMessage: ChatMessage = {
      id: Date.now() + '-user',
      role: 'user',
      content: prompt,
      timestamp: new Date()
    }
    setMessages(prev => [...prev, userMessage])
    setIsGenerating(true)

    // Set up SSE client options
    const sseOptions: SSEOptions = {
      onMessage: (data) => {
        console.log('SSE Message received:', data)
        if (data.content) {
          // Add assistant message parts as they come in
          setMessages(prev => {
            const lastMessage = prev[prev.length - 1]
            if (lastMessage && lastMessage.role === 'assistant') {
              // Update existing assistant message
              return prev.map((msg, index) => 
                index === prev.length - 1 
                  ? { ...msg, content: msg.content + data.content }
                  : msg
              )
            } else {
              // Create new assistant message
              return [...prev, {
                id: Date.now() + '-assistant',
                role: 'assistant' as const,
                content: data.content,
                timestamp: new Date()
              }]
            }
          })
        }
      },
      onOpen: () => {
        console.log('SSE connection opened')
      },
      onClose: () => {
        console.log('SSE connection closed')
        setIsGenerating(false)
      },
      onError: (error) => {
        console.error('SSE error:', error)
        setIsGenerating(false)
      },
      onStreamEnd: () => {
        console.log('SSE stream ended')
        setIsGenerating(false)
      }
    }

    // Create and connect mock streaming client
    const client = createMockStreamingClient(sseOptions)
    client.connect()
  }

  return (
    <div className="h-full flex">
      {/* Chat Interface */}
      <div className="w-1/2 border-r">
        <ChatInterface
          onPromptSubmit={handlePromptSubmit}
          isGenerating={isGenerating}
          messages={messages}
        />
      </div>

      {/* Preview Pane */}
      <div className="w-1/2">
        <PreviewPane
          generatedCode={generatedCode}
          isLoading={isGenerating}
          language="tsx"
        />
      </div>
    </div>
  )
}
