// Server-Sent Events client for streaming AI responses

export interface SSEOptions {
  onMessage?: (data: any) => void
  onError?: (error: Event) => void
  onOpen?: () => void
  onClose?: () => void
  onStreamEnd?: () => void
}

export interface GenerationRequest {
  model?: string
  prompt: string
  maxTokens?: number
  temperature?: number
  projectId?: string
}

export class SSEClient {
  private eventSource: EventSource | null = null
  private url: string
  protected options: SSEOptions

  constructor(url: string, options: SSEOptions = {}) {
    this.url = url
    this.options = options
  }

  connect(): void {
    if (this.eventSource) {
      this.disconnect()
    }

    try {
      this.eventSource = new EventSource(this.url)
      
      this.eventSource.onopen = () => {
        console.log('SSE connection opened')
        this.options.onOpen?.()
      }
      
      this.eventSource.onmessage = (event) => {
        try {
          // Handle special events
          if (event.data === '[DONE]') {
            this.options.onStreamEnd?.()
            this.disconnect()
            return
          }

          // Try to parse JSON data
          const data = JSON.parse(event.data)
          this.options.onMessage?.(data)
        } catch (error) {
          console.error('Failed to parse SSE message:', error)
          // Pass raw data if JSON parsing fails
          this.options.onMessage?.(event.data)
        }
      }
      
      this.eventSource.onerror = (error) => {
        console.error('SSE connection error:', error)
        this.options.onError?.(error)
      }
      
    } catch (error) {
      console.error('Failed to create SSE connection:', error)
      this.options.onError?.(error as Event)
    }
  }

  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
      console.log('SSE connection closed')
      this.options.onClose?.()
    }
  }

  isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN
  }
}

// Utility functions for API calls
export async function startGeneration(request: GenerationRequest, accessToken?: string): Promise<{ sessionId: string }> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
  
  try {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`
    }
    
    const response = await fetch(`${apiUrl}/api/v1/generate`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        ...request,
        stream: false, // Start with non-streaming for initial setup
      }),
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data = await response.json()
    
    // For now, return a mock session ID since backend streaming isn't fully connected
    return { sessionId: data.sessionId || 'mock-session-' + Date.now() }
    
  } catch (error) {
    console.error('Failed to start generation:', error)
    // Return mock data for development
    return { sessionId: 'mock-session-' + Date.now() }
  }
}

export function createStreamingClient(sessionId: string, options: SSEOptions): SSEClient {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
  const streamUrl = `${apiUrl}/api/v1/generate/stream?sessionId=${sessionId}`
  
  return new SSEClient(streamUrl, options)
}

// Mock SSE client for development/testing
export class MockSSEClient extends SSEClient {
  private mockInterval: NodeJS.Timeout | null = null
  private mockData = [
    'Creating a beautiful button component...',
    'Adding responsive styles...',
    'Implementing hover effects...',
    'Finalizing accessibility features...',
    'Component generation complete!'
  ]

  connect(): void {
    console.log('Mock SSE connection started')
    this.options.onOpen?.()
    
    let index = 0
    this.mockInterval = setInterval(() => {
      if (index < this.mockData.length) {
        this.options.onMessage?.({
          type: 'delta',
          content: this.mockData[index],
          timestamp: new Date().toISOString()
        })
        index++
      } else {
        this.options.onStreamEnd?.()
        this.disconnect()
      }
    }, 1000)
  }

  disconnect(): void {
    if (this.mockInterval) {
      clearInterval(this.mockInterval)
      this.mockInterval = null
    }
    console.log('Mock SSE connection closed')
    this.options.onClose?.()
  }

  isConnected(): boolean {
    return this.mockInterval !== null
  }
}

// Development helper to create mock streaming client
export function createMockStreamingClient(options: SSEOptions): MockSSEClient {
  return new MockSSEClient('mock://stream', options)
}
